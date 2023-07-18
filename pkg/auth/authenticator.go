package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/0xa1-red/empires-of-avalon/config"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
	"golang.org/x/oauth2"
)

var authenticator *Authenticator

// Authenticator is used to authenticate our users.
type Authenticator struct {
	*oidc.Provider
	oauth2.Config
}

func Init() error {
	if authenticator == nil {
		domain := viper.GetString(config.Authenticator_Domain)
		clientID := viper.GetString(config.Authenticator_Client_ID)
		clientSecret := viper.GetString(config.Authenticator_Client_Secret)
		callback := viper.GetString(config.Authenticator_Callback)

		slog.Info("initializing authenticator",
			"domain", domain,
			"client_id", clientID,
			"callback", callback,
		)

		provider, err := oidc.NewProvider(
			context.Background(),
			fmt.Sprintf("https://%s/", domain),
		)
		if err != nil {
			return err
		}

		conf := oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  callback,
			Endpoint:     provider.Endpoint(),
			Scopes:       []string{oidc.ScopeOpenID, "profile"},
		}

		authenticator = &Authenticator{
			Provider: provider,
			Config:   conf,
		}
	}

	return nil
}

func Get() *Authenticator {
	return authenticator
}

// VerifyIDToken verifies that an *oauth2.Token is a valid *oidc.IDToken.
func (a *Authenticator) VerifyIDToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("no id_token field in oauth2 token")
	}

	oidcConfig := &oidc.Config{ // nolint:exhaustruct
		ClientID: a.ClientID,
	}

	return a.Verifier(oidcConfig).Verify(ctx, rawIDToken)
}

func GenerateNonce() (string, error) {
	b := make([]byte, 32)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	state := base64.StdEncoding.EncodeToString(b)

	return state, nil
}
