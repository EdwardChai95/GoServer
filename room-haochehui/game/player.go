package game

import (
	"strconv"
	"sync"

	"github.com/lonng/nano/session"
	log "github.com/sirupsen/logrus"
	"gitlab.com/wolfplus/gamespace-lhd/db"
	"gitlab.com/wolfplus/gamespace-lhd/protocol"
)

type (
	CurrentBetting struct {
		key int
	}

	Player struct {
		uid             int64  // 用户ID
		faceUri         string // 头像地址
		userName        string // 玩家名字
		session         *session.Session
		logger          *log.Entry // 日志
		currentBettings map[CurrentBetting]int64
		betResult       protocol.BetResult
		awarded         int64
		// bonus           int64
		manager *Manager
		sync.Mutex
	}
)

func (p *Player) expectedWinnings(key int, amount int64) int64 {
	m := p.manager
	var expectedWinnings int64
	// for k := range p.currentBettings {
	// 	expectedWinnings += p.currentBettings[k] * int64(m.betZones[k.key].Odds)
	// }
	// return expectedWinnings
	if amount > 0 {
		expectedWinnings += amount * int64(m.betZones[key].Odds)
	} else {
		expectedWinnings += p.getBet(key) * int64(m.betZones[key].Odds)
	}

	return expectedWinnings
}

func (p *Player) calcPlayerResult(winningBetZone *protocol.BetItem) (int64, int64, int64, int64) {
	m := p.manager
	// log.Infof("player: %d", p)
	var awarded int64 // this player's total raw winning
	// var returnAmt int64    // this player's betting that is not forfeitted
	var totalbetting int64 // this player's total gamecoin spend into betting this round
	// log.Infof("PLAYER CURRENT BETTINGS: %v", p.currentBettings)
	var bankerwinning int64 // bankerwinning of this player

	bet := "[押注：" // for logging
	i := 0

	for k := range p.currentBettings {
		i++
		bet += "" + m.betZones[k.key].Name + "｜" + strconv.Itoa(int(p.currentBettings[k]))
		if i != len(p.currentBettings) {
			bet += "，"
		}
		if m.betZones[k.key] == m.winningBetZone { // player bet on the right bet zone
			reward := p.currentBettings[k] * int64(winningBetZone.Odds) // reward = bet amount x odds
			p.betResult = protocol.BetResult{
				Key:       k.key,
				Odds:      winningBetZone.Odds,
				PlacedBet: p.currentBettings[k],
				Reward:    reward,
			}
			awarded += reward // add to your current winning
			bankerwinning -= (reward - p.currentBettings[k])
		} else {
			bankerwinning += p.currentBettings[k]
		}
		totalbetting += p.currentBettings[k]

	}
	bet += "]"

	var playerWinnings int64 = 0 // amount player win or lose
	var playerTax int64 = 0      // amount that player deduct if they win

	playerWinnings = awarded - totalbetting

	// log.Infof("awarded : %d", awarded)
	// log.Infof("returnAmt : %d", returnAmt)
	// log.Infof("awarded + returnAmt - totalbetting : %d", playerWinnings)
	if playerWinnings > 0 {
		playerTax = int64(float64(playerWinnings) * 0.05)
		playerWinnings -= playerTax
		// playerWinnings = int64(float64(awarded) * 0.95) - totalbetting
		// playerWinnings = awarded - totalbetting - int64(float64(awarded-totalbetting)*0.05)
		// playerTax = int64(float64(awarded) * 0.05)
	}
	// log.Infof("playerWinnings : %d", playerWinnings)
	// log.Infof("totalbetting : %d", totalbetting)

	p.awarded = playerWinnings //  winlose of this round

	if totalbetting > 0 {
		otherInforStr := ""
		otherInforStr += "[总押注：" + strconv.FormatInt(totalbetting, 10) + "，输赢：" + strconv.FormatInt(playerWinnings, 10) +
			", 税收：" + strconv.FormatInt(playerTax, 10) + "] "
		otherInforStr += bet

		game_coin := db.GetGameCoinByUid(p.Uid())

		logInformation := map[string]string{
			"uid":       strconv.FormatInt(p.Uid(), 10),
			"game":      "豪车汇",
			"betTotal":  strconv.FormatInt(totalbetting, 10),
			"result":    strconv.FormatInt(playerWinnings, 10), // player bet
			"level":     "",
			"otherInfo": otherInforStr,
			"before":    strconv.FormatInt(game_coin, 10),
			"used":      strconv.FormatInt(playerWinnings, 10),
			"after":     strconv.FormatInt(game_coin+playerWinnings, 10),
			"tax":       strconv.Itoa(int(playerTax)),
		}
		db.NewLogInformation(logInformation)
	}

	go db.UpdateGameCoinByUid(p.Uid(), playerWinnings) // update gamecoin of player in db
	//add 1007
	go db.UpdateWinGameCoinByUid(p.Uid(), playerWinnings) // update wingamecoin of player in db
	go db.NewGameCoinTransaction(p.Uid(), playerWinnings)

	return playerWinnings, playerTax, totalbetting, bankerwinning
}

func (p *Player) clearAllBets() {
	m := p.manager
	for currentBetting, currentBettingAmount := range p.currentBettings {
		m.betZones[currentBetting.key].Total -= currentBettingAmount
		m.group.Broadcast("updateZone", &protocol.PlaceBetResponse{
			Key:   currentBetting.key,
			Total: m.betZones[currentBetting.key].Total,
			MyBet: 0,
			Uid:   p.uid,
		})
	}
	p.currentBettings = map[CurrentBetting]int64{}
}

func (p *Player) getAllBetsTotal() int64 {
	var total int64 = 0
	for k := range p.currentBettings {
		total += p.currentBettings[k]
	}
	return total
}

func (p *Player) getBet(key int) int64 {
	return p.currentBettings[CurrentBetting{
		key: key,
	}]
}

func (p *Player) placeBet(key int, amount int64) {
	p.Lock()
	defer p.Unlock()
	betkey := CurrentBetting{
		key: key,
	}
	if val, ok := p.currentBettings[betkey]; ok {
		val += amount
		p.currentBettings[betkey] = val
	} else {
		p.currentBettings[betkey] = amount
	}
}

func newPlayer(s *session.Session, uid int64, name string, faceUri string, m *Manager) *Player {
	p := &Player{
		uid:             uid,
		userName:        name,
		faceUri:         faceUri,
		logger:          log.WithField("player", uid),
		currentBettings: map[CurrentBetting]int64{},
		betResult:       protocol.BetResult{},
		manager:         m,
	}

	p.bindSession(s)

	return p
}

func (p *Player) bindManager(m *Manager) {
	p.manager = m
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
