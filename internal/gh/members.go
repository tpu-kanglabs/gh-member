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

// FetchMembers は org のメンバーを取得する。
// limit < 0 の場合は全件取得。limit >= 0 の場合は最大 limit 件。
// limit=0 の場合は0件を返す（取得しない）。
// role は "all", "admin", "member" を受け付け、クライアント側でフィルタリングする。
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
			if remaining < pageSize {
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
			// role フィルタリング（クライアント側）
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

// matchRole は edge の role が指定の role フィルタにマッチするか確認する。
func matchRole(edgeRole, filter string) bool {
	switch strings.ToLower(filter) {
	case "admin":
		return edgeRole == "ADMIN"
	case "member":
		return edgeRole == "MEMBER"
	default: // "all" またはその他
		return true
	}
}
