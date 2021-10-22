package game

import (
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"time"

	"gitlab.com/wolfplus/gamespace-lhd/db"
	"gitlab.com/wolfplus/gamespace-lhd/protocol"

	"github.com/lonng/nano"
	"github.com/lonng/nano/component"
	"github.com/lonng/nano/session"
	log "github.com/sirupsen/logrus"
)

const kickResetBacklog = 8

var defaultManager = NewManager()

// robot options
var robotNames = []string{"Minh Huệ", "Ngọc Thanh", "Lý Mỹ Kỳ", "Hồ Vĩnh Khoa", "Nguyễn Kim Hồng",
	"Phạm Gia Chi Bảo", "Ngoc Trinh", "Nguyễn Hoàng Bích", "Đặng Thu Thảo", "Nguyen Thanh Tung"}

var RoomRobotPlayers = []*RobotPlayer{newRobotPlayer(), newRobotPlayer(), newRobotPlayer(),
	newRobotPlayer(), newRobotPlayer(), newRobotPlayer(), newRobotPlayer(),
	newRobotPlayer(), newRobotPlayer(), newRobotPlayer()}

type (
	Manager struct {
		component.Base
		group                  *nano.Group       // 广播channel
		players                map[int64]*Player // 所有的玩家
		chKick                 chan int64        // 退出队列
		chReset                chan int64        // 重置队列
		gameStatus             string
		deadline               time.Time // when to go to the next phase
		betZones               []*protocol.BetZone
		winningItems           []*protocol.WinningItem
		selectedWinningItems   []*protocol.WinningItem
		selectedWinningItem    *protocol.WinningItem
		selectedWinningItemKey int
		selectedWinningItemOdd int // for logging
		historyList            []*protocol.WinningItem
		winners                []protocol.Winner
		totalBetting           int64
		specialPrize           string
		specialPrizeChinese    string // for logging
		totalWinLose           int64
		logInformations        []map[string]string
		sync.RWMutex
	}
)

// game phases cycle function
func (m *Manager) resultphase() {
	// defensive programming
	if m.gameStatus != "animation" {
		return
	}

	// 结算界面显示6S
	m.deadline = time.Now().UTC().Add(time.Second * time.Duration(6))

	// update game status
	m.gameStatus = "result"
	// var singleSelectedWinningItem *protocol.WinningItem
	// if m.specialPrize == "bigFruit" {
	// 	singleSelectedWinningItem = &protocol.WinningItem{WinningBetZoneIndex: -1,
	// 		Image: "history_special_1"}
	// } else if m.specialPrize == "train" {
	//	singleSelectedWinningItem = &protocol.WinningItem{WinningBetZoneIndex: -1,
	//		Image: "history_special_2"}
	// } else if m.specialPrize == "fairy" {
	// 	singleSelectedWinningItem = &protocol.WinningItem{WinningBetZoneIndex: -1,
	// 		Image: "history_special_3"}
	// } else {
	// 	singleSelectedWinningItem = m.selectedWinningItem
	// }

	var wg sync.WaitGroup
	var x = 0
	players := m.getPlayers()

	for uid := range players {
		wg.Add(1)
		go func(uid int64, x int) {
			updateRobotWinners := false
			if x == 0 {
				updateRobotWinners = true
			}
			if p, ok := m.players[uid]; ok {
				gamecoin := db.GetGameCoinByUid(uid) // get current gamecoin
				if p.session != nil {
					p.session.Push("resultPhase", &protocol.ResultPhaseResponse{
						Deadline:             m.deadline,
						PersonalBetResult:    p.betResult,
						Awarded:              p.awarded,
						GameCoin:             gamecoin,
						Winners:              m.winners,
						SelectedWinningItem:  m.selectedWinningItem,
						SelectedWinningItems: m.selectedWinningItems,
						UpdateRobotWinners:   updateRobotWinners,
						WinningKey:           m.selectedWinningItemKey,
						SpecialPrize:         m.specialPrize,
					})
				}
			}
			wg.Done()
		}(uid, x)
		x++
	}

	wg.Wait()
	m.resetForBetting() //reset robot betting

	// count down to next game status
	//if m.sessionCount() == 0 {
	//	m.gameStatus = "nogame"
	//} else {
	// go func() {
	s := m.deadline.Sub(time.Now()).Seconds()
	time.Sleep(time.Duration(s) * time.Second)
	m.bettingPhase()
	// }()
	//}
}

func (m *Manager) animationPhase() {
	// defensive programming
	if m.gameStatus != "betting" {
		return
	}

	// update game status
	m.gameStatus = "animation"
	m.totalWinLose = 0                                 // reset
	m.specialPrizeChinese = ""                         // reset wk add "普通"
	m.selectedWinningItems = []*protocol.WinningItem{} // reset

	var totalBetting int64 = 0
	for i := range m.betZones {
		totalBetting += m.betZones[i].Total
	}
	m.totalBetting = totalBetting

	//if totalBetting > 100000 {
	if randomRange(1, 100) <= 10 { // special
		if randomRange(1, 100) <= 40 {
			m.resultPasskill()
			return
		} else { // goodluck
			m.deadline = time.Now().UTC().Add(time.Second * time.Duration(25))

			m.resultSpecial()

			players := m.getPlayers()
			var wg sync.WaitGroup

			for _, p := range players {
				wg.Add(1) // counter ++
				go func(p *Player) {
					p.calcPlayerSpecialResult()
					wg.Done()
				}(p)
			}

			for i := range RoomRobotPlayers {
				wg.Add(1)
				go func(rp *RobotPlayer) {
					m.calcRobotSpecialResult(rp)
					wg.Done()
				}(RoomRobotPlayers[i])
			}

			wg.Wait()

			// END goodluck
		}
	} else { // 94%
		m.resultCommon()
	}
	//} else { // less than 1000 万
	//	m.resultCommon()
	//}

	db.RecordWinningItem(m.selectedWinningItemKey) //insert to db

	// sort winner
	var awards = []int{}
	keys := map[int]int64{}

	players := m.getPlayers()

	for k, p := range players {
		keys[int(p.awarded)] = k
		awards = append(awards, int(p.awarded))
	}
	for i, rp := range RoomRobotPlayers { // robot players added to sort
		keys[int(rp.awarded)] = int64(i) // this is not a real uid
		awards = append(awards, int(rp.awarded))
	}
	sort.Sort(sort.Reverse(sort.IntSlice(awards))) // big to small

	m.setGameSystemLoggingInfo() // logging to db

	// populate winners with the sorted result
	m.winners = []protocol.Winner{}
	for i := range awards {
		key := keys[awards[i]] // uid
		if p, ok := m.players[key]; ok && awards[i] > 0 {
			m.winners = append(m.winners, protocol.Winner{
				Username: p.userName,
				Image:    p.faceUri,
				Level:    p.level,
				Awarded:  awards[i],
				IsRobot:  false,
			})
		} else if awards[i] > 0 {
			m.winners = append(m.winners, protocol.Winner{
				Username: RoomRobotPlayers[key].userName,
				Image:    strconv.Itoa(RoomRobotPlayers[key].faceUri), //random image
				Level:    RoomRobotPlayers[key].level,                 // random level
				Awarded:  awards[i],
				IsRobot:  true,
			})
		}
	}

	s := m.deadline.Sub(time.Now()).Seconds() + 1
	time.Sleep(time.Duration(s) * time.Second)
	m.resultphase()
}

func (m *Manager) bettingPhase() {
	// how long this phase will last i.e. 30 seconds
	m.deadline = time.Now().UTC().Add(time.Second * time.Duration(30))
	// m.deadline = time.Now().UTC().Add(time.Second * time.Duration(5))

	// update game status
	m.gameStatus = "betting"

	// set up betzone for betting
	for i := range m.betZones {
		m.betZones[i].Total = 0
	}
	m.specialPrize = "" //resetSpecial

	// tell to the slave when to expect next phase
	m.group.Broadcast("bettingPhase", &protocol.BettingPhaseResponse{
		Deadline: m.deadline,
	})

	var wg sync.WaitGroup
	for i := range RoomRobotPlayers {
		var betTimes = randomRange(3, 10)
		for x := 0; x < betTimes; x++ {
			wg.Add(1)
			go func(robotPlayer *RobotPlayer) {
				s := m.deadline.Sub(time.Now()).Seconds() + 1
				n := randomRange(3000, int(s)*1000)
				time.Sleep(time.Duration(n) * time.Millisecond)
				m.robotPlayerPlaceBet(robotPlayer)
				wg.Done()
			}(RoomRobotPlayers[i])
		}
	}
	wg.Wait()

	// count down to next game status
	// go func() {
	s := m.deadline.Sub(time.Now()).Seconds() + 1
	time.Sleep(time.Duration(s) * time.Second)
	m.animationPhase()
	// }()
}

// business func ( mostly should be a handler)
func (m *Manager) PlaceBet(s *session.Session, req *protocol.PlaceBetRequest) error {
	if p, ok := m.player(s.UID()); ok { // get session player
		if m.gameStatus == "betting" {

			if m.betZones[req.BetZoneKey].Total > 2500000 { // 1000w
				//return s.Response(&protocol.SendMessageErrorResponse{Error: "图案下注上限是9999万"})
				return s.Response(&protocol.SendMessageErrorResponse{Error: "Lên đến 250 triệu cược"})
			}

			gamecoin := db.GetGameCoinByUid(p.uid)
			if gamecoin >= req.PlaceBetAmount+p.getAllBetsTotal() {
				m.Lock()
				m.betZones[req.BetZoneKey].Total += req.PlaceBetAmount
				m.Unlock()
				p.placeBet(req.BetZoneKey, req.PlaceBetAmount)

				m.RLock()
				m.group.Broadcast("updateZone", &protocol.PlaceBetResponse{
					BetZoneKey: req.BetZoneKey,
					TotalBet:   m.betZones[req.BetZoneKey].Total,
					MyBet:      p.getBet(req.BetZoneKey),
					Uid:        p.uid,
				})
				m.RUnlock()
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

func (m *Manager) ClearBet(s *session.Session, msg []byte) error {
	if p, ok := m.player(s.UID()); ok { // get session player
		if m.gameStatus == "betting" { // only in betting status then can clear bet!
			p.clearAllBets()
			gamecoin := db.GetGameCoinByUid(p.uid) // get current gamecoin
			return s.Response(&protocol.ClearBetResponse{
				BetZones: m.betZones,
				GameCoin: gamecoin})
		}
	} else {
		return s.Response(&protocol.SendMessageErrorResponse{Error: "Session has expired."})
	}
	return s.Response(nil)
}

func (m *Manager) RoomStatus(s *session.Session, msg []byte) error {
	if _, ok := m.player(s.UID()); ok {

		chip := []int{10, 100, 1000, 10000, 100000, 1000000}
		levelToChat := 5
		gamecoin := db.GetGameCoinByUid(s.UID())
		// lastWinningItem := m.selectedWinningItem
		// if lastWinningItem == nil {
		// 	lastWinningItem = m.winningItems[0]
		// }

		log.Infof("gameStatus: %v", m.gameStatus)

		m.RLock()
		defer m.RUnlock()
		return s.Response(&protocol.RoomStatusResponse{
			GameStatus:         m.gameStatus,
			Deadline:           m.deadline,
			BetZones:           m.betZones,
			Chips:              chip,
			HistoryList:        m.historyList,
			WinningItem:        m.winningItems,
			LevelToChat:        levelToChat,
			GameCoin:           gamecoin,
			LastWinningItemKey: m.selectedWinningItemKey,
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

	// get current gamecoin
	// gamecoin := db.GetGameCoinByUid(uid)

	if p, ok := m.player(uid); !ok {
		log.Infof("玩家: %d不在线，创建新的玩家", uid)

		p = newPlayer(s, uid, req.Name, req.FaceUri, req.Level, m)
		m.setPlayer(uid, p) // set player basic info i.e. connect a player to this game

	} else {
		log.Infof("玩家: %d已经在线", uid)
		m.group.Leave(s)

		// 重置之前的session
		if prevSession := p.session; prevSession != nil && prevSession != s {
			prevSession.Clear()
			prevSession.Close()
		}

		// 绑定新session
		p.bindSession(s)
	}
	m.group.Add(s)

	// if m.gameStatus == "nogame" && m.sessionCount() > 0 {
	//	go m.bettingPhase()
	// }

	return s.Response(&protocol.JoinRoomResponse{
		Name:     req.Name,
		FaceUri:  req.FaceUri,
		BetZones: m.betZones,
	})
}

func (m *Manager) getPlayers() map[int64]*Player {
	m.RLock()
	defer m.RUnlock()
	return m.players
}

// get player basic info
func (m *Manager) player(uid int64) (*Player, bool) {
	p, ok := m.players[uid]

	return p, ok
}

// init player
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

// helpers

func (m *Manager) calculateWinLoseCoin() {
	m.selectedWinningItems = []*protocol.WinningItem{m.selectedWinningItem}

	// calculate player win/lose amount
	var wg sync.WaitGroup
	players := m.getPlayers()

	for _, p := range players {
		wg.Add(1) // counter ++
		go func(p *Player) {
			p.calcPlayerResult()
			wg.Done()
		}(p)
	}

	wg.Wait()

	for i := range RoomRobotPlayers {
		wg.Add(1)
		go func(rp *RobotPlayer) {
			m.robotPlayerCalcResult(rp)
			wg.Done()
		}(RoomRobotPlayers[i])
	}

	wg.Wait()
}

func (m *Manager) setGameSystemLoggingInfo() {
	// //开奖：[宝马｜5] [总押注：35476000， 总输赢：35421000]
	// if len(m.selectedWinningItems) != 0 {
	winningItemLogInfo := m.getWinningItemLoggingInfo()
	// }
	// logger.Infoln("selectedWinningItemKey", m.selectedWinningItemKey)
	if m.specialPrizeChinese != "" {
		winningItemLogInfo = m.specialPrizeChinese + "：" + winningItemLogInfo
	}

	otherInfo := "开奖：[" + winningItemLogInfo + "]" + "[总押注：" + strconv.Itoa(int(m.totalBetting)) +
		"，总输赢：" + strconv.Itoa(int(-m.totalWinLose)) + "]"
	logInformation := map[string]string{
		"uid":            "0",
		"game":           "水果",
		"level":          "",
		"otherInfo":      otherInfo, //开奖：[大众｜5] [总押注：40837000， 总输赢：30832000]
		"result":         strconv.Itoa(m.selectedWinningItemKey),
		"rate":           strconv.Itoa(m.selectedWinningItemOdd),
		"betTotal":       strconv.Itoa(int(m.totalBetting)),
		"winTotal":       strconv.Itoa(int(-m.totalWinLose)),
		"bankerWinTotal": strconv.Itoa(int(-m.totalWinLose)),
	}
	// log.Infoln("logInformation", logInformation)
	m.InsertAllLogInformations(logInformation, winningItemLogInfo)
}

func (m *Manager) getWinningItemLoggingInfo() string {
	winningItems := ""
	if len(m.selectedWinningItems) == 0 || // common
		len(m.selectedWinningItems) > 0 && m.selectedWinningItems[0].WinningBetZoneIndex == -1 {
		// passkill
		return winningItems
	}
	for i, k := range m.selectedWinningItems {
		winningItemName := m.betZones[k.WinningBetZoneIndex].Name
		winningItems += "" + winningItemName + "｜" + strconv.Itoa(int(k.Odds))
		if (i + 1) != len(m.selectedWinningItems) {
			winningItems += "，"
		}
	}
	// logger.Infoln("winningItems", winningItems)
	return winningItems
}

func NewManager() *Manager {
	betZones := []*protocol.BetZone{
		{Total: 0, IsFruit: false, Name: "BAR"}, // 0- bar
		{Total: 0, IsFruit: false, Name: "双7"},  // 1- 77
		{Total: 0, IsFruit: false, Name: "双星"},  // 2- 双星
		{Total: 0, IsFruit: true, Name: "西瓜"},   // 3- 西瓜
		{Total: 0, IsFruit: false, Name: "铃铛"},  // 4- 铃铛
		{Total: 0, IsFruit: true, Name: "木瓜"},   // 5- 木瓜
		{Total: 0, IsFruit: true, Name: "橘子"},   // 6- 橘子
		{Total: 0, IsFruit: true, Name: "苹果"},   // 7- 苹果
	}
	winningItems := []*protocol.WinningItem{
		{WinningBetZoneIndex: 6, Probability: 300, Odds: 10, Image: "reel_slot_0", IsBigFruit: true, Music: "橘子"},
		{WinningBetZoneIndex: 4, Probability: 150, Odds: 20, Image: "reel_slot_1", IsBigFruit: true, Music: "铃铛"},
		{WinningBetZoneIndex: 0, Probability: 50, Odds: 60, Image: "reel_slot_2", IsBigFruit: false, Music: "BAR"},
		{WinningBetZoneIndex: 0, Probability: 20, Odds: 120, Image: "reel_slot_3", IsBigFruit: false, Music: "BAR"},
		{WinningBetZoneIndex: 7, Probability: 205, Odds: 5, Image: "reel_slot_4", IsBigFruit: true, Music: "苹果"},
		{WinningBetZoneIndex: 7, Probability: 1100, Odds: 2, Image: "reel_slot_5", IsBigFruit: false, Music: "苹果"},
		{WinningBetZoneIndex: 5, Probability: 200, Odds: 15, Image: "reel_slot_6", IsBigFruit: true, Music: "木瓜"},
		{WinningBetZoneIndex: 3, Probability: 300, Odds: 20, Image: "reel_slot_7", IsBigFruit: true, Music: "西瓜"},
		{WinningBetZoneIndex: 3, Probability: 960, Odds: 2, Image: "reel_slot_8", IsBigFruit: false, Music: "西瓜"},
		{WinningBetZoneIndex: -1, Probability: 0, Odds: 1, Image: "reel_slot_9", IsBigFruit: false, Music: "LUCK"},
		{WinningBetZoneIndex: 7, Probability: 205, Odds: 5, Image: "reel_slot_10", IsBigFruit: true, Music: "苹果"},
		{WinningBetZoneIndex: 6, Probability: 1100, Odds: 2, Image: "reel_slot_11", IsBigFruit: false, Music: "橘子"},
		{WinningBetZoneIndex: 6, Probability: 300, Odds: 10, Image: "reel_slot_12", IsBigFruit: true, Music: "橘子"},
		{WinningBetZoneIndex: 4, Probability: 150, Odds: 20, Image: "reel_slot_13", IsBigFruit: true, Music: "铃铛"},
		{WinningBetZoneIndex: 1, Probability: 940, Odds: 2, Image: "reel_slot_14", IsBigFruit: false, Music: "双7"},
		{WinningBetZoneIndex: 1, Probability: 140, Odds: 40, Image: "reel_slot_15", IsBigFruit: true, Music: "双7"},
		{WinningBetZoneIndex: 7, Probability: 205, Odds: 5, Image: "reel_slot_16", IsBigFruit: true, Music: "苹果"},
		{WinningBetZoneIndex: 5, Probability: 1020, Odds: 2, Image: "reel_slot_17", IsBigFruit: false, Music: "木瓜"},
		{WinningBetZoneIndex: 5, Probability: 200, Odds: 15, Image: "reel_slot_18", IsBigFruit: true, Music: "木瓜"},
		{WinningBetZoneIndex: 2, Probability: 200, Odds: 30, Image: "reel_slot_19", IsBigFruit: true, Music: "双星"},
		{WinningBetZoneIndex: 2, Probability: 1100, Odds: 2, Image: "reel_slot_20", IsBigFruit: false, Music: "双星"},
		{WinningBetZoneIndex: -1, Probability: 0, Odds: 20, Image: "reel_slot_21", IsBigFruit: false, Music: "LUCK"},
		{WinningBetZoneIndex: 7, Probability: 205, Odds: 5, Image: "reel_slot_22", IsBigFruit: true, Music: "苹果"},
		{WinningBetZoneIndex: 4, Probability: 950, Odds: 2, Image: "reel_slot_23", IsBigFruit: false, Music: "铃铛"},
	}

	return &Manager{
		group:                  nano.NewGroup("Manager"),
		players:                map[int64]*Player{},
		chKick:                 make(chan int64, kickResetBacklog),
		chReset:                make(chan int64, kickResetBacklog),
		gameStatus:             "nogame",
		betZones:               betZones,
		winningItems:           winningItems,
		selectedWinningItemKey: 0,
	}
}

func (m *Manager) AfterInit() {
	session.Lifetime.OnClosed(func(s *session.Session) {
		uid := s.UID()
		m.offline(uid)
		m.group.Leave(s)
	})

	go m.bettingPhase()
}

func randomRange(min int, max int) int {
	// return rand.Intn(max-min) + min
	return min + rand.Intn(max-min+1)
}
