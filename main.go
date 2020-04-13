package main

import (
	"log"
	"net/http"
)

func main() {

	// start hub
	hub := newHub()
	go hub.run()

	// serve client connection url
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	// start and listen server until error
	err := http.ListenAndServe(":8080", nil)

	// server exit
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
