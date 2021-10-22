package game

import (
	"strconv"

	// "github.com/ethereum/go-ethereum/log"
	"github.com/lonng/nano/session"
	"gitlab.com/wolfplus/gamespace-lhd/db"
	"gitlab.com/wolfplus/gamespace-lhd/protocol"
)

type (
	Player struct {
		uid             int64  // 用户ID
		faceUri         string // 头像地址
		level           int
		userName        string // 玩家名字
		session         *session.Session
		currentBettings map[int]int64 // [gambler key] amount
		awarded         int64
		CurrentWinnings map[int]int64 // [gambler key] amount
		betResult       protocol.BetResult
		manager         *Manager
	}
)

// business func
func (p *Player) calcPlayerSpecialResult() {
	var awarded int64 // raw winnings
	var totalbetting int64
	var winningBet int64
	var playerWinnings int64 = 0 // amount player win or lose
	// check if player betting inside selectedWinningItems
	bet := "[押注：" // for logging
	i := 0
	for k := range p.currentBettings {
		// isWin := false
		i++

		playerBetZoneName := p.manager.betZones[k].Name
		bet += "" + playerBetZoneName + "｜" + strconv.Itoa(int(p.currentBettings[k]))
		if i != len(p.currentBettings) {
			bet += "，"
		}
		for _, selectedWinningItem := range p.manager.selectedWinningItems {
			if k == selectedWinningItem.WinningBetZoneIndex {
				// if yes, add to award
				// isWin = true
				awarded = p.addToAwarded(awarded, selectedWinningItem)
				winningBet += p.currentBettings[k]
			}
		} // end of for
		// if !isWin {
		// 	// if not minus from award
		// 	awarded = p.substractFromAwarded(awarded, k)
		// }
		totalbetting += p.currentBettings[k]
	}
	bet += "]" // for logging
	// if awarded > 0 {
	// playerWinnings = awarded - winningBet
	playerWinnings = awarded - totalbetting // wk add
	// } else {
	// 	playerWinnings = awarded
	// }
	p.awarded = awarded // wk add
	p.setAwarded(playerWinnings, totalbetting, bet)
}

func (p *Player) calcPlayerResult() { //, m Manager
	var awarded int64 // raw winnings
	var totalbetting int64
	var winningBet int64
	var playerWinnings int64 = 0 // amount player win or lose
	// var specialPrizeChinese = "普通"

	bet := "[押注：" // for logging
	i := 0
	for k := range p.currentBettings {
		i++

		playerBetZoneName := p.manager.betZones[k].Name
		bet += "" + playerBetZoneName + "｜" + strconv.Itoa(int(p.currentBettings[k]))
		if i != len(p.currentBettings) {
			bet += "，"
		}
		if k == p.manager.selectedWinningItem.WinningBetZoneIndex { //selected winnerItem
			awarded = p.addToAwarded(awarded, p.manager.selectedWinningItem)
			winningBet += p.currentBettings[k]
		}
		// else {
		// 	awarded = p.substractFromAwarded(awarded, k)
		// }
		totalbetting += p.currentBettings[k] // this is bet_total in log
	}
	bet += "]" // for logging
	// if awarded > 0 {
	// playerWinnings = awarded - winningBet
	playerWinnings = awarded - totalbetting // wk add
	// } else {
	// playerWinnings = awarded
	// }
	// logger.Logger.Infoln("totalbetting", totalbetting)
	// logger.Logger.Infoln("awarded", awarded)
	// logger.Logger.Infoln("winningBet", winningBet)
	// logger.Logger.Infoln("playerWinnings", playerWinnings)
	p.awarded = awarded // wk add
	p.setAwarded(playerWinnings, totalbetting, bet)
}

// helper functions
func (p *Player) setAwarded(playerWinnings int64, totalBetting int64, bet string) { // for calculate
	m := p.manager
	// db.UpdateGameCoinByUid(p.Uid(), awarded)
	// if awarded > 0 {
	// 	db.NewGameCoinTransaction(p.Uid(), awarded)
	// }
	if totalBetting > 0 {
		// p.awarded = playerWinnings
		otherInforStr := "" //"[开奖：" + winningItems + m.specialPrizeChinese + "]" //+ "，"
		otherInforStr += "[总押注：" + strconv.FormatInt(totalBetting, 10) +
			"，输赢：" + strconv.FormatInt(playerWinnings, 10) + "] "
		otherInforStr += bet
		game_coin := db.GetGameCoinByUid(p.Uid())
		logInformation := map[string]string{
			"uid":       strconv.FormatInt(p.Uid(), 10),
			"game":      "水果",
			"betTotal":  strconv.FormatInt(totalBetting, 10),
			"result":    strconv.FormatInt(playerWinnings, 10), // player bet
			"level":     "",
			"otherInfo": otherInforStr,
			"before":    strconv.FormatInt(game_coin, 10),
			"used":      strconv.FormatInt(playerWinnings, 10),
			"after":     strconv.FormatInt(game_coin+playerWinnings, 10),
		}
		m.NewLogInformation(logInformation)
		m.Lock()
		m.totalWinLose += playerWinnings
		m.Unlock()
	}
	db.UpdateGameCoinByUid(p.Uid(), (playerWinnings)) // update gamecoin of player in db
	//add 1007
	db.UpdateWinGameCoinByUid(p.Uid(), (playerWinnings)) // update wingamecoin of player in db
	go db.NewGameCoinTransaction(p.Uid(), playerWinnings)
}

func (p *Player) substractFromAwarded(awarded int64, k int) int64 {
	// for calculate
	awarded -= p.currentBettings[k]
	return awarded
}

func (p *Player) addToAwarded(awarded int64, selectedWinningItem *protocol.WinningItem) int64 {
	// for calculate
	k := selectedWinningItem.WinningBetZoneIndex
	reward := p.currentBettings[k] * int64(selectedWinningItem.Odds)
	p.betResult = protocol.BetResult{
		BetZoneKey: k,
		Odds:       selectedWinningItem.Odds,
		PlacedBet:  p.currentBettings[k],
		Reward:     reward,
	}
	awarded += reward
	return awarded
}

func (p *Player) clearAllBets() {
	m := p.manager
	for key, currentBettingAmount := range p.currentBettings {
		m.betZones[key].Total -= currentBettingAmount
		m.group.Broadcast("updateZone", &protocol.PlaceBetResponse{
			BetZoneKey: key,
			TotalBet:   m.betZones[key].Total,
			MyBet:      0,
			Uid:        p.uid,
		})
	}
	p.currentBettings = map[int]int64{}
}

func (p *Player) getAllBetsTotal() int64 {
	var total int64 = 0
	for k := range p.currentBettings {
		total += p.currentBettings[k]
	}
	return total
}

func (p *Player) getBet(key int) int64 {
	return p.currentBettings[key]
}

func (p *Player) placeBet(key int, amount int64) {
	if val, ok := p.currentBettings[key]; ok {
		val += amount
		p.currentBettings[key] = val
	} else {
		p.currentBettings[key] = amount
	}
}

// init functions

func newPlayer(s *session.Session, uid int64, name string, faceUri string, level int, m *Manager) *Player {
	p := &Player{
		uid:             uid,
		userName:        name,
		faceUri:         faceUri,
		level:           level,
		currentBettings: map[int]int64{},
		CurrentWinnings: map[int]int64{},
		manager:         m,
	}

	p.bindSession(s)

	return p
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
