package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gopkg.in/xmlpath.v2"
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

var (
	conf  *oauth2.Config
	store = sessions.NewCookieStore([]byte("Sup4r-s3cRet0"))

	entriesPath    = xmlpath.MustCompile("/feed/entry")
	entryURLPath   = xmlpath.MustCompile("content/@src")
	entryDatePath  = xmlpath.MustCompile("timestamp") // gphoto:
	entryTitlePath = xmlpath.MustCompile("title")
)

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func init() {
	cred := &struct {
		Cid     string `json:"cid"`
		Csecret string `json:"csecret"`
	}{
		Cid:     os.Getenv("GOOGLE_PROJECT_CLIENT_ID"),
		Csecret: os.Getenv("GOOGLE_PROJECT_SECRET"),
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

func getLoginURL(state string) string {
	return conf.AuthCodeURL(state)
}

func getUser(w http.ResponseWriter, client *http.Client) (*user, error) {
	res, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
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

func makePicasaURL(id string, maxResults int) *url.URL {
	u, err := url.Parse(fmt.Sprintf("https://picasaweb.google.com/data/feed/api/user/%s", id))
	if err != nil {
		log.Fatal("Hey You!" + err.Error())
	}

	q := u.Query()
	q.Set("kind", "photo")
	q.Set("max-results", fmt.Sprintf("%d", maxResults))
	q.Set("imgmax", "1600")
	u.RawQuery = q.Encode()

	return u
}

func parseTime(ts string) (time.Time, error) {
	nt, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(nt/1000, nt%1000), nil
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, defaultKey)

	retrievedState := session.Values[stateKey]
	if retrievedState != r.URL.Query().Get(stateKey) {
		http.Error(w, fmt.Sprintf("Invalid session state: %q", retrievedState), http.StatusUnauthorized)
		return
	}

	tok, err := conf.Exchange(oauth2.NoContext, r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid session state: %q", retrievedState), http.StatusBadRequest)
		return
	}

	client := conf.Client(oauth2.NoContext, tok)

	currentUser, err := getUser(w, client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var root *xmlpath.Node
	{
		u := makePicasaURL(currentUser.Sub, 3)

		res, err := client.Do(&http.Request{Method: "GET", URL: u})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer res.Body.Close()

		root, err = xmlpath.Parse(res.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	type image struct {
		Date       time.Time
		Title, URL string
	}

	var info struct {
		AuthorName   string
		AuthorPicURL string
		imgs         []*image
	}
	info.AuthorName = currentUser.Name
	info.AuthorPicURL = currentUser.Picture

	itr := entriesPath.Iter(root)
	for itr.Next() {
		u, _ := entryURLPath.String(itr.Node())
		t, _ := entryTitlePath.String(itr.Node())

		d, found := entryDatePath.String(itr.Node())
		if !found {
			log.Println("date not found in entry: %q", itr.Node().String())
			continue
		}

		pd, err := parseTime(d)
		if err != nil {
			log.Println("error parsing unix time: ", err)
			continue
		}
		fmt.Println(pd)
		info.imgs = append(info.imgs, &image{
			Title: t, URL: u, Date: pd,
		})
	}

	var first, t string
	if len(info.imgs) > 0 {
		first = info.imgs[0].URL
		t = info.imgs[0].Title
	}

	//
	// TODO: find a way to render JSON response to a single web app
	//
	w.Write([]byte(fmt.Sprintf(`<html>
		<title>Welcome to GO Fotos</title>
		<body>
			<h3>Hola %s! <img src="%s" width="128" height="128"/></h3>
			<br>
			<br>
			<img src="%s" width="128" height="128" alt="%s"/>
		</body>
	</html>`, info.AuthorName, info.AuthorPicURL, first, t,
	)))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	state := randToken()

	session, _ := store.Get(r, defaultKey)
	session.Values[stateKey] = state
	session.Save(r, w)

	//
	// TODO: render JSON response with URL
	//
	w.Write([]byte(fmt.Sprintf(`<html>
		<title>GO Fotos</title>
		<body>
			<br><h2>Fotos para todos!</h2><br>
			<a href='%s'><button>Login with Google!</button></a>
		</body>
	</html>`,
		getLoginURL(state),
	)))
}

func main() {
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/auth", authHandler)

	addr := ":8080"
	log.Printf("Listening on %s ...\n", addr)

	log.Println(http.ListenAndServe(addr, nil))
}

const (
	defaultKey = "default"
	stateKey   = "state"
)
