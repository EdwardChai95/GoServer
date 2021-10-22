package game

import (
	"sync"
	"time"

	"gitlab.com/wolfplus/gamespace-lhd/protocol"

	"github.com/lonng/nano"
	"github.com/lonng/nano/component"
	"github.com/lonng/nano/session"
	log "github.com/sirupsen/logrus"

	"gitlab.com/wolfplus/gamespace-lhd/db"
)

const kickResetBacklog = 8

var defaultManager = NewManager()

type (
	Manager struct {
		component.Base
		group      *nano.Group       // 广播channel
		players    map[int64]*Player // 所有的玩家
		chKick     chan int64        // 退出队列
		chReset    chan int64        // 重置队列
		chRecharge chan RechargeInfo // 充值信息

		//fields from room struct
		level          string
		name           string
		minGameCoin    int64
		chip           []int64
		gameStatus     string // facilitate game
		icon           string
		winningItems   []*protocol.BetItem
		betZones       []*protocol.BetItem
		deadline       time.Time
		winningBetZone *protocol.BetItem
		betItems       map[float64]CurrentBetting
		winners        []protocol.Winner
		bankerImg      int
		bankerAmt      int
		historyList    []*protocol.BetItem // history
		sync.RWMutex
	}

	RechargeInfo struct {
		Uid  int64 // 用户ID
		Coin int64 // 房卡数量
	}
)

func NewBetZones() []*protocol.BetItem {
	return []*protocol.BetItem{
		{Odds: 40, Name: "保时捷", Size: "big", Image: "porsche", NickName: "保大"},
		{Odds: 30, Name: "宝马", Size: "big", Image: "bmw", NickName: "宝大"},
		{Odds: 20, Name: "奔驰", Size: "big", Image: "benz", NickName: "奔大"},
		{Odds: 10, Name: "大众", Size: "big", Image: "volks", NickName: "众大"},
		{Odds: 5, Name: "保时捷", Size: "small", Image: "porsche", NickName: "保小"},
		{Odds: 5, Name: "宝马", Size: "small", Image: "bmw", NickName: "宝小"},
		{Odds: 5, Name: "奔驰", Size: "small", Image: "benz", NickName: "奔小"},
		{Odds: 5, Name: "大众", Size: "small", Image: "volks", NickName: "众小"},
	}
}

// business func

func (m *Manager) RoomStatus(s *session.Session, msg []byte) error {
	if _, ok := m.player(s.UID()); ok {
		gamecoin := db.GetGameCoinByUid(s.UID())

		m.RLock()
		defer m.RUnlock()
		return s.Response(&protocol.RoomStatus{
			GameStatus:   m.gameStatus,
			BetZones:     m.betZones,
			Deadline:     m.deadline,
			GameCoin:     gamecoin,
			WinningItems: m.winningItems,
			BankerImg:    m.bankerImg,
			BankerAmt:    m.bankerAmt,
			HistoryList:  m.historyList,
		})
	} else {
		return s.Response(&protocol.SendMessageErrorResponse{
			Error: "Session has expired.",
		})
	}
	// return s.Response(nil)
}

func (m *Manager) ClearBet(s *session.Session, msg []byte) error {
	if p, ok := m.player(s.UID()); ok { // get session player
		if m.gameStatus == "betting" { // only in betting status then can clear bet!
			p.clearAllBets()
			gamecoin := db.GetGameCoinByUid(p.uid) // get current gamecoin
			return s.Response(&protocol.ClearBetResponse{BetZones: m.betZones, GameCoin: gamecoin})
		}
	} else {
		return s.Response(&protocol.SendMessageErrorResponse{Error: "Session has expired."})
	}
	return s.Response(nil)
}

func (m *Manager) totalExpectedWinnings(key int) int64 {
	var totalExpectedWinnings int64
	for _, p := range m.players {
		totalExpectedWinnings += p.expectedWinnings(key, 0)
	}
	for _, rp := range RoomRobotPlayers {
		rp.RLock()
		totalExpectedWinnings += rp.currentBettings[CurrentBetting{
			key: key,
		}] * int64(m.betZones[key].Odds)
		rp.RUnlock()
	}
	// now also includes robotPlayers
	return totalExpectedWinnings
}

func (m *Manager) PlaceBet(s *session.Session, req *protocol.PlaceBetRequest) error {
	if p, ok := m.player(s.UID()); ok { // get session player
		m.Lock()
		defer m.Unlock()
		if m.gameStatus == "betting" { // only in betting status then can place bet!
			// req : key,  amount
			// bet amt less than banker amt
			// if p.expectedWinnings(req.Key, req.Amount) >= int64(m.bankerAmt) {
			if m.totalExpectedWinnings(req.Key)+p.expectedWinnings(req.Key, req.Amount) >= int64(m.bankerAmt) {
				//return s.Response(&protocol.SendMessageErrorResponse{Error: "所押注金额不能大于庄家赔付最大金额"})
				return s.Response(&protocol.SendMessageErrorResponse{Error: "Nhà cái không có đủ tiền"})
			}
			// 20,000,000 max betting
			if req.Amount+p.getAllBetsTotal() > 2000000 {
				//return s.Response(&protocol.SendMessageErrorResponse{Error: "当局最多押注2000万"})
				return s.Response(&protocol.SendMessageErrorResponse{Error: "Lên đến 20 triệu cược"})
			}
			gamecoin := db.GetGameCoinByUid(p.uid) // get current gamecoin first
			// // log.Infof("** player current gamecoin: %d", gamecoin)
			if gamecoin >= req.Amount+p.getAllBetsTotal() {
				m.betZones[req.Key].Total += req.Amount
				p.placeBet(req.Key, req.Amount)
				// log.Infof("NEWTOTAL: %d", zone[req.Key].Total)
				m.group.Broadcast("updateZone", &protocol.PlaceBetResponse{
					Key:   req.Key,
					Total: m.betZones[req.Key].Total,
					MyBet: p.getBet(req.Key),
					Uid:   p.uid,
				})
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
	log.Infof("玩家session: %+v", s)
	log.Infof("玩家: %d登录: %+v", uid, req)

	// gamecoin := db.GetGameCoinByUid(uid) // get current gamecoin

	//把玩家加入nano群组
	if p, ok := m.player(uid); !ok {
		log.Infof("玩家: %d不在线，创建新的玩家", uid)

		p = newPlayer(s, uid, req.Name, req.FaceUri, m)
		m.setPlayer(uid, p) // set player basic info
	} else {
		log.Infof("玩家: %d已经在线", uid)
		// 移除广播频道
		m.group.Leave(s)

		// 重置之前的session
		if prevSession := p.session; prevSession != nil && prevSession != s {
			prevSession.Clear()
			prevSession.Close()
		}

		// 绑定新session
		p.bindSession(s)
		p.bindManager(m)
	}
	logger.Infof("m.group: %v", m.group)
	m.group.Add(s)

	log.Infof("number of 玩家: %d", m.sessionCount())

	//检查并进入下一阶段
	//
	// if m.gameStatus == "nogame" && m.sessionCount() > 0 {
	//	go m.bettingPhase()
	//}

	return s.Response(
		&protocol.JoinRoomResponse{
			Name:     req.Name,
			FaceUri:  req.FaceUri,
			BetZones: m.betZones,
		})
}

func NewManager() *Manager {
	// add winning item here
	// declear betzone here
	betzones := NewBetZones()

	winningItems := []*protocol.BetItem{
		// 众大 3，众小 7，保大 0 ，保小 4，宝大 1，宝小 5，奔大 2，奔小 6
		betzones[3],
		betzones[7],
		betzones[0],
		betzones[4],
		betzones[1],
		betzones[5],
		betzones[2],
		betzones[6],
		//
		betzones[3],
		betzones[7],
		betzones[0],
		betzones[4],
		betzones[1],
		betzones[5],
		betzones[2],
		betzones[6],
		//
		betzones[3],
		betzones[7],
		betzones[0],
		betzones[4],
		betzones[1],
		betzones[5],
		betzones[2],
		betzones[6],
		//
		betzones[3],
		betzones[7],
		betzones[0],
		betzones[4],
		betzones[1],
		betzones[5],
		betzones[2],
		betzones[6],
	}

	return &Manager{
		group:        nano.NewGroup("HAOCHEHUI"),
		players:      map[int64]*Player{},
		chKick:       make(chan int64, kickResetBacklog),
		chReset:      make(chan int64, kickResetBacklog),
		chRecharge:   make(chan RechargeInfo, 32),
		gameStatus:   "nogame",
		betZones:     betzones,
		winningItems: winningItems,
	}
}

func (m *Manager) AfterInit() {
	session.Lifetime.OnClosed(func(s *session.Session) {

		uid := s.UID()
		if p, ok := m.player(uid); ok {
			manager := p.manager
			manager.offline(uid)
			manager.group.Leave(s)

		}
		m.offline(uid)
		m.group.Leave(s)
	})

	go m.bettingPhase()
}

func (m *Manager) getPlayers() map[int64]*Player {
	m.RLock()
	defer m.RUnlock()
	return m.players
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
