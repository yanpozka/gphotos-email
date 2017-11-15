package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"log"
	"net/http"
	"time"

	"github.com/yanpozka/gphotos-email/api/kvstore"
)

const authHeader = "X-Auth-Token"

type sessionCtx struct{}

var sessionKey sessionCtx

func (h *handler) authMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tok := r.Header.Get(authHeader)
		if tok == "" {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		data, err := h.store.Get(kvstore.DefaultBucket, []byte(tok))
		panicIfErr(err)
		if data == nil {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		buff := bytes.NewBuffer(data)
		var si sessionInfo

		if err := gob.NewDecoder(buff).Decode(&si); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), sessionKey, &si)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *handler) logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %q (%s) Time consumed: %s", r.Method, r.URL, r.RemoteAddr, time.Since(start))
	})
}

func (h *handler) commonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "MioServer")
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		next.ServeHTTP(w, r)
	})
}
