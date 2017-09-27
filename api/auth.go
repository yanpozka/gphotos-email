package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

// user is a retrieved and authentiacted user.
type user struct {
	Sub        string `json:"sub"`
	Name       string `json:"name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Profile    string `json:"profile"`
	Picture    string `json:"picture"`
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, defaultKey)

	retrievedState := session.Values[stateKey]
	if retrievedState != r.URL.Query().Get(stateKey) {
		http.Error(w, fmt.Sprintf("Invalid session state: %q", retrievedState), http.StatusUnauthorized)
		return
	}

	token, err := conf.Exchange(oauth2.NoContext, r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid session state: %q", retrievedState), http.StatusBadRequest)
		return
	}

	client := conf.Client(oauth2.NoContext, token)

	currentUser, err := getUser(w, client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	session.Values[tokenKey] = token
	session.Values[token] = currentUser
	session.Save(r, w)
}

func loginURLHandler(w http.ResponseWriter, r *http.Request) {
	state := randToken()

	session, _ := store.Get(r, defaultKey)
	session.Values[stateKey] = state
	session.Save(r, w)

	data := map[string]string{
		"url": getLoginURL(state),
	}

	json.NewEncoder(w).Encode(&data) // in JSON we trust
}

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func getLoginURL(state string) string {
	return conf.AuthCodeURL(state)
}

func getUser(w http.ResponseWriter, client *http.Client) (*user, error) {
	res, err := client.Get(userinfoURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var u user
	if err := json.NewDecoder(res.Body).Decode(&u); err != nil {
		return nil, err
	}
	if u.Sub == "" {
		return nil, fmt.Errorf("user id not found :(")
	}

	return &u, nil
}

const (
	defaultKey  = "default"
	stateKey    = "state"
	tokenKey    = "token"
	userinfoURL = "https://www.googleapis.com/oauth2/v3/userinfo"
)
