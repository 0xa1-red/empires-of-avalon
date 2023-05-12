package httpserver

import (
	"context"
	"net/http"

	"golang.org/x/exp/slog"
)

type ContextKey struct {
	name string
}

var (
	ContextAuth = ContextKey{name: "auth"}
)

func Authentication(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		auth := r.Header.Get("Authorization")
		if auth == "" {
			E(w, http.StatusUnauthorized)
		}
		slog.Info("authenticated", "user_id", auth)
		ctx = context.WithValue(ctx, ContextAuth, auth)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
