package router

import (
	"context"
	"fmt"
	"net/http"
	"time"

	intmw "github.com/0xa1-red/empires-of-avalon/pkg/middleware"
	"github.com/0xa1-red/empires-of-avalon/pkg/model"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/auth"
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"golang.org/x/exp/slog"
)

func New() *chi.Mux {
	s := chi.NewRouter()

	s.Use(intmw.AvalonLogger)
	s.Use(middleware.AllowContentType("application/json"))
	s.Use(middleware.Timeout(60 * time.Second))

	s.Mount("/api", GameRouter())
	s.Mount("/admin", AdminRouter())

	s.Get("/healthz", Healthcheck)

	return s
}

func E(w http.ResponseWriter, r *http.Request, status int, err error) {
	res := model.ErrorResponse{
		Status:     status,
		StatusText: http.StatusText(status),
		Error:      err,
	}

	slog.Error("HTTP error", err)

	render.Status(r, status)
	render.JSON(w, r, res)
}

func authFromContext(w http.ResponseWriter, r *http.Request, ctx context.Context) string {
	claims := ctx.Value(jwtmiddleware.ContextKey{})
	if claims == nil {
		E(w, r, http.StatusInternalServerError, fmt.Errorf("claims not found"))
		return ""
	}

	validatedClaims, ok := claims.(*validator.ValidatedClaims)
	if !ok {
		E(w, r, http.StatusInternalServerError, fmt.Errorf("failed to validate claims 1"))
		return ""
	}

	customClaims, ok := validatedClaims.CustomClaims.(*auth.CustomClaims)
	if !ok {
		E(w, r, http.StatusInternalServerError, fmt.Errorf("failed to validate claims 2"))
		return ""
	}

	id := customClaims.Subject

	profile, err := auth.GetUserProfile(id)
	if err != nil {
		E(w, r, http.StatusInternalServerError, err)
		return ""
	}

	if metadata, ok := profile["app_metadata"].(map[string]interface{}); ok {
		if eid, ok := metadata["external_id"].(string); ok {
			return eid
		}
	}

	return ""
}
