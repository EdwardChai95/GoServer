package game

import (
	"strconv"
	"time"

	"gitlab.com/wolfplus/gamespace-lhd/protocol"

	"github.com/lonng/nano"
	"github.com/lonng/nano/component"
	"github.com/lonng/nano/session"
	log "github.com/sirupsen/logrus"

	"gitlab.com/wolfplus/gamespace-lhd/db"
)

const (
	kickResetBacklog            = 8
	maxRobotBettingOffset int64 = 2
)

var clubs = []*Club{}

var rooms = []*Room{
	//NewRoom("初级场", "凯旋门", "room_pic_108_1", "fp_01", 100, 100, 0, []int64{1, 2, 5, 10}),
	NewRoom("Trận sơ cấp", "Khải Hoàn Môn", "room_pic_108_1", "fp_01", 100, 100, 0, []int64{1, 2, 5, 10}),
	//NewRoom("中级场", "威尼斯场", "room_pic_108_2", "fp_02", 10000, 500, 0, []int64{10, 20, 50, 100}),
	NewRoom("Trận trung cấp", "Cánh đồng Venice", "room_pic_108_2", "fp_02", 10000, 500, 0, []int64{10, 20, 50, 100}),
	//NewRoom("高级场", "拉斯维加斯", "room_pic_108_3", "fp_03", 100000, 10000, 0, []int64{100, 200, 500, 1000}),
	NewRoom("Trận cao cấp", "Las Vegas", "room_pic_108_3", "fp_03", 100000, 10000, 0, []int64{100, 200, 500, 1000}),
	//NewRoom("超级场", "加勒比游轮", "room_pic_108_4", "fp_04", 1000000, 500000, 0, []int64{1000, 2000, 5000, 10000}),
	NewRoom("Trận đặc biệt", "Du thuyền Caribe", "room_pic_108_4", "fp_04", 1000000, 500000, 0, []int64{1000, 2000, 5000, 10000}),
	//NewRoom("VIP场", "迪拜亚特兰蒂斯", "room_pic_108_5", "fp_06", 50000, 2000000, 3000000, []int64{1000, 10000, 100000, 500000}),
	NewRoom("Khu VIP", "Atlantis Dubai", "room_pic_108_5", "fp_06", 50000, 2000000, 3000000, []int64{1000, 10000, 100000, 500000}),
}

var defaultManager = NewManager()

type (
	Club struct {
		group  *nano.Group
		clubId int
		name   string
	}

	Manager struct {
		component.Base
		group      *nano.Group       // 广播channel
		players    map[int64]*Player // 所有的玩家
		chKick     chan int64        // 退出队列
		chReset    chan int64        // 重置队列
		chRecharge chan RechargeInfo // 充值信息
	}

	RechargeInfo struct {
		Uid  int64 // 用户ID
		Coin int64 // 房卡数量
	}
)

func NewBetZones() protocol.BetZones {
	var defaultBetZones = protocol.BetZones{
		Red: []protocol.BetItem{
			{Bg: "red", IconBg: "chair_red", Icon: "lion", Odds: 0, Total: 0, Name: "红狮"},
			{Bg: "red", IconBg: "chair_red", Icon: "panda", Odds: 0, Total: 0, Name: "红猫"},
			{Bg: "red", IconBg: "chair_red", Icon: "monkey", Odds: 0, Total: 0, Name: "红猴"},
			{Bg: "red", IconBg: "chair_red", Icon: "rabit", Odds: 0, Total: 0, Name: "红兔"},
			{Bg: "blue", IconBg: "", Icon: "zhuang", Odds: 0, Total: 0, Name: "庄"},
		},
		Green: []protocol.BetItem{
			{Bg: "green", IconBg: "chair_green", Icon: "lion", Odds: 0, Total: 0, Name: "绿狮"},
			{Bg: "green", IconBg: "chair_green", Icon: "panda", Odds: 0, Total: 0, Name: "绿猫"},
			{Bg: "green", IconBg: "chair_green", Icon: "monkey", Odds: 0, Total: 0, Name: "绿猴"},
			{Bg: "green", IconBg: "chair_green", Icon: "rabit", Odds: 0, Total: 0, Name: "绿兔"},
			{Bg: "blue", IconBg: "", Icon: "he", Odds: 0, Total: 0, Name: "和"},
		},
		Yellow: []protocol.BetItem{
			{Bg: "yellow", IconBg: "chair_yellow", Icon: "lion", Odds: 0, Total: 0, Name: "黄狮"},
			{Bg: "yellow", IconBg: "chair_yellow", Icon: "panda", Odds: 0, Total: 0, Name: "黄猫"},
			{Bg: "yellow", IconBg: "chair_yellow", Icon: "monkey", Odds: 0, Total: 0, Name: "黄猴"},
			{Bg: "yellow", IconBg: "chair_yellow", Icon: "rabit", Odds: 0, Total: 0, Name: "黄兔"},
			{Bg: "blue", IconBg: "", Icon: "xian", Odds: 0, Total: 0, Name: "闲"},
		},
	}
	return defaultBetZones
}

// business func

func (m *Manager) JoinSeat(s *session.Session, req *protocol.JoinSeatRequest) error {
	if p, ok := m.player(s.UID()); ok {
		p.joinSeat(req.SeatIndex)

		p.room.seatsUpdate()
	}
	return s.Response(nil)
}

func (m *Manager) RoomStatus(s *session.Session, msg []byte) error {
	if p, ok := m.player(s.UID()); ok {
		gamecoin := db.GetGameCoinByUid(s.UID())
		room := p.room
		room.seatsUpdate()
		return s.Response(&protocol.RoomStatus{
			GameStatus:  room.gameStatus,
			BetZones:    room.betZones,
			Deadline:    room.deadline,
			Bg:          room.bg,
			GameCoin:    gamecoin,
			RandAnimals: room.randAnimals,
			RandLights:  room.randLights,
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
		if req.Channel == 2 {
			if p.club != nil {
				log.Infof("club: %d", p.club.clubId)
				p.club.group.Broadcast("chat", broadcastRes)
			} else {
				return s.Response(&protocol.SendMessageErrorResponse{
					Error: "Not in club.",
				})
			}
		}
	} else {
		return s.Response(&protocol.SendMessageErrorResponse{
			Error: "Session has expired.",
		})
	}
	return s.Response(nil)
}

func (m *Manager) ClearBet(s *session.Session, msg []byte) error {
	if p, ok := m.player(s.UID()); ok { // get session player
		if p.room.gameStatus == "betting" { // only in betting status then can clear bet!
			p.clearAllBets()
			gamecoin := db.GetGameCoinByUid(p.uid) // get current gamecoin
			return s.Response(&protocol.ClearBetResponse{BetZones: p.room.betZones, GameCoin: gamecoin})
		}
	} else {
		return s.Response(&protocol.SendMessageErrorResponse{Error: "Session has expired."})
	}
	return s.Response(nil)
}

func (m *Manager) PlaceBet(s *session.Session, req *protocol.PlaceBetRequest) error {
	if p, ok := m.player(s.UID()); ok { // get session player
		if p.room.gameStatus == "betting" { // only in betting status then can place bet!
			// req : zone, key,  amount
			// if req.Amount+p.getAllBetsTotal() > p.room.maxGameCoin {
			// 	return s.Response(&protocol.SendMessageErrorResponse{Error: "当局最多押注" +
			// 		strconv.FormatInt(p.room.maxGameCoin, 10)})
			// }
			gamecoin := db.GetGameCoinByUid(p.uid) // get current gamecoin first
			// log.Infof("** player current gamecoin: %d", gamecoin)
			if gamecoin >= req.Amount+p.getAllBetsTotal() {
				zone := p.room.betZones.Red
				if req.Zone == "Green" {
					zone = p.room.betZones.Green
				} else if req.Zone == "Yellow" {
					zone = p.room.betZones.Yellow
				}
				zone[req.Key].Total += req.Amount
				if ok := p.placeBet(req.Zone, req.Key, req.Amount); ok {
					// log.Infof("NEWTOTAL: %d", zone[req.Key].Total)
					p.room.group.Broadcast("updateZone", &protocol.PlaceBetResponse{
						Zone:  req.Zone,
						Key:   req.Key,
						Total: zone[req.Key].Total,
						MyBet: p.getBet(req.Zone, req.Key),
						Uid:   p.uid,
					})
				} else {
					// return s.Response(&protocol.SendMessageErrorResponse{Error: "当局最多押注" +
					// 	strconv.FormatInt(p.room.maxGameCoin, 10)})
					return s.Response(&protocol.SendMessageErrorResponse{Error: "Lên đến" +
						strconv.FormatInt(p.room.maxGameCoin, 10) + "cược"})
				}
			} else {
				//return s.Response(&protocol.SendMessageErrorResponse{Error: "游戏币不足"})
				return s.Response(&protocol.SendMessageErrorResponse{Error: "Không đủ tiền"})
			}
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
			//Error: "游戏币不足",
			Error: "Không đủ tiền",
		})
	}

	if p, ok := m.player(uid); !ok {
		log.Infof("玩家: %d不在线，创建新的玩家", uid)

		// get player club
		club_user := db.GetClubByUid(uid)
		var club *Club = nil
		if club_user != nil {
			club_id, err := strconv.Atoi(club_user["club_id"])
			if err != nil {
				log.Println(err)
			}
			club = m.club(club_id)
			club.group.Add(s)
		}

		p = newPlayer(s, uid, req.Name, req.FaceUri, rooms[req.RoomNumber], club, req.Level)
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
	// log.Infof("betZones", rooms[req.RoomNumber].betZones)
	log.Infof("number of 玩家: %d", rooms[req.RoomNumber].sessionCount())
	//if rooms[req.RoomNumber].gameStatus == "nogame" && rooms[req.RoomNumber].sessionCount() > 0 {
	//	go rooms[req.RoomNumber].bettingPhase()
	//}

	return s.Response(
		&protocol.JoinRoomResponse{
			Name:      req.Name,
			RoomLevel: rooms[req.RoomNumber].level,
			FaceUri:   req.FaceUri,
			Chip:      rooms[req.RoomNumber].chip,
			BetZones:  rooms[req.RoomNumber].betZones,
		})
}

func (m *Manager) GetRooms(s *session.Session, msg []byte) error {
	availRooms := make([]protocol.RoomItem, len(rooms))
	for i := range rooms {
		availRooms[i] = protocol.RoomItem{
			Level:       rooms[i].level,
			Name:        rooms[i].name,
			MinGameCoin: rooms[i].minGameCoin,
			Chip:        rooms[i].chip,
			Icon:        rooms[i].icon,
		}
	}

	return s.Response(&protocol.GetRoomResponse{Rooms: availRooms})
}

func NewManager() *Manager {
	return &Manager{
		group:      nano.NewGroup("_SYSTEM_MESSAGE_BROADCAST"),
		players:    map[int64]*Player{},
		chKick:     make(chan int64, kickResetBacklog),
		chReset:    make(chan int64, kickResetBacklog),
		chRecharge: make(chan RechargeInfo, 32),
	}
}

func (m *Manager) AfterInit() {
	session.Lifetime.OnClosed(func(s *session.Session) {
		uid := s.UID()
		if p, ok := m.player(uid); ok {
			p.leaveSeat()
			p.clearAllBets()
			room := p.room
			room.offline(uid)
			room.group.Leave(s)
			if p.club != nil {
				p.club.group.Leave(s)
			}

		}
		m.offline(uid)
		m.group.Leave(s)
	})

	for _, room := range rooms {
		go room.bettingPhase()
	}
}

func (m *Manager) club(club_id int) *Club {
	// check if in clubs
	for i := range clubs {
		if clubs[i].clubId == club_id { // if yes, return existing club to player
			return clubs[i]
		}
	}
	// else create newclub return to player
	newClub := &Club{
		group:  nano.NewGroup(strconv.Itoa(club_id)),
		clubId: club_id,
	}
	clubs = append(clubs, newClub)
	return newClub
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

func NewRoom(level, name, icon, bg string, minGameCoin, maxGameCoin, minSeatCoin int64, chip []int64) *Room {
	numRobotPlayers := 10
	var robotPlayers []*protocol.RobotPlayer
	for i := 0; i < numRobotPlayers; i++ {
		robotPlayers = append(robotPlayers, NewRobotPlayer())
	}

	return &Room{
		level:                  level,
		name:                   name,
		minGameCoin:            minGameCoin,
		maxGameCoin:            maxGameCoin,
		minSeatCoin:            minSeatCoin,
		chip:                   chip,
		gameStatus:             "nogame",
		icon:                   icon,
		bg:                     bg,
		group:                  nano.NewGroup(level),
		betZones:               NewBetZones(),
		players:                map[int64]*Player{},
		seats:                  [6]*Player{},
		collection_lastupdated: time.Now().UTC(),
		collection_amount:      0,
		collection_betting:     0,
		robotPlayers:           robotPlayers,
		maxRobotBetting:        chip[len(chip)-2] * maxRobotBettingOffset,
		LogInformations:        []map[string]string{},
		randColors:             []string{"Red", "Green", "Yellow"},
	}
}
