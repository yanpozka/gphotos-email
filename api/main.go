package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	conf  *oauth2.Config
	store = sessions.NewCookieStore([]byte("Sup4r-s3cRet0"))
)

func init() {
	cred := &struct {
		Cid     string `json:"cid"`
		Csecret string `json:"csecret"`
	}{
		Cid:     os.Getenv("GOOGLE_CLIENT_ID"),
		Csecret: os.Getenv("GOOGLE_SECRET"),
	}
	if cred.Cid == "" || cred.Csecret == "" {
		log.Panic("client_id and secret are required")
	}

	conf = &oauth2.Config{
		ClientID:     cred.Cid,
		ClientSecret: cred.Csecret,
		RedirectURL:  "http://127.0.0.1:8080/auth",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://picasaweb.google.com/data/", // select your own scope here -> https://developers.google.com/identity/protocols/googlescopes
		},
		Endpoint: google.Endpoint,
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	http.HandleFunc("/loginurl", loginURLHandler)
	http.HandleFunc("/auth", authHandler)
	http.HandleFunc("/photos", photosHandler)

	addr := ":8080"
	log.Printf("Listening on %s ...\n", addr)

	log.Println(http.ListenAndServe(addr, nil))
}
