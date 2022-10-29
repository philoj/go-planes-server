package game

import (
	"github.com/gorilla/websocket"
	"net/http"
)

type Player interface {
	Id() int
	JoinLobby()
	LeaveLobby()
	Run()
	Update(msg []byte) error
}

type Lobby interface {
	Run()
	JoinLobby(p Player)
	LeaveLobby(p Player)
	Update(id int, msg []byte)
	CreatePlayer(id int, conn *websocket.Conn) Player
	LobbyHandler(id int, w http.ResponseWriter, r *http.Request)
}
