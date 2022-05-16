package main

import (
	"goplanesserver/lobby"
	"log"
	"net/http"
)

const port = ":8080"

func main() {

	// start lobby
	l := lobby.NewLobby()
	go l.Run()

	// serve client connection url
	http.HandleFunc("/join", func(w http.ResponseWriter, r *http.Request) {
		l.LobbyHandler(w, r)
	})

	log.Printf("Listening on port %s", port)
	err := http.ListenAndServe(port, nil)

	// server exit
	if err != nil {
		log.Fatal("Server failure: ", err)
	}
}
