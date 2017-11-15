package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/yanpozka/gphotos-email/api/kvstore"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gmail "google.golang.org/api/gmail/v1"
)

var (
	// Version as semantic notation + hash commit
	Version string
	// BuildTime or compile time
	BuildTime string
)

type handler struct {
	conf  *oauth2.Config
	store kvstore.Store
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix(fmt.Sprintf("[%s] ", Version))

	cid := os.Getenv("GOOGLE_CLIENT_ID")
	csecret := os.Getenv("GOOGLE_SECRET")
	redirectURL := os.Getenv("REDIRECT_URL")

	if cid == "" || csecret == "" || redirectURL == "" {
		log.Panic("client_id, secret and redirect URL are required")
	}

	// oauth configuration:
	//
	conf := &oauth2.Config{
		ClientID:     cid,
		ClientSecret: csecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://picasaweb.google.com/data/",
			gmail.GmailSendScope,
		},
		Endpoint: google.Endpoint,
	}

	// db configuration
	//
	path := os.Getenv("BOLTDB_PATH")
	if path == "" {
		path = "data.db"
	}

	db, err := kvstore.NewBoltDBStore(path)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	addr := os.Getenv("API_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: createRouter(&handler{conf: conf, store: db}),

		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 8 * time.Second,
		WriteTimeout:      15 * time.Second,
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill)

	go func() {
		log.Printf("Starting listening on address %q BuildTime: %s", addr, BuildTime)
		log.Println(srv.ListenAndServe())
	}()

	log.Printf("Got signal: %v", <-ch)
}

func createRouter(h *handler) http.Handler {
	router := httprouter.New()

	router.PanicHandler = func(w http.ResponseWriter, r *http.Request, v interface{}) {
		log.Printf("Recovering with value: %+v\nrequest: %s %q %s", v, r.Method, r.URL, r.RemoteAddr)
		debug.PrintStack()
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Detected Not Found: %s %q %s", r.Method, r.URL, r.RemoteAddr)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})

	router.MethodNotAllowed = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Detected Method Now Allowed: %s %q %s", r.Method, r.URL, r.RemoteAddr)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	})

	stdCh := alice.New(h.logger, h.commonHeaders)
	authCh := stdCh.Append(h.authMW)

	// routers:
	{
		router.Handler(http.MethodGet, "/loginurl", stdCh.ThenFunc(h.loginURL))
		router.Handler(http.MethodGet, "/auth", stdCh.ThenFunc(h.auth))

		router.Handler(http.MethodGet, "/photos", authCh.ThenFunc(h.photoList))
		router.Handler(http.MethodPost, "/email", authCh.ThenFunc(h.sendEmail))
	}

	return router
}

func panicIfErr(err error) {
	if err != nil {
		log.Panicf("%#v", err)
	}
}
