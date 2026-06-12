package gh

import (
	"context"
	"fmt"
)

const orgsPageSize = 100

const orgsQuery = `
query($first:Int!, $after:String) {
  viewer {
    organizations(first:$first, after:$after) {
      pageInfo { hasNextPage endCursor }
      nodes { login name }
    }
  }
}`

type orgsResponse struct {
	Viewer struct {
		Organizations struct {
			PageInfo struct {
				HasNextPage bool
				EndCursor   string
			}
			Nodes []struct {
				Login string
				Name  string
			}
		}
	}
}

// FetchViewerOrgs は認証ユーザーが参加している org を最大 limit 件取得する。
// limit < 0 で全件。limit=0 の場合は0件を返す（取得しない）。
func FetchViewerOrgs(ctx context.Context, client GraphQLDoer, limit int) ([]Org, error) {
	var orgs []Org
	var cursor string

	for {
		pageSize := orgsPageSize
		if limit >= 0 {
			remaining := limit - len(orgs)
			if remaining <= 0 {
				break
			}
			if remaining < pageSize {
				pageSize = remaining
			}
		}

		variables := map[string]interface{}{
			"first": pageSize,
		}
		if cursor != "" {
			variables["after"] = cursor
		}

		var resp orgsResponse
		if err := client.DoWithContext(ctx, orgsQuery, variables, &resp); err != nil {
			return nil, fmt.Errorf("fetch viewer orgs: %w", err)
		}

		orgsData := resp.Viewer.Organizations
		for _, node := range orgsData.Nodes {
			orgs = append(orgs, Org{
				Login: node.Login,
				Name:  node.Name,
			})

			if limit >= 0 && len(orgs) >= limit {
				break
			}
		}

		if !orgsData.PageInfo.HasNextPage {
			break
		}
		if limit >= 0 && len(orgs) >= limit {
			break
		}

		cursor = orgsData.PageInfo.EndCursor
	}

	return orgs, nil
}
