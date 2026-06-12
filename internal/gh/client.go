package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/api"
)

// GraphQLDoer は GraphQL クエリを実行するインターフェース（テスト注入用）。
type GraphQLDoer interface {
	DoWithContext(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error
}

// NewDefaultClient は api.DefaultGraphQLClient を返す。
func NewDefaultClient() (GraphQLDoer, error) {
	return api.DefaultGraphQLClient()
}
