// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type Broadcast struct {
	id  int
	msg []byte
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[int]*Client

	// Inbound messages from the clients.
	broadcast chan Broadcast

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan Broadcast),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[int]*Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.playerId] = client
		case client := <-h.unregister:
			if _, ok := h.clients[client.playerId]; ok {
				delete(h.clients, client.playerId)
				close(client.send)
			}
		case b := <-h.broadcast:
			for id := range h.clients {
				// do not write back to the same client
				if id != 0 && b.id != id {
					//log.Println("Sending to id:", client.playerId, string(b.msg))
					select {
					case h.clients[id].send <- b.msg:
					default:
						close(h.clients[id].send)
						delete(h.clients, id)
					}
				} else {
					//log.Println("Skipping id:", id)
				}
			}
		}
	}
}
