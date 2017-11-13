package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/yanpozka/gphotos-email/api/kvstore"
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

type sessionInfo struct {
	User   *user         `json:"user"`
	GToken *oauth2.Token `json:"gtoken"`
}

func (h *handler) auth(w http.ResponseWriter, r *http.Request) {
	receivedState := r.URL.Query().Get(stateKey)

	savedState, err := h.store.Get(kvstore.DefaultBucket, []byte(receivedState))
	if err != nil || savedState == nil {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
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

	var buff bytes.Buffer
	si := sessionInfo{User: currentUser, GToken: tokenObj}

	if err := gob.NewEncoder(&buff).Encode(si); err != nil {
		http.Error(w, "encoding session info error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.store.Set(kvstore.DefaultBucket, []byte(receivedState), buff.Bytes())
	panicIfErr(err)

	w.WriteHeader(http.StatusOK)
}

func (h *handler) loginURL(w http.ResponseWriter, r *http.Request) {
	state := randToken()
	epoch := fmt.Sprintf("%d", time.Now().UnixNano())

	err := h.store.Set(kvstore.DefaultBucket, []byte(state), []byte(epoch))
	panicIfErr(err)

	data := map[string]string{
		"url":   h.getLoginURL(state),
		"token": state,
	}
	log.Print(data["url"])

	if err := json.NewEncoder(w).Encode(&data); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError)+" "+err.Error(), http.StatusInternalServerError)
	}
}

func (h *handler) getLoginURL(state string) string {
	return h.conf.AuthCodeURL(state)
}

func panicIfErr(err error) {
	if err != nil {
		log.Panic(err)
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

const (
	stateKey    = "state"
	userinfoURL = "https://www.googleapis.com/oauth2/v3/userinfo"
)
