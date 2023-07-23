package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/0xa1-red/empires-of-avalon/config"
	"github.com/spf13/viper"
)

var accessToken *Token

type Token struct {
	Kind      string
	Token     string
	Scopes    []string
	ExpiresAt time.Time
}

func GetToken() (*Token, error) {
	if accessToken == nil || accessToken.ExpiresAt.Before(time.Now()) {
		domain := viper.GetString(config.Authenticator_Domain)
		url := fmt.Sprintf("https://%s/oauth/token", domain)

		payload := map[string]string{
			"client_id":     viper.GetString(config.Authenticator_Client_ID),
			"client_secret": viper.GetString(config.Authenticator_Client_Secret),
			"audience":      fmt.Sprintf("https://%s/api/v2/", domain),
			"grant_type":    "client_credentials",
		}

		w := bytes.NewBuffer([]byte(""))

		encoder := json.NewEncoder(w)
		if err := encoder.Encode(payload); err != nil {
			return nil, err
		}

		req, err := http.NewRequest("POST", url, w)
		if err != nil {
			return nil, err
		}

		req.Header.Add("content-type", "application/json")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		defer res.Body.Close()

		decoder := json.NewDecoder(res.Body)
		data := make(map[string]interface{})

		if err := decoder.Decode(&data); err != nil {
			return nil, err
		}

		accessToken = &Token{
			Kind:      data["token_type"].(string),
			Token:     data["access_token"].(string),
			Scopes:    strings.Split(data["scope"].(string), " "),
			ExpiresAt: time.Now().Add(time.Duration(data["expires_in"].(float64)) * time.Second),
		}
	}

	return accessToken, nil
}

func GetUserProfile(id string) (map[string]interface{}, error) {
	managementURL := fmt.Sprintf("https://%s/api/v2/users/%s", viper.GetString(config.Authenticator_Domain), id)

	client := http.DefaultClient
	req, err := http.NewRequest(http.MethodGet, managementURL, nil)

	if err != nil {
		return nil, err
	}

	token, err := GetToken()
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("%s %s", token.Kind, token.Token))

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	profile := make(map[string]interface{})

	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&profile); err != nil {
		return nil, err
	}

	return profile, nil
}
