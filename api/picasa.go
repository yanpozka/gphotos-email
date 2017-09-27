package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	xmlpath "gopkg.in/xmlpath.v2"
)

var (
	entriesPath    = xmlpath.MustCompile("/feed/entry")
	entryURLPath   = xmlpath.MustCompile("content/@src")
	entryDatePath  = xmlpath.MustCompile("timestamp") // has a prefix `gphoto:`
	entryTitlePath = xmlpath.MustCompile("title")
)

func photosHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, defaultKey)

	// IMPROVE (yandry): add middleware here
	//
	var currentUser *user
	var token *oauth2.Token
	{
		tok, contains := session.Values[tokenKey]
		if !contains {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		token = tok.(*oauth2.Token)

		cuser, contains := session.Values[token]
		if !contains {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		currentUser = cuser.(*user)
	}
	client := conf.Client(context.Background(), token)

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
		// fmt.Println(pd)
		info.imgs = append(info.imgs, &image{
			Title: t, URL: u, Date: pd,
		})
	}

	if err := json.NewEncoder(w).Encode(info); err != nil {
		log.Println(err)
	}
}

func makePicasaURL(id string, maxResults int) *url.URL {
	u, _ := url.Parse(fmt.Sprintf("https://picasaweb.google.com/data/feed/api/user/%s", id))

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
