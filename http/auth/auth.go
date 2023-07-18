package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/0xa1-red/empires-of-avalon/config"
	"github.com/0xa1-red/empires-of-avalon/pkg/auth"
	"github.com/go-chi/chi"
	session "github.com/go-session/session/v3"
	"github.com/spf13/viper"
)

type Router struct {
	chi.Router

	auth *auth.Authenticator
}

func NewRouter() *Router {
	r := chi.NewRouter()

	router := &Router{
		Router: r,

		auth: auth.Get(),
	}

	r.Get("/login", router.Login)
	r.Get("/logout", router.Logout)
	r.Get("/callback", router.Callback)
	// r.Get("/profile", router.Profile)
	// r.With(auth.IsAuthenticated).Get("/test", router.Test)

	return router
}

func (rt *Router) Login(w http.ResponseWriter, r *http.Request) {
	store, err := session.Start(r.Context(), w, r)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	nonce, err := auth.GenerateNonce()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	store.Set("state", nonce)

	err = store.Save()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, rt.auth.AuthCodeURL(nonce), http.StatusTemporaryRedirect)
}

func (rt *Router) Callback(w http.ResponseWriter, r *http.Request) {
	store, err := session.Start(r.Context(), w, r)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if state, ok := store.Get("state"); ok && state != r.URL.Query().Get("state") {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	token, err := rt.auth.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	idToken, err := rt.auth.VerifyIDToken(r.Context(), token)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var profile map[string]interface{}
	if err := idToken.Claims(&profile); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	store.Set("access_token", token.AccessToken)
	store.Set("profile", profile)

	userProfile, err := getUserProfile(profile["sub"].(string))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	store.Set("userprofile", userProfile)

	if err := store.Save(); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/auth/profile", http.StatusTemporaryRedirect)
}

func (rt *Router) Logout(w http.ResponseWriter, r *http.Request) {
	logoutUrl, err := url.Parse("https://" + viper.GetString(config.Authenticator_Domain) + "/auth/logout")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	returnTo, err := url.Parse(viper.GetString(config.Authenticator_Callback))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	returnTo.Path = ""

	parameters := url.Values{}
	parameters.Add("returnTo", returnTo.String())
	parameters.Add("client_id", viper.GetString(config.Authenticator_Client_ID))
	logoutUrl.RawQuery = parameters.Encode()

	http.Redirect(w, r, logoutUrl.String(), http.StatusTemporaryRedirect)
}

// func (rt *Router) Profile(w http.ResponseWriter, r *http.Request) {
// 	store, err := session.Start(r.Context(), w, r)
// 	if err != nil {
// 		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
// 		return
// 	}

// 	profile, ok := store.Get("profile")
// 	if !ok {
// 		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
// 		return
// 	}

// 	userProfile, ok := store.Get("userprofile")
// 	if !ok {
// 		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
// 		return
// 	}

// 	spew.Fdump(w, profile)
// 	spew.Fdump(w, userProfile)

// 	// spew.Fdump(w, profile)
// }

func getUserProfile(id string) (map[string]interface{}, error) {
	managementURL := fmt.Sprintf("https://%s/api/v2/users/%s", viper.GetString(config.Authenticator_Domain), id)

	client := http.DefaultClient
	req, err := http.NewRequest(http.MethodGet, managementURL, nil)

	if err != nil {
		return nil, err
	}

	token, err := auth.GetToken()
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("%s %s", token.Kind, token.Token))
	log.Println(req.Header.Get("Authorization"))

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

// func (rt *Router) Test(w http.ResponseWriter, r *http.Request) {
// 	prof := r.Context().Value(auth.ContextProfile)
// 	userProf := r.Context().Value(auth.ContextUserProfile)

// 	spew.Fdump(w, prof)
// 	spew.Fdump(w, userProf)
// }
