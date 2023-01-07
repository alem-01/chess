package main

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type Player struct {
	ChessColor string // black or white
}

type Room struct {
	Players    map[string]*Player
	MovesCount int
}

type GameMove struct {
	ClientRoom ClientRoom
	Move       string // black or white
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[string]*Client
	newUser chan string

	rooms      map[string]*Room
	gamesMoves chan *GameMove

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
		newUser:    make(chan string, 100),
		rooms:      make(map[string]*Room),
		gamesMoves: make(chan *GameMove, 100),
	}
}

func (h *Hub) clientsRegister() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.ID] = client
			h.newUser <- client.ID
		case client := <-h.unregister:
			if c, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(c.send)
			}
		}
	}
}

func (h *Hub) pickUser() {
	userIDs := make(map[string]bool, 0)
	for userID := range h.newUser {
		if h.clients[userID].RoomID != "" {
			continue
		}

		userIDs[userID] = true
		if len(userIDs) >= 2 {
			var (
				roomID = uuid.NewString()
				i      = 0
			)

			h.rooms[roomID] = &Room{
				Players: make(map[string]*Player),
			}

			for k := range userIDs {
				if i == 2 {
					break
				}
				h.clients[k].RoomID = roomID

				cr := ClientRoom{
					ClientID: k,
					RoomID:   roomID,
				}

				h.rooms[roomID].Players[k] = &Player{
					ChessColor: getChessColor(i),
				}

				h.clients[k].send <- cr.json()

				h.unregister <- h.clients[k]
				delete(userIDs, k)
				i++
			}

			fmt.Printf("%#v\n", h.rooms[roomID])
		}
	}
}

func (h *Hub) gameMoves() {
	for gameMove := range h.gamesMoves {
		fmt.Printf("%#v\n", gameMove)

		for k := range h.rooms[gameMove.ClientRoom.RoomID].Players {
			if k == gameMove.ClientRoom.ClientID {
				continue
			}
			opponentID := k
			client, ok := h.clients[opponentID]
			if !ok {
				continue
			}

			client.send <- []byte(gameMove.Move)
		}
	}
}

func getChessColor(n int) string {
	if n%2 == 0 {
		return "white"
	}
	return "black"
}

type ClientRoom struct {
	ClientID string `json:"client_id"`
	RoomID   string `json:"room_id"`
}

func (c *ClientRoom) json() []byte {
	b, _ := json.Marshal(c)
	return b
}
