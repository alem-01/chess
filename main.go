package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	hub := newHub()
	go hub.clientsRegister()
	go hub.pickUser()
	go hub.gameMoves()

	r := mux.NewRouter()
	r.HandleFunc("/room", func(w http.ResponseWriter, r *http.Request) {
		roomQueue(hub, w, r)
	})
	r.HandleFunc("/room/{room_id}/{client_id}", func(w http.ResponseWriter, r *http.Request) {
		game(hub, w, r)
	})

	http.Handle("/", r)
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
