package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/textproto"

	"github.com/jordan-wright/email"
	"golang.org/x/net/context"
	gmail "google.golang.org/api/gmail/v1"
)

func (h *handler) sendEmail(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var data struct {
		URL       string `json:"url"`
		Text      string `json:"text"`
		Subject   string `json:"subject,omitempty"`
		EmailAddr string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "invalid JSON format", http.StatusBadRequest)
		return
	}
	if data.URL == "" || data.EmailAddr == "" {
		http.Error(w, "URL and email address are required", http.StatusBadRequest)
		return
	}

	si, ok := r.Context().Value(sessionKey).(*sessionInfo)
	if !ok {
		log.Panic("sessionInfo not found")
	}

	// fetch image
	res, err := http.Get(data.URL)
	if err != nil {
		log.Printf("%#v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer res.Body.Close()

	srv, err := gmail.New(h.conf.Client(context.Background(), si.GToken))
	panicIfErr(err)

	var message gmail.Message

	e := &email.Email{
		To:      []string{data.EmailAddr},
		From:    fmt.Sprintf("%s <%s>", si.User.GivenName, data.EmailAddr),
		Subject: data.Subject,
		Text:    []byte(data.Text),
		Headers: textproto.MIMEHeader{},
	}
	// pipe response reader
	e.Attach(res.Body, "photo.jpg", res.Header.Get("Content-Type"))

	bb, err := e.Bytes()
	panicIfErr(err)
	message.Raw = base64.URLEncoding.EncodeToString(bb)

	_, err = srv.Users.Messages.Send("me", &message).Do()
	panicIfErr(err)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "email sent",
	})
}
