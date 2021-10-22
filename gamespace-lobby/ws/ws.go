package ws

import (
	"github.com/gorilla/websocket"
)

type LobbySession struct {
	Uid      string
	Username string
	FaceUri  string
	ClubID   string
	Level    int
	Conn     *websocket.Conn
}

type SocketRequest struct {
	Type string
	Data map[string]interface{}
}

// type InitSession struct {
// 	Uid      string
// 	Username string
// 	FaceUri  string
// }

// idiot code below
type CReq struct {
	Type string
	//

	UserName string
	Message  string
}

type CRes struct {
	Chat []Chat
}

/////

type ChatPool struct {
	Chat           []Chat
	ConnectionPool []*websocket.Conn
}

type Chat struct {
	UserName string
	Message  string
}

// CPR = chat pool runtime , newCRes := CRes{ChatPool:CPR}
var CPR ChatPool
