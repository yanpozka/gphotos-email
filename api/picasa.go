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
	xmlpath "gopkg.in/xmlpath.v2"
)

var (
	entriesPath    = xmlpath.MustCompile("/feed/entry")
	entryURLPath   = xmlpath.MustCompile("group/content/@url")
	entryDatePath  = xmlpath.MustCompile("timestamp") // has a prefix `gphoto:`
	entryTitlePath = xmlpath.MustCompile("title")
)

func (h *handler) photoList(w http.ResponseWriter, r *http.Request) {
	si, ok := r.Context().Value(sessionKey).(*sessionInfo)
	if !ok {
		log.Panic("sessionInfo not found")
	}

	client := h.conf.Client(context.Background(), si.GToken)

	var root *xmlpath.Node
	{
		u := makePicasaURL(si.User.Sub, 11)
		log.Printf("Getting url: %v", u)

		res, err := client.Do(&http.Request{Method: http.MethodGet, URL: u})
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
		Images []*image `json:"images"`
	}

	itr := entriesPath.Iter(root)
	for itr.Next() {
		u, found := entryURLPath.String(itr.Node())
		if !found {
			log.Printf("URL not found in entry: %q", itr.Node().String())
			continue
		}
		t, _ := entryTitlePath.String(itr.Node())

		d, found := entryDatePath.String(itr.Node())
		if !found {
			log.Printf("date not found in entry: %q", itr.Node().String())
			continue
		}

		pd, err := parseTime(d)
		if err != nil {
			log.Println("error parsing unix time: ", err)
			continue
		}

		info.Images = append(info.Images, &image{
			Title: t,
			URL:   u,
			Date:  pd,
		})
	}

	if err := json.NewEncoder(w).Encode(info); err != nil {
		log.Println(err)
	}
}

func makePicasaURL(id string, maxResults int) *url.URL {
	u, _ := url.Parse(fmt.Sprintf("https://picasaweb.google.com/data/feed/api/user/%s", id))

	q := u.Query()
	// q.Set("kind", "photo")
	// q.Set("imgmax", "1600")
	q.Set("max-results", fmt.Sprintf("%d", maxResults))
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
