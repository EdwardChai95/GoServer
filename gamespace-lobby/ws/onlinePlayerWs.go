package ws

import (
	"github.com/gorilla/websocket"
)

type OnlinePlayerReq struct {
	Type string
	//

	GuestAcc int64
	UserName string

	Signature string
	GameCoin int64

	//more to be extended
}

type OnlinePlayerRes struct {
	OnlinePlayer []OnlinePlayer
}

/////

type OnlinePlayerPool struct {
	OnlinePlayer   []OnlinePlayer
	ConnectionPool []*websocket.Conn
}

type OnlinePlayer struct {
	GuestAcc int64
	UserName string
	Signature string
	GameCoin int64
}

// CPR = OnlinePlayer pool runtime , newCRes := CRes{OnlinePlayerPool:CPR}
var OPR OnlinePlayerPool
