package middleware

import (
	"context"
	"net/http"
)

// Jwt Middleware decodes the share Authorization and packs into context
func Jwt(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := r.Header.Get("Authorization")
		ctx := context.WithValue(r.Context(), "Authorization", c)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

