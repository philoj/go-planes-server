package lobby

import (
	"github.com/gorilla/websocket"
	"goplanesserver/game"
	"goplanesserver/players"
	"log"
	"net/http"
)

type SocketPayload struct {
	Id  int
	Msg []byte
}

// Lobby the game lobby for players to join
type Lobby struct {
	// Registered clients.
	players map[int]game.Player

	//  msg messages to the clients.
	msg chan SocketPayload

	// join requests from the clients.
	join chan game.Player

	// leave requests from clients.
	leave chan game.Player
}

func NewLobby() game.Lobby {
	return &Lobby{
		msg:     make(chan SocketPayload),
		join:    make(chan game.Player),
		leave:   make(chan game.Player),
		players: make(map[int]game.Player),
	}
}

func (l *Lobby) Run() {
	for {
		select {
		case p := <-l.join:
			log.Printf("player %d join", p.Id())
			l.players[p.Id()] = p
		case client := <-l.leave:
			if _, ok := l.players[client.Id()]; ok {
				delete(l.players, client.Id())
			}
		case b := <-l.msg:
			for id := range l.players {
				// do not write back to the same player
				if b.Id == id {
					continue
				}
				//log.Println("Sending to id:", p.playerId, string(b.msg))
				err := l.players[id].Update(b.Msg)
				if err != nil {
					delete(l.players, id)
				}
			}
		}
	}
}

func (l *Lobby) JoinLobby(p game.Player) {
	l.join <- p
}
func (l *Lobby) LeaveLobby(p game.Player) {
	l.leave <- p
}
func (l *Lobby) Update(id int, msg []byte) {
	l.msg <- SocketPayload{id, msg}
}

func (l *Lobby) CreatePlayer(id int, conn *websocket.Conn) game.Player {
	p := players.NewPlayer(id, l, conn)
	l.JoinLobby(p)
	p.Run()
	return p
}

func (l *Lobby) PlayerExists(id int) bool {
	_, exists := l.players[id]
	return exists
}

var u = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		log.Print("origin: ", r.Header.Get("origin"))
		return r.Header.Get("origin") == "http://localhost:8081" // FIXME better origin value
	},
}

// LobbyHandler handles websocket requests from the peer.
func (l *Lobby) LobbyHandler(id int, w http.ResponseWriter, r *http.Request) {
	conn, err := u.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	l.CreatePlayer(id, conn)
}
