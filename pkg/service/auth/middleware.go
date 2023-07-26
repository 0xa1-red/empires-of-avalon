package auth

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/0xa1-red/empires-of-avalon/config"
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
)

// CustomClaims contains custom data we want from the token.
type CustomClaims struct {
	Scope   string `json:"scope"`
	Subject string `json:"sub"`
}

// Validate does nothing for this example, but we need
// it to satisfy validator.CustomClaims interface.
func (c CustomClaims) Validate(ctx context.Context) error {
	return nil
}

// EnsureValidToken is a middleware that will check the validity of our JWT.
func EnsureValidToken() func(next http.Handler) http.Handler {
	issuerURL, err := url.Parse("https://" + viper.GetString(config.Authenticator_Domain) + "/")
	if err != nil {
		log.Fatalf("Failed to parse the issuer url: %v", err)
	}

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{viper.GetString(config.Authenticator_Audience)},
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &CustomClaims{} // nolint:exhaustruct
			},
		),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		log.Fatalf("Failed to set up the jwt validator")
	}

	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Encountered error while validating JWT: %v", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)

		if _, err := w.Write([]byte(`{"message":"Failed to validate JWT."}`)); err != nil {
			slog.Error("failed to write response", err)
		}
	}

	middleware := jwtmiddleware.New(
		jwtValidator.ValidateToken,
		jwtmiddleware.WithErrorHandler(errorHandler),
	)

	return func(next http.Handler) http.Handler {
		return middleware.CheckJWT(next)
	}
}
