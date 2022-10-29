// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package player

import (
	"fmt"
	"goplanesserver/game"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// msg pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// Player is a middleman between the websocket connection and the lobby.
type Player struct {
	id    int
	lobby game.Lobby

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	msg chan []byte
}

// readPump pumps messages from the websocket connection to the lobby.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (p *Player) readPump() {
	defer func() {
		p.LeaveLobby()
		p.conn.Close()
	}()
	p.conn.SetReadLimit(maxMessageSize)
	p.conn.SetReadDeadline(time.Now().Add(pongWait))
	p.conn.SetPongHandler(func(string) error { p.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, msg, err := p.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		//message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		//log.Println("Incoming: ", message)
		id := extractId(msg)
		//if p.id == 0 {
		//	log.Println("player has 0 id?")
		//	p.setId(id)
		//	log.Println("Added player:", p.id)
		//	if p.lobby.PlayerExists(id) {
		//		log.Println("Rejecting duplicate connection", id)
		//		break
		//	}
		//}
		p.lobby.Update(id, msg)
	}
}

// writePump pumps messages from the lobby to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (p *Player) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		p.conn.Close()
	}()
	for {
		select {
		case message, ok := <-p.msg:
			p.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The lobby closed the channel.
				p.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			//log.Println("writing:", message)
			w, err := p.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(p.msg)
			for i := 0; i < n; i++ {
				//w.Write(newline)
				w.Write(<-p.msg)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			p.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := p.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
func (p *Player) setId(id int) {
	p.id = id
	//p.lobby.players[p.id] = p
}

func extractId(msg []byte) int {
	data := strings.Split(string(msg), ",")
	id, _ := strconv.Atoi(data[0])
	return id
}

func (p *Player) Id() int {
	return p.id
}

func (p *Player) JoinLobby() {
	p.lobby.JoinLobby(p)
}
func (p *Player) LeaveLobby() {
	p.lobby.LeaveLobby(p)
	close(p.msg)
}

func (p *Player) Run() {
	go p.writePump()
	go p.readPump()
}
func (p *Player) Update(msg []byte) error {
	select {
	case p.msg <- msg:
	default:
		close(p.msg)
		return fmt.Errorf("failed to update game")
	}
	return nil
}

func NewPlayer(id int, l game.Lobby, conn *websocket.Conn) game.Player {
	return &Player{id: id, lobby: l, conn: conn, msg: make(chan []byte, 256)}
}
