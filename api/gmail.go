package main

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"

	"golang.org/x/net/context"
	gmail "google.golang.org/api/gmail/v1"
)

func (h *handler) sendEmail(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// {
	//	"url": "http://",
	//	"text": "abcde 1234"
	// }
	data := map[string]string{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "invalid JSON format", http.StatusBadRequest)
		return
	}

	si, ok := r.Context().Value(sessionKey).(*sessionInfo)
	if !ok {
		log.Panic("sessionInfo not found")
	}

	srv, err := gmail.New(h.conf.Client(context.Background(), si.GToken))
	panicIfErr(err)

	var message gmail.Message

	// text := base64.URLEncoding.EncodeToString([]byte(data["text"]))
	text := data["text"]

	// payload := gmail.MessagePart{
	// 	Headers: []*gmail.MessagePartHeader{
	// 		{Name: "From", Value: "ypozoka@gmail.com"},
	// 		{Name: "To", Value: "ypozoka@gmail.com"},
	// 		{Name: "Subject", Value: "Este si carajo"},
	// 	},
	// 	Body: &gmail.MessagePartBody{Data: text},
	// }
	// message.Payload = &payload

	messageStr := []byte("From: ypozoka@gmail.com\r\n" +
		"To: ypozoka@gmail.com\r\n" +
		"Subject: Un two y tres\r\n\r\n" +
		text)
	message.Raw = base64.URLEncoding.EncodeToString(messageStr)

	_, err = srv.Users.Messages.Send("me", &message).Do()
	panicIfErr(err)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "email sent",
	})
}
