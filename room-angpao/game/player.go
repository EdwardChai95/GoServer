package game

import (
	"sync"

	"github.com/lonng/nano/session"
	log "github.com/sirupsen/logrus"
)

type (
	Player struct {
		uid            int64  // 用户ID
		faceUri        string // 头像地址
		userName       string // 玩家名字
		session        *session.Session
		logger         *log.Entry // 日志
		room           *Room
		awarded        int64
		auto           int64
		logInformation map[string]string
		sync.RWMutex
	}
)

func newPlayer(s *session.Session, uid int64, name string, faceUri string,
	room *Room) *Player {
	p := &Player{
		uid:      uid,
		userName: name,
		faceUri:  faceUri,
		logger:   log.WithField("player", uid),
	}

	p.bindSession(s)
	p.bindRoom(room)

	return p
}

func (p *Player) bindRoom(room *Room) {
	p.room = room
}

func (p *Player) bindSession(s *session.Session) {
	p.session = s
	p.session.Set(kCurPlayer, p)
}

func (p *Player) removeSession() {
	p.session.Remove(kCurPlayer)
	p.session = nil
}

func (p *Player) Uid() int64 {
	return p.uid
}
