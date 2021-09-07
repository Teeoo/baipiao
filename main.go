package main

import (
	"context"
	"errors"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/teeoo/baipiao/graph"
	"github.com/teeoo/baipiao/graph/generated"
	_ "github.com/teeoo/baipiao/jd"
	"log"
	"net/http"
)

func main() {
	c := generated.Config{Resolvers: &graph.Resolver{}}
	c.Directives.Authorization = func(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
		// TODO:JWT
		auth := ctx.Value("Authorization")
		if auth == "" {
			return nil, errors.New("A token is required")
		}
		return next(ctx)
	}
	http.Handle("/graphql", Middleware(handler.NewDefaultServer(generated.NewExecutableSchema(c))))
	log.Fatal(http.ListenAndServe(":1234", nil))
}

// Middleware decodes the share Authorization and packs into context
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := r.Header.Get("Authorization")
		ctx := context.WithValue(r.Context(), "Authorization", c)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
