package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/api"
)

// GraphQLDoer は GraphQL クエリを実行するインターフェース（テスト注入用）。
type GraphQLDoer interface {
	DoWithContext(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error
}

// defaultClient は api.DefaultGraphQLClient を返す（本番用）。
func defaultClient() (GraphQLDoer, error) {
	return api.DefaultGraphQLClient()
}

// NewDefaultClient は api.DefaultGraphQLClient を返す（外部パッケージから利用可能）。
func NewDefaultClient() (GraphQLDoer, error) {
	return api.DefaultGraphQLClient()
}
