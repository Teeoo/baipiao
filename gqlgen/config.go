package gqlgen

import (
	"context"
	"errors"
	"github.com/99designs/gqlgen/graphql"
	. "github.com/teeoo/baipiao/graph"
	. "github.com/teeoo/baipiao/graph/generated"
)

func New() Config {
	c := Config{Resolvers: &Resolver{}}
	c.Directives.Authorization = func(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
		// TODO:JWT
		auth := ctx.Value("Authorization")
		if auth == "" {
			return nil, errors.New("A token is required")
		}
		return next(ctx)
	}
	return c
}
