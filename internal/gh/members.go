package gh

import (
	"context"
	"fmt"
	"strings"
)

const membersPageSize = 100

const membersQuery = `
query($org:String!, $first:Int!, $after:String) {
  organization(login:$org) {
    membersWithRole(first:$first, after:$after) {
      totalCount
      pageInfo { hasNextPage endCursor }
      edges { role node { name login databaseId url } }
    }
  }
}`

type membersResponse struct {
	Organization struct {
		MembersWithRole struct {
			TotalCount int
			PageInfo   struct {
				HasNextPage bool
				EndCursor   string
			}
			Edges []struct {
				Role string
				Node struct {
					Name       string
					Login      string
					DatabaseID int `json:"databaseId"`
					URL        string `json:"url"`
				}
			}
		}
	}
}

// FetchMembers fetches members of org up to limit items (limit < 0 means all).
// role accepts "all", "admin", or "member" and is filtered client-side.
func FetchMembers(ctx context.Context, client GraphQLDoer, org string, limit int, role string) ([]Member, error) {
	var members []Member
	var cursor string

	for {
		pageSize := membersPageSize
		if limit >= 0 {
			remaining := limit - len(members)
			if remaining <= 0 {
				break
			}
			// When a role filter is active, unfiltered API results count towards
			// the page but not towards the limit, so always fetch a full page to
			// avoid O(n/adminRatio) round-trips on sparse populations.
			if strings.ToLower(role) == "all" && remaining < pageSize {
				pageSize = remaining
			}
		}

		variables := map[string]interface{}{
			"org":   org,
			"first": pageSize,
		}
		if cursor != "" {
			variables["after"] = cursor
		}

		var resp membersResponse
		if err := client.DoWithContext(ctx, membersQuery, variables, &resp); err != nil {
			return nil, fmt.Errorf("fetch members for org %q: %w", org, err)
		}

		mwr := resp.Organization.MembersWithRole
		for _, edge := range mwr.Edges {
				if !matchRole(edge.Role, role) {
				continue
			}

			members = append(members, Member{
				Name:       edge.Node.Name,
				Login:      edge.Node.Login,
				Role:       edge.Role,
				DatabaseID: edge.Node.DatabaseID,
				URL:        edge.Node.URL,
			})

			if limit >= 0 && len(members) >= limit {
				break
			}
		}

		if !mwr.PageInfo.HasNextPage {
			break
		}
		if limit >= 0 && len(members) >= limit {
			break
		}

		cursor = mwr.PageInfo.EndCursor
	}

	return members, nil
}

// matchRole reports whether edgeRole satisfies the given filter.
func matchRole(edgeRole, filter string) bool {
	switch strings.ToLower(filter) {
	case "admin":
		return edgeRole == "ADMIN"
	case "member":
		return edgeRole == "MEMBER"
	default: // "all" or unknown
		return true
	}
}
