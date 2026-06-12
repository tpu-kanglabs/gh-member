package gh_test

import (
	"context"
	"testing"

	"github.com/tpu-kanglabs/gh-member/internal/gh"
)

// TestFetchViewerOrgs_Basic tests successful retrieval of organizations.
func TestFetchViewerOrgs_Basic(t *testing.T) {
	transport := &mockTransport{
		responses: []mockResponse{
			{
				statusCode: 200,
				body: `{
					"data": {
						"viewer": {
							"organizations": {
								"pageInfo": {"hasNextPage": false, "endCursor": "cursor1"},
								"nodes": [
									{"login": "org1", "name": "Organization One"},
									{"login": "org2", "name": "Organization Two"}
								]
							}
						}
					}
				}`,
			},
		},
	}

	client := newTestGraphQLClient(t, transport)
	orgs, err := gh.FetchViewerOrgs(context.Background(), client, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(orgs) != 2 {
		t.Fatalf("expected 2 orgs, got %d", len(orgs))
	}

	if orgs[0].Login != "org1" {
		t.Errorf("expected login org1, got %s", orgs[0].Login)
	}
	if orgs[0].Name != "Organization One" {
		t.Errorf("expected name 'Organization One', got %s", orgs[0].Name)
	}
	if orgs[1].Login != "org2" {
		t.Errorf("expected login org2, got %s", orgs[1].Login)
	}
}

// TestFetchViewerOrgs_Pagination tests retrieval across multiple pages.
func TestFetchViewerOrgs_Pagination(t *testing.T) {
	transport := &mockTransport{
		responses: []mockResponse{
			{
				statusCode: 200,
				body: `{
					"data": {
						"viewer": {
							"organizations": {
								"pageInfo": {"hasNextPage": true, "endCursor": "cursor1"},
								"nodes": [
									{"login": "org1", "name": "Organization One"},
									{"login": "org2", "name": "Organization Two"}
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
						"viewer": {
							"organizations": {
								"pageInfo": {"hasNextPage": false, "endCursor": "cursor2"},
								"nodes": [
									{"login": "org3", "name": "Organization Three"}
								]
							}
						}
					}
				}`,
			},
		},
	}

	client := newTestGraphQLClient(t, transport)
	orgs, err := gh.FetchViewerOrgs(context.Background(), client, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(orgs) != 3 {
		t.Fatalf("expected 3 orgs, got %d", len(orgs))
	}

	logins := []string{orgs[0].Login, orgs[1].Login, orgs[2].Login}
	expected := []string{"org1", "org2", "org3"}
	for i, login := range logins {
		if login != expected[i] {
			t.Errorf("expected login[%d]=%s, got %s", i, expected[i], login)
		}
	}

	if transport.callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", transport.callCount)
	}
}

// TestFetchViewerOrgs_Limit tests that fetching stops once the limit is reached.
func TestFetchViewerOrgs_Limit(t *testing.T) {
	transport := &mockTransport{
		responses: []mockResponse{
			{
				statusCode: 200,
				body: `{
					"data": {
						"viewer": {
							"organizations": {
								"pageInfo": {"hasNextPage": true, "endCursor": "cursor1"},
								"nodes": [
									{"login": "org1", "name": "Organization One"},
									{"login": "org2", "name": "Organization Two"},
									{"login": "org3", "name": "Organization Three"}
								]
							}
						}
					}
				}`,
			},
		},
	}

	client := newTestGraphQLClient(t, transport)
	orgs, err := gh.FetchViewerOrgs(context.Background(), client, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(orgs) != 2 {
		t.Fatalf("expected 2 orgs (limit), got %d", len(orgs))
	}

	if orgs[0].Login != "org1" {
		t.Errorf("expected org1, got %s", orgs[0].Login)
	}
	if orgs[1].Login != "org2" {
		t.Errorf("expected org2, got %s", orgs[1].Login)
	}

	// 2 orgs fetched; no further page request expected
	if transport.callCount != 1 {
		t.Errorf("expected 1 API call, got %d", transport.callCount)
	}
}
