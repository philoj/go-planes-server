package main

import (
	"goplanesserver/lobby"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

const port = ":8080"

func main() {

	// start lobby
	l := lobby.NewLobby()
	go l.Run()

	// serve client connection url
	http.HandleFunc("/lobby/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("socket hit", r.URL.Path)
		id, err := strconv.Atoi(regexp.MustCompile("/lobby/(\\d+)").FindStringSubmatch(r.URL.Path)[1])
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		l.LobbyHandler(id, w, r)
	})

	log.Printf("Listening on port %s", port)
	err := http.ListenAndServe(port, nil)

	// server exit
	if err != nil {
		log.Fatal("Server failure: ", err)
	}
}
