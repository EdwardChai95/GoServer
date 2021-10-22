package game

import (
	"gitlab.com/wolfplus/gamespace-lhd/protocol"

	"github.com/lonng/nano"
	"github.com/lonng/nano/component"
	"github.com/lonng/nano/session"
	log "github.com/sirupsen/logrus"

	"gitlab.com/wolfplus/gamespace-lhd/db"
)

const kickResetBacklog = 8

//var rooms = []*Room{NewRoom(0, "天子一号桌"), NewRoom(0, "天子二号桌"), NewRoom(0, "天子三号桌"), NewRoom(0, "天子四号桌"), NewRoom(1, "通比一号桌"), NewRoom(1, "通比二号桌"), NewRoom(0, "天子五号桌"), NewRoom(0, "天子六号桌"), NewRoom(0, "天子七号桌")}
var rooms = []*Room{NewRoom(0, "Bàn một"), NewRoom(0, "Bảng hai"), NewRoom(0, "Bảng ba"), NewRoom(0, "Bảng bốn"), NewRoom(1, "Bàn một"), NewRoom(1, "Bảng hai"), NewRoom(0, "Bảng năm"), NewRoom(0, "Bảng sáu"), NewRoom(0, "Bảng bảy")}

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
func (m *Manager) PlaceBet(s *session.Session, req *protocol.PlaceBetRequest) error {
	if p, ok := m.player(s.UID()); ok {
		currentRoom := p.room

		// log.Infof("key: %d", req.Key)
		// log.Infof("amount: %d", req.Amount)
		if ok, errmsg := p.PlaceBet(req.Key, req.Amount); ok {
			currentRoom.group.Broadcast("updateGambler", &protocol.PlaceBetResponse{
				Key:    req.Key,
				Total:  currentRoom.gamblers[req.Key].Totalbetting,
				MyBet:  p.GetBet(req.Key),
				Uid:    p.uid,
				Amount: req.Amount,
			})
		} else {
			return s.Response(&protocol.SendMessageErrorResponse{Error: errmsg})
		}
	}
	return s.Response(nil)
}

func (m *Manager) RoomStatus(s *session.Session, msg []byte) error {
	if p, ok := m.player(s.UID()); ok {
		gamecoin := db.GetGameCoinByUid(s.UID())
		room := p.room
		return s.Response(&protocol.RoomStatus{
			GameStatus:  room.gameStatus,
			Deadline:    room.deadline,
			GameCoin:    gamecoin,
			BankerCoins: room.bankerCoins,
			BankerImg:   room.bankerImg,
			HistoryList: room.historyList,
		})
	} else {
		return s.Response(&protocol.SendMessageErrorResponse{
			Error: "Session has expired.",
		})
	}
	// return s.Response(nil)
}

func (m *Manager) JoinRoom(s *session.Session, req *protocol.JoinRoomRequest) error {
	uid := req.Uid
	s.Bind(uid)
	log.Infof("玩家: %d登录: %+v", uid, req)

	gamecoin := db.GetGameCoinByUid(uid) // get current gamecoin
	if gamecoin < rooms[req.RoomNumber].minGameCoin {
		return s.Response(&protocol.SendMessageErrorResponse{
			//Error: "游戏币不足",
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

	// log.Infof("ROOM JOINED: %d", rooms[req.RoomNumber].label)
	// log.Infof("number of 玩家: %d", rooms[req.RoomNumber].sessionCount())
	if rooms[req.RoomNumber].gameStatus == "nogame" && rooms[req.RoomNumber].sessionCount() > 0 {
		go rooms[req.RoomNumber].bettingPhase()
	}

	return s.Response(
		&protocol.JoinRoomResponse{
			Name:    req.Name,
			FaceUri: req.FaceUri,
			Chip:    rooms[req.RoomNumber].chip,
		})
}

func (m *Manager) GetRooms(s *session.Session, msg []byte) error {

	availRooms := make([]protocol.RoomItem, len(rooms))
	for i := range rooms {
		availRooms[i] = protocol.RoomItem{
			Title:       rooms[i].title,
			Label:       rooms[i].label,
			MinGameCoin: rooms[i].minGameCoin,
			Chip:        rooms[i].chip,
			RoomType:    rooms[i].roomType,
		}
	}

	return s.Response(&protocol.GetRoomResponse{Rooms: availRooms})
}

func NewRoom(roomType int, label string) *Room {
	//title := "百人场"
	title := "Vòng lớn"
	if roomType == 1 {
		//title = "通比场"
		title = "Tomby tròn"
	}

	numRobotPlayers := 10
	var robotPlayers []*protocol.RobotPlayer
	for i := 0; i < numRobotPlayers; i++ {
		robotPlayers = append(robotPlayers, NewRobotPlayer())
	}

	return &Room{
		group:               nano.NewGroup(label),
		title:               title,
		label:               label,
		minGameCoin:         10000,
		roomType:            roomType,
		gamblers:            NewGamblers(),
		banker:              NewGambler("banker"),
		chip:                []int64{100, 1000, 10000, 100000, 1000000, 5000000},
		players:             map[int64]*Player{},
		gameStatus:          "nogame",
		bankerRounds:        0,
		bankerImg:           randomRange(1, 10),
		bankerCoins:         randomRange(300000000, 900000000),
		totalPlayerWinnings: 0,
		robotPlayers:        robotPlayers,
		logInformations:     []map[string]string{},
	}
}

func NewGamblers() []*protocol.Gambler {
	return []*protocol.Gambler{
		NewGambler("青龙"),
		NewGambler("白虎"),
		NewGambler("朱雀"),
		NewGambler("玄武"),
	}
}

func NewGambler(title string) *protocol.Gambler {
	return &protocol.Gambler{Title: title, Totalbetting: 0, Odds: 0, Combo: "",
		RoundCards: []*protocol.Card{}}
}

func NewManager() *Manager {
	return &Manager{
		group:   nano.NewGroup("PinshiManager"),
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

	//for _, room := range rooms {
	//	go room.bettingPhase()
	//}
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
