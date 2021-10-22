package game

import (
	"time"

	"gitlab.com/wolfplus/gamespace-lhd/protocol"

	"github.com/lonng/nano"
	"github.com/lonng/nano/component"
	"github.com/lonng/nano/session"
	log "github.com/sirupsen/logrus"

	"gitlab.com/wolfplus/gamespace-lhd/db"
)

const (
	kickResetBacklog = 8
)

var rooms = []*Room{
	NewRoom("初级场", "100金币5包场", "room_pic_108_1", 100, 5),
	NewRoom("中级场", "1000金币5包场", "room_pic_108_2", 1000, 5),
	NewRoom("高级场", "10000金币10包场", "room_pic_108_3", 10000, 10),
	NewRoom("超级场", "100000金币10包场", "room_pic_108_4", 100000, 10),
	NewRoom("豪华场", "1000000金币10包场", "room_pic_108_5", 1000000, 10),
}

var defaultManager = NewManager()

type (
	Manager struct {
		component.Base
		group   *nano.Group       // 广播channel
		players map[int64]*Player // 所有的玩家
		chKick  chan int64        // 退出队列
		chReset chan int64        // 重置队列
	}
)

// business func
func (m *Manager) RoomStatus(s *session.Session, msg []byte) error {
	if p, ok := m.player(s.UID()); ok {
		gamecoin := db.GetGameCoinByUid(s.UID())
		room := p.room
		return s.Response(&protocol.RoomStatus{
			GameStatus:  room.gameStatus,
			Deadline:    room.deadline,
			GameCoin:    gamecoin,
			HistoryList: room.historyList,
		})
	} else {
		return s.Response(&protocol.SendMessageErrorResponse{
			Error: "Session has expired.",
		})
	}
	// return s.Response(nil)
}

func (m *Manager) SendMessage(s *session.Session, req *protocol.SendMessageRequest) error {
	if p, ok := m.player(s.UID()); ok {
		broadcastRes := &protocol.SendMessageResponse{
			Uid:     p.uid,
			Label:   p.userName,
			FaceUri: p.faceUri,
			Message: req.Message,
			Channel: req.Channel,
		}
		if req.Channel == 0 {
			m.group.Broadcast("chat", broadcastRes)
		}
		if req.Channel == 1 {
			log.Infof("Room: %d", p.room.level)
			p.room.group.Broadcast("chat", broadcastRes)
		}
	} else {
		return s.Response(&protocol.SendMessageErrorResponse{
			Error: "Session has expired.",
		})
	}
	return s.Response(nil)
}

func (m *Manager) AutoAngPao(s *session.Session, req *protocol.AutoAngPaoRequest) error {
	if p, ok := m.player(s.UID()); ok {
		r := p.room
		if r.gameStatus == "playing" {
			gamecoin := db.GetGameCoinByUid(p.uid)
			if gamecoin >= rooms[req.RoomNumber].minGameCoin {
				r.autoAngPao(p.uid)
			} else {
				r.group.Broadcast("kickPlayer", &protocol.KickPlayerResponse{
					Uid:   p.uid,
					Error: "游戏币不足，自动退出房间",
				})
			}
		} else {
			deadline := time.Now().UTC().Add(time.Second * time.Duration(3))
			r.group.Broadcast("autoAnimation", &protocol.AutoAnimationResponse{
				Uid:      p.uid,
				Deadline: deadline,
			})
		}
	} else {
		return s.Response(&protocol.SendMessageErrorResponse{Error: "Session has expired."})
	}
	return s.Response(nil)
}

func (m *Manager) GetAngPao(s *session.Session, req *protocol.GetAngPaoRequest) error {
	if p, ok := m.player(s.UID()); ok {
		r := p.room
		if req.Status == "played" {
			r.group.Broadcast("message", &protocol.SendMessageErrorResponse{
				Error: "已经领取红包",
			})
		} else if r.gameStatus == "playing" {
			gamecoin := db.GetGameCoinByUid(p.uid)
			if gamecoin >= rooms[req.RoomNumber].minGameCoin {
				r.getAngPao(p.uid, p.userName, p.faceUri)
			} else {
				r.group.Broadcast("kickPlayer", &protocol.KickPlayerResponse{
					Uid:   p.uid,
					Error: "游戏币不足，自动退出房间",
				})
			}
		} else {
			deadline := time.Now().UTC().Add(time.Second * time.Duration(3))
			r.group.Broadcast("autoAnimation", &protocol.AutoAnimationResponse{
				Uid:      p.uid,
				Deadline: deadline,
			})
		}
	} else {
		return s.Response(&protocol.SendMessageErrorResponse{Error: "Session has expired."})
	}
	return s.Response(nil)
}

func (m *Manager) JoinRoom(s *session.Session, req *protocol.JoinRoomRequest) error {
	uid := req.Uid
	s.Bind(uid)
	log.Infof("玩家: %d登录: %+v", uid, req)

	gamecoin := db.GetGameCoinByUid(uid) // get current gamecoin
	if gamecoin < rooms[req.RoomNumber].minGameCoin {
		return s.Response(&protocol.SendMessageErrorResponse{
			Error: "Không đủ tiền",
		})
	}

	if p, ok := m.player(uid); !ok {
		log.Infof("玩家: %d不在线，创建新的玩家", uid)

		p = newPlayer(s, uid, req.Name, req.FaceUri, rooms[req.RoomNumber])
		m.setPlayer(uid, p) // set player basic info
		rooms[req.RoomNumber].setPlayer(uid, p)
	} else {
		log.Infof("玩家: %d已经在线", uid)
		// 移除广播频道
		m.group.Leave(s)
		p.room.group.Leave(s) // remove the session from the room

		// 重置之前的session
		if prevSession := p.session; prevSession != nil && prevSession != s {
			prevSession.Clear()
			prevSession.Close()
		}

		// 绑定新session
		p.bindSession(s)
		p.bindRoom(rooms[req.RoomNumber])
	}
	m.group.Add(s)
	rooms[req.RoomNumber].group.Add(s)
	log.Infof("number of 玩家: %d", rooms[req.RoomNumber].sessionCount())

	return s.Response(
		&protocol.JoinRoomResponse{
			Name:      req.Name,
			RoomLevel: rooms[req.RoomNumber].level,
			FaceUri:   req.FaceUri,
		})
}

func (m *Manager) GetRooms(s *session.Session, msg []byte) error {
	availRooms := make([]protocol.RoomItem, len(rooms))
	for i := range rooms {
		availRooms[i] = protocol.RoomItem{
			Level:       rooms[i].level,
			Name:        rooms[i].name,
			MinGameCoin: rooms[i].minGameCoin,
			Icon:        rooms[i].icon,
		}
	}

	return s.Response(&protocol.GetRoomResponse{Rooms: availRooms})
}

func NewManager() *Manager {
	return &Manager{
		group:   nano.NewGroup("_SYSTEM_MESSAGE_BROADCAST"),
		players: map[int64]*Player{},
		chKick:  make(chan int64, kickResetBacklog),
		chReset: make(chan int64, kickResetBacklog),
	}
}

func (m *Manager) AfterInit() {
	session.Lifetime.OnClosed(func(s *session.Session) {
		uid := s.UID()
		if p, ok := m.player(uid); ok {
			room := p.room
			room.offline(uid)
			room.group.Leave(s)

		}
		m.offline(uid)
		m.group.Leave(s)
	})

	for _, room := range rooms {
		go room.playingPhase()
	}
}

func (m *Manager) player(uid int64) (*Player, bool) { // get player basic info
	p, ok := m.players[uid]

	return p, ok
}

func (m *Manager) setPlayer(uid int64, p *Player) {
	if _, ok := m.players[uid]; ok {
		log.Warnf("玩家已经存在，正在覆盖玩家， UID=%d", uid)
	}
	m.players[uid] = p
}

func (m *Manager) sessionCount() int {
	return len(m.players)
}

func (m *Manager) offline(uid int64) {
	delete(m.players, uid) // golang func
	log.Infof("manager 玩家: %d从在线列表中删除, 剩余：%d", uid, len(m.players))
}

func NewRoom(level, name, icon string, minGameCoin, total int64) *Room {
	numRobotPlayers := 15
	var robotPlayers []*protocol.RobotPlayer
	for i := 0; i < numRobotPlayers; i++ {
		robotPlayers = append(robotPlayers, NewRobotPlayer())
	}

	return &Room{
		level:           level,
		name:            name,
		minGameCoin:     minGameCoin,
		gameStatus:      "nogame",
		icon:            icon,
		group:           nano.NewGroup(level),
		players:         map[int64]*Player{},
		total:           total,
		robotPlayers:    robotPlayers,
		logInformations: []map[string]string{},
	}
}
