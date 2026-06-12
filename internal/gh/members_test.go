package gh_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/rmuraix/gh-member/internal/gh"
)

// mockTransport is an http.RoundTripper that returns pre-registered responses.
type mockTransport struct {
	responses []mockResponse
	callCount int
}

type mockResponse struct {
	body       string
	statusCode int
}

func (m *mockTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	if m.callCount >= len(m.responses) {
		panic(fmt.Sprintf("unexpected API call #%d (only %d responses registered)", m.callCount+1, len(m.responses)))
	}
	r := m.responses[m.callCount]
	m.callCount++
	return &http.Response{
		StatusCode: r.statusCode,
		Body:       io.NopCloser(strings.NewReader(r.body)),
		Header:     make(http.Header),
	}, nil
}

func newTestGraphQLClient(t *testing.T, transport http.RoundTripper) gh.GraphQLDoer {
	t.Helper()
	client, err := api.NewGraphQLClient(api.ClientOptions{
		Host:      "github.com",
		AuthToken: "fake-token",
		Transport: transport,
	})
	if err != nil {
		t.Fatalf("failed to create GraphQL client: %v", err)
	}
	return client
}

// TestFetchMembers_SinglePage tests successful retrieval when all results fit in a single page.
func TestFetchMembers_SinglePage(t *testing.T) {
	transport := &mockTransport{
		responses: []mockResponse{
			{
				statusCode: 200,
				body: `{
					"data": {
						"organization": {
							"membersWithRole": {
								"totalCount": 2,
								"pageInfo": {"hasNextPage": false, "endCursor": "cursor1"},
								"edges": [
									{"role": "ADMIN", "node": {"name": "Alice", "login": "alice", "databaseId": 1, "url": "https://github.com/alice"}},
									{"role": "MEMBER", "node": {"name": "Bob", "login": "bob", "databaseId": 2, "url": "https://github.com/bob"}}
								]
							}
						}
					}
				}`,
			},
		},
	}

	client := newTestGraphQLClient(t, transport)
	members, err := gh.FetchMembers(context.Background(), client, "myorg", -1, "all")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(members))
	}

	if members[0].Login != "alice" {
		t.Errorf("expected login alice, got %s", members[0].Login)
	}
	if members[0].Name != "Alice" {
		t.Errorf("expected name Alice, got %s", members[0].Name)
	}
	if members[0].Role != "ADMIN" {
		t.Errorf("expected role ADMIN, got %s", members[0].Role)
	}
	if members[0].DatabaseID != 1 {
		t.Errorf("expected databaseId 1, got %d", members[0].DatabaseID)
	}
	if members[0].URL != "https://github.com/alice" {
		t.Errorf("expected url https://github.com/alice, got %s", members[0].URL)
	}

	if members[1].Login != "bob" {
		t.Errorf("expected login bob, got %s", members[1].Login)
	}
	if members[1].Role != "MEMBER" {
		t.Errorf("expected role MEMBER, got %s", members[1].Role)
	}
}

// TestFetchMembers_Pagination tests successful retrieval across multiple pages.
func TestFetchMembers_Pagination(t *testing.T) {
	transport := &mockTransport{
		responses: []mockResponse{
			{
				statusCode: 200,
				body: `{
					"data": {
						"organization": {
							"membersWithRole": {
								"totalCount": 3,
								"pageInfo": {"hasNextPage": true, "endCursor": "cursor1"},
								"edges": [
									{"role": "ADMIN", "node": {"name": "Alice", "login": "alice", "databaseId": 1, "url": "https://github.com/alice"}},
									{"role": "MEMBER", "node": {"name": "Bob", "login": "bob", "databaseId": 2, "url": "https://github.com/bob"}}
								]
							}
						}
					}
				}`,
			},
			{
				statusCode: 200,
				body: `{
					"data": {
						"organization": {
							"membersWithRole": {
								"totalCount": 3,
								"pageInfo": {"hasNextPage": false, "endCursor": "cursor2"},
								"edges": [
									{"role": "MEMBER", "node": {"name": "Carol", "login": "carol", "databaseId": 3, "url": "https://github.com/carol"}}
								]
							}
						}
					}
				}`,
			},
		},
	}

	client := newTestGraphQLClient(t, transport)
	members, err := gh.FetchMembers(context.Background(), client, "myorg", -1, "all")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(members) != 3 {
		t.Fatalf("expected 3 members, got %d", len(members))
	}

	logins := []string{members[0].Login, members[1].Login, members[2].Login}
	expected := []string{"alice", "bob", "carol"}
	for i, login := range logins {
		if login != expected[i] {
			t.Errorf("expected login[%d]=%s, got %s", i, expected[i], login)
		}
	}

	if transport.callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", transport.callCount)
	}
}

// TestFetchMembers_Limit tests that fetching stops once the limit is reached.
func TestFetchMembers_Limit(t *testing.T) {
	transport := &mockTransport{
		responses: []mockResponse{
			{
				statusCode: 200,
				body: `{
					"data": {
						"organization": {
							"membersWithRole": {
								"totalCount": 5,
								"pageInfo": {"hasNextPage": true, "endCursor": "cursor1"},
								"edges": [
									{"role": "ADMIN", "node": {"name": "Alice", "login": "alice", "databaseId": 1, "url": "https://github.com/alice"}},
									{"role": "MEMBER", "node": {"name": "Bob", "login": "bob", "databaseId": 2, "url": "https://github.com/bob"}},
									{"role": "MEMBER", "node": {"name": "Carol", "login": "carol", "databaseId": 3, "url": "https://github.com/carol"}}
								]
							}
						}
					}
				}`,
			},
		},
	}

	client := newTestGraphQLClient(t, transport)
	members, err := gh.FetchMembers(context.Background(), client, "myorg", 2, "all")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(members) != 2 {
		t.Fatalf("expected 2 members (limit), got %d", len(members))
	}

	if members[0].Login != "alice" {
		t.Errorf("expected alice, got %s", members[0].Login)
	}
	if members[1].Login != "bob" {
		t.Errorf("expected bob, got %s", members[1].Login)
	}

	// 2 members fetched; no further page request expected
	if transport.callCount != 1 {
		t.Errorf("expected 1 API call, got %d", transport.callCount)
	}
}

// TestFetchMembers_RoleFilter tests that only admin members are returned when role="admin".
func TestFetchMembers_RoleFilter(t *testing.T) {
	transport := &mockTransport{
		responses: []mockResponse{
			{
				statusCode: 200,
				body: `{
					"data": {
						"organization": {
							"membersWithRole": {
								"totalCount": 3,
								"pageInfo": {"hasNextPage": false, "endCursor": "cursor1"},
								"edges": [
									{"role": "ADMIN", "node": {"name": "Alice", "login": "alice", "databaseId": 1, "url": "https://github.com/alice"}},
									{"role": "MEMBER", "node": {"name": "Bob", "login": "bob", "databaseId": 2, "url": "https://github.com/bob"}},
									{"role": "ADMIN", "node": {"name": "Dave", "login": "dave", "databaseId": 4, "url": "https://github.com/dave"}}
								]
							}
						}
					}
				}`,
			},
		},
	}

	client := newTestGraphQLClient(t, transport)
	members, err := gh.FetchMembers(context.Background(), client, "myorg", -1, "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(members) != 2 {
		t.Fatalf("expected 2 admin members, got %d", len(members))
	}

	for _, m := range members {
		if m.Role != "ADMIN" {
			t.Errorf("expected role ADMIN, got %s for login %s", m.Role, m.Login)
		}
	}

	if members[0].Login != "alice" {
		t.Errorf("expected alice, got %s", members[0].Login)
	}
	if members[1].Login != "dave" {
		t.Errorf("expected dave, got %s", members[1].Login)
	}
}

// TestFetchMembers_EmptyName tests that an empty Name from the API is stored as an empty string.
func TestFetchMembers_EmptyName(t *testing.T) {
	transport := &mockTransport{
		responses: []mockResponse{
			{
				statusCode: 200,
				body: `{
					"data": {
						"organization": {
							"membersWithRole": {
								"totalCount": 1,
								"pageInfo": {"hasNextPage": false, "endCursor": "cursor1"},
								"edges": [
									{"role": "MEMBER", "node": {"name": "", "login": "noname", "databaseId": 5, "url": "https://github.com/noname"}}
								]
							}
						}
					}
				}`,
			},
		},
	}

	client := newTestGraphQLClient(t, transport)
	members, err := gh.FetchMembers(context.Background(), client, "myorg", -1, "all")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(members))
	}

	// Name is preserved as an empty string; the presentation layer handles the fallback.
	if members[0].Name != "" {
		t.Errorf("expected empty name, got %q", members[0].Name)
	}
	if members[0].Login != "noname" {
		t.Errorf("expected login noname, got %s", members[0].Login)
	}
}

// TestFetchMembers_ErrorResponse tests that a GraphQL error (e.g. org not found) is returned as an error.
func TestFetchMembers_ErrorResponse(t *testing.T) {
	transport := &mockTransport{
		responses: []mockResponse{
			{
				statusCode: 200,
				body: `{
					"data": null,
					"errors": [
						{"message": "Could not resolve to an Organization with the login of 'nonexistent'.", "type": "NOT_FOUND"}
					]
				}`,
			},
		},
	}

	client := newTestGraphQLClient(t, transport)
	_, err := gh.FetchMembers(context.Background(), client, "nonexistent", -1, "all")
	if err == nil {
		t.Fatal("expected an error but got nil")
	}
}

// TestFetchMembers_MemberRoleFilter tests that only member-role members are returned when role="member".
func TestFetchMembers_MemberRoleFilter(t *testing.T) {
	transport := &mockTransport{
		responses: []mockResponse{
			{
				statusCode: 200,
				body: `{
					"data": {
						"organization": {
							"membersWithRole": {
								"totalCount": 3,
								"pageInfo": {"hasNextPage": false, "endCursor": "cursor1"},
								"edges": [
									{"role": "ADMIN", "node": {"name": "Alice", "login": "alice", "databaseId": 1, "url": "https://github.com/alice"}},
									{"role": "MEMBER", "node": {"name": "Bob", "login": "bob", "databaseId": 2, "url": "https://github.com/bob"}},
									{"role": "MEMBER", "node": {"name": "Carol", "login": "carol", "databaseId": 3, "url": "https://github.com/carol"}}
								]
							}
						}
					}
				}`,
			},
		},
	}

	client := newTestGraphQLClient(t, transport)
	members, err := gh.FetchMembers(context.Background(), client, "myorg", -1, "member")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(members) != 2 {
		t.Fatalf("expected 2 MEMBER role members, got %d", len(members))
	}

	for _, m := range members {
		if m.Role != "MEMBER" {
			t.Errorf("expected role MEMBER, got %s for login %s", m.Role, m.Login)
		}
	}
}
