package main

import (
	"log"
	"net/http"
	"os"

	"github.com/yanpozka/gphotos-email/api/kvstore"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type handler struct {
	conf  *oauth2.Config
	store kvstore.Store
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cid := os.Getenv("GOOGLE_CLIENT_ID")
	csecret := os.Getenv("GOOGLE_SECRET")

	if cid == "" || csecret == "" {
		log.Panic("client_id and secret are required")
	}

	// oauth configuration:
	//
	conf := &oauth2.Config{
		ClientID:     cid,
		ClientSecret: csecret,
		RedirectURL:  "http://127.0.0.1:8080/auth",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://picasaweb.google.com/data/", // select your own scope here -> https://developers.google.com/identity/protocols/googlescopes
			// TODO(yandry): add Gmail scope here
		},
		Endpoint: google.Endpoint,
	}

	// db configuration
	//
	path := os.Getenv("BOLTDB_PATH")
	if path == "" {
		path = "data.db"
	}

	db, err := kvstore.NewBoltDBStore(path, "")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		os.Remove(path)
		if err := db.Close(); err != nil {
			log.Println(err)
		}
	}()

	h := handler{conf: conf, store: db}

	http.HandleFunc("/loginurl", h.loginURL)
	http.HandleFunc("/auth", h.auth)
	http.HandleFunc("/photos", h.photoList)

	addr := ":8080"
	log.Printf("Listening on muy lindo %s ...\n", addr)

	log.Println(http.ListenAndServe(addr, nil))
}
