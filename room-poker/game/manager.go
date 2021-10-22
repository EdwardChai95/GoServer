package game

import (
	"gitlab.com/wolfplus/gamespace-lhd/db"
	"gitlab.com/wolfplus/gamespace-lhd/protocol"

	"github.com/lonng/nano"
	"github.com/lonng/nano/component"
	"github.com/lonng/nano/session"
	log "github.com/sirupsen/logrus"
)

const kickResetBacklog = 8

var defaultManager = NewManager()
var rooms = []*Room{
	NewRoom("初级场", 1000, 10, 20),
	NewRoom("中级场", 10000, 100, 200),
	NewRoom("高级场", 1000000, 10000, 20000),
	NewRoom("超级场", 5000000, 100000, 200000),
	// 有限注场次
	NewLimitedRoom("初级场", 2000, 500, 1000, 1000, 2000),
	NewLimitedRoom("中级场", 200000, 5000, 10000, 10000, 20000),
	NewLimitedRoom("高级场", 2000000, 50000, 100000, 100000, 200000),
	NewLimitedRoom("超级场", 20000000, 500000, 1000000, 1000000, 2000000),
}

type (
	Manager struct {
		component.Base
		group   *nano.Group       // 广播channel
		players map[int64]*Player // 所有的玩家
		chKick  chan int64        // 退出队列
		chReset chan int64        // 重置队列
	}
)

// handlers:

func (m *Manager) PlaceBet(s *session.Session, req *protocol.PlaceBetRequest) error {
	if p, ok := m.player(s.UID()); ok {
		if p.PlaceBet(req) {
			return nil
		} else {
			return s.Response(&protocol.SendMessageErrorResponse{
				Error: "无法进行此操作",
			})
		}

	}

	return s.Response(&protocol.SendMessageErrorResponse{
		Error: "error",
	})
}

func (m *Manager) LeaveSeat(s *session.Session, req *protocol.TakeSeatRequest) error {
	if p, ok := m.player(s.UID()); ok {
		room := p.table
		room.offline(s.UID())
	}
	return nil
}

func (m *Manager) TakeSeat(s *session.Session, req *protocol.TakeSeatRequest) error {
	if p, ok := m.player(s.UID()); ok {
		room := p.table
		if ok, message := room.TakeSeat(s.UID(), req); ok {
			// successfully seated
			return s.Response(&protocol.TakeSeatRes{
				SeatNumber:          req.SeatNumber,
				UseableGameCoin:     req.UseableGameCoin,
				SeatedPlayersUpdate: room.seatedPlayers,
			})
		} else {
			return s.Response(&protocol.SendMessageErrorResponse{
				Error: message,
			})
		}
	}
	return s.Response(&protocol.SendMessageErrorResponse{
		Error: "error",
	})
}

// init room in front end
func (m *Manager) RoomStatus(s *session.Session, msg []byte) error {
	logger.Println("roomstatus")
	if p, ok := m.player(s.UID()); ok {
		gamecoin := db.GetGameCoinByUid(s.UID())
		room := p.table

		log.Infof("gameStatus: %v", room.gameStatus)

		return s.Response(&protocol.RoomStatusResponse{
			GameStatus:    room.gameStatus,
			Deadline:      room.deadline,
			GameCoin:      gamecoin,
			SeatedPlayers: room.seatedPlayers,
		})
	} else {
		return s.Response(&protocol.SendMessageErrorResponse{
			Error: "Session has expired.",
		})
	}
	// return s.Response(nil)
}

// join a room so u need to bind a room unless this game does not have second level rooms
func (m *Manager) JoinRoom(s *session.Session, req *protocol.JoinRoomRequest) error {

	jwt := req.Jwt
	uidStr, isValid := db.VerifyJWT(jwt)
	if !isValid {
		return s.Response(&protocol.SendMessageErrorResponse{
			Error: "请稍后再试",
		})
	}

	uid := db.StringToInt64(uidStr)

	s.Bind(uid)
	log.Infof("玩家: %d登录", uid)

	selectedRoom := rooms[req.RoomNumber]
	logger.Printf("Num of tables: %v", len(selectedRoom.tables))

	var selectedTable *Table = selectedRoom.ChooseBestTable()

	gamecoin := db.GetGameCoinByUid(uid) // get current gamecoin
	if gamecoin < selectedRoom.GameCoinToEnter {
		return s.Response(&protocol.SendMessageErrorResponse{
			Error: "游戏币不足",
		})
	}

	if p, ok := m.player(uid); !ok {
		log.Infof("玩家: %d不在线，创建新的玩家", uid)

		p = newPlayer(s, uid, req.Name, req.FaceUri, req.Level, m, selectedTable)
		m.setPlayer(uid, p) // set player basic info i.e. connect a player to this game
		selectedTable.setPlayer(uid, p)
	} else {
		log.Infof("玩家: %d已经在线", uid)
		m.group.Leave(s)
		p.table.group.Leave(s) // remove the session from the room

		// 重置之前的session
		if prevSession := p.session; prevSession != nil && prevSession != s {
			prevSession.Clear()
			prevSession.Close()
		}

		// 绑定新session
		p.bindSession(s, selectedTable)
	}
	m.group.Add(s)
	selectedTable.group.Add(s)

	// at this point u have connected to the room

	// if selectedRoom.gameStatus == GAMESTATUS_NOGAME && selectedRoom.sessionCount() != 0 {
	// 	// here means nobody is in the room or the game yet
	// 	// init the room
	// }

	return s.Response(&protocol.JoinRoomResponse{
		Name:         req.Name,
		FaceUri:      req.FaceUri,
		SelectedRoom: selectedRoom,
	})
}

func (m *Manager) GetRooms(s *session.Session, msg []byte) error {
	return s.Response(rooms)
}

// get player basic info
func (m *Manager) player(uid int64) (*Player, bool) {
	p, ok := m.players[uid]

	return p, ok
}

// add player to players
func (m *Manager) setPlayer(uid int64, p *Player) {
	if _, ok := m.players[uid]; ok {
		log.Warnf("玩家已经存在，正在覆盖玩家， UID=%d", uid)
	}
	m.players[uid] = p
}

// get total count of connected players
func (m *Manager) sessionCount() int {

	log.Infof("几个玩家: %v", len(m.players))
	return len(m.players)
}

// remove players by uid
func (m *Manager) offline(uid int64) {

	delete(m.players, uid) // golang func
	log.Infof("manager 玩家: %d从在线列表中删除, 剩余：%d", uid, len(m.players))
}

// helpers:

func NewManager() *Manager {
	return &Manager{
		group:   nano.NewGroup("PokerManager"),
		players: map[int64]*Player{},
		chKick:  make(chan int64, kickResetBacklog),
		chReset: make(chan int64, kickResetBacklog),
	}
}

func (m *Manager) AfterInit() {
	session.Lifetime.OnClosed(func(s *session.Session) {
		uid := s.UID()
		if p, ok := m.player(uid); ok {
			p.table.offline(uid)
			p.table.group.Leave(s)
		}
		m.offline(uid)
		m.group.Leave(s)
	})
}
