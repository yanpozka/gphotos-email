package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

func (h *handler) auth(w http.ResponseWriter, r *http.Request) {
	receivedState := r.URL.Query().Get(stateKey)

	savedState, _ := h.store.Get([]byte(receivedState)) // TODO(yandry): 500 in case of err
	if savedState == nil {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	if string(savedState) != receivedState {
		http.Error(w, fmt.Sprintf("Saved state: %q mismatch received state from url: %q", savedState, receivedState), http.StatusUnauthorized)
		return
	}

	tokenObj, err := h.conf.Exchange(oauth2.NoContext, r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed exchange with code=%q", r.URL.Query().Get("code")), http.StatusInternalServerError)
		return
	}

	currentUser, err := getUser(w, h.conf.Client(oauth2.NoContext, tokenObj))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	genToken := randToken()
	var buff bytes.Buffer

	sessionInfo := map[string]interface{}{
		"user":   currentUser,
		"gtoken": tokenObj,
	}
	if err := gob.NewEncoder(&buff).Encode(sessionInfo); err != nil {
		http.Error(w, "encoding session info error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	h.store.Set([]byte(genToken), buff.Bytes())

	if err := json.NewEncoder(w).Encode(map[string]string{"token": genToken}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *handler) loginURL(w http.ResponseWriter, r *http.Request) {
	state := randToken()
	epoch := fmt.Sprintf("%d", time.Now().UnixNano())

	h.store.Set([]byte(state), []byte(epoch)) // TODO(yandry): 500 in case of err

	data := map[string]string{"url": h.getLoginURL(state)}

	if err := json.NewEncoder(w).Encode(&data); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError)+" "+err.Error(), http.StatusInternalServerError)
	}
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
		return nil, fmt.Errorf("getUser(): user id not found :(")
	}

	return &u, nil
}

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func (h *handler) getLoginURL(state string) string {
	return h.conf.AuthCodeURL(state)
}

const (
	stateKey    = "state"
	userinfoURL = "https://www.googleapis.com/oauth2/v3/userinfo"
)
