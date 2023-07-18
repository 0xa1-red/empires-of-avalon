package auth

import (
	"context"
	"encoding/gob"
	"net/http"

	session "github.com/go-session/session/v3"
)

type AuthContextKey int

const (
	ContextProfile AuthContextKey = iota
	ContextUserProfile
)

// IsAuthenticated is a middleware that checks if
// the user has already been authenticated previously.
func IsAuthenticated(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		store, err := session.Start(context.Background(), w, r)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		profile, profOk := store.Get("profile")
		userProfile, userProfOk := store.Get("userprofile")

		if !profOk || !userProfOk {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		gob.Register(map[string]interface{}{})

		ctx := context.WithValue(r.Context(), ContextProfile, profile.(map[string]interface{}))
		ctx = context.WithValue(ctx, ContextUserProfile, userProfile.(map[string]interface{}))

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
