package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/api"
)

// GraphQLDoer executes a GraphQL query. Defined as an interface to allow test injection.
type GraphQLDoer interface {
	DoWithContext(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error
}

// NewDefaultClient returns the default GraphQL client backed by the active gh login session.
func NewDefaultClient() (GraphQLDoer, error) {
	return api.DefaultGraphQLClient()
}
