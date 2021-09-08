package gqlgen

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/golang-jwt/jwt/v4"
	"github.com/teeoo/baipiao/cache"
	. "github.com/teeoo/baipiao/graph"
	. "github.com/teeoo/baipiao/graph/generated"
	. "github.com/teeoo/baipiao/middleware"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"log"
	"net/http"
	"strings"
	"time"
)

func init() {
	c, err := cache.NewCache("127.0.0.1:6379", 24*time.Hour)
	if err != nil {
		log.Fatalf("cannot create APQ redis cache: %v", err)
	}
	cfg := Config{Resolvers: &Resolver{}}
	cfg.Directives.Authorization = func(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
		authHeader := ctx.Value("Authorization")
		if authHeader == "" {
			graphql.AddError(ctx, &gqlerror.Error{
				Path:    graphql.GetPath(ctx),
				Message: "A token is required",
				Extensions: map[string]interface{}{
					"statusCode": 401,
				},
			})
			return nil, nil
		}
		authHeaderParts := strings.Split(authHeader.(string), " ")
		if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
			graphql.AddError(ctx, &gqlerror.Error{
				Path:    graphql.GetPath(ctx),
				Message: "Authorization header format must be Bearer {token}",
				Extensions: map[string]interface{}{
					"statusCode": 401,
				},
			})
			return nil, nil
		}
		_, err := parseToken(authHeaderParts[1], "e3ffb9840f367eb8f5d8c579b813ea35d0689c04")
		if err != nil {
			graphql.AddError(ctx, &gqlerror.Error{
				Path:    graphql.GetPath(ctx),
				Message: err.Error(),
				Extensions: map[string]interface{}{
					"statusCode": 401,
				},
			})
			return nil, nil
		}
		return next(ctx)
	}
	handle := handler.NewDefaultServer(NewExecutableSchema(cfg))
	handle.AddTransport(transport.POST{})
	handle.Use(extension.AutomaticPersistedQuery{Cache: c})
	http.Handle("/graphql", Jwt(handle))
}

func parseToken(token string, secret string) (jwt.Claims, error) {
	claim, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	return claim.Claims, nil
}
