package game

import (
	"strconv"
	"sync"

	"gitlab.com/wolfplus/gamespace-lhd/db"
	"gitlab.com/wolfplus/gamespace-lhd/protocol"
)

// robot options
var robotNames = []string{"Minh Huệ", "Ngọc Thanh", "Lý Mỹ Kỳ", "Hồ Vĩnh Khoa", "Nguyễn Kim Hồng",
	"Phạm Gia Chi Bảo", "Ngoc Trinh", "Nguyễn Hoàng Bích", "Đặng Thu Thảo", "Nguyen Thanh Tung"}

var RoomRobotPlayers = []*RobotPlayer{newRobotPlayer(), newRobotPlayer(), newRobotPlayer(), newRobotPlayer(), newRobotPlayer(), newRobotPlayer(), newRobotPlayer(), newRobotPlayer(), newRobotPlayer(), newRobotPlayer()}

func newRobotPlayer() *RobotPlayer {
	return &RobotPlayer{
		userName:        robotNames[randomRange(0, len(robotNames)-1)],
		currentBettings: map[CurrentBetting]int64{},
		awarded:         0,
		gameCoin:        int64(randomRange(300000000, 900000000)),
	}
}

// robot player business logic

func (m *Manager) robotPlayerCalcResult(rp *RobotPlayer, winningBetZone *protocol.BetItem) (int64, int64, int64, int64) {
	var awarded int64       // this player's total raw winning
	var totalbetting int64  // this player's total gamecoin spend into betting this round
	var bankerwinning int64 // bankerwinning of this player

	bet := "[押注：" // for logging
	i := 0

	for k := range rp.currentBettings {
		i++
		bet += "" + m.betZones[k.key].Name + "｜" + strconv.Itoa(int(rp.currentBettings[k]))
		if i != len(rp.currentBettings) {
			bet += "，"
		}
		if m.betZones[k.key] == m.winningBetZone { // player bet on the right bet zone
			reward := rp.currentBettings[k] * int64(winningBetZone.Odds) // rewards = bet amount x odds
			awarded += reward                                            // add to your current winning
			bankerwinning -= (reward - rp.currentBettings[k])
		} else {
			bankerwinning += rp.currentBettings[k]
			// log.Infof("bankerwinning: %d", rp.currentBettings[k])
		}
		totalbetting += rp.currentBettings[k]
	}
	bet += "]"

	var playerWinnings int64 = 0 // amount player win or lose
	var playerTax int64 = 0      // amount that player deduct if they win

	playerWinnings = -totalbetting

	if awarded > 0 {
		playerWinnings = awarded - totalbetting - int64(float64(awarded-totalbetting)*0.05)
		playerTax = int64(float64(awarded) * 0.05)
	}

	rp.awarded = playerWinnings //  winlose of this round

	if totalbetting > 0 {
		otherInforStr := ""
		otherInforStr += "[总押注：" + strconv.FormatInt(totalbetting, 10) + "，输赢：" + strconv.FormatInt(playerWinnings, 10) +
			"，税收：" + strconv.FormatInt(playerTax, 10) + "] "
		otherInforStr += bet

		game_coin := rp.gameCoin

		logInformation := map[string]string{
			"uid":       strconv.FormatInt(-1, 10),
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
	rp.gameCoin += playerWinnings

	return playerWinnings, playerTax, totalbetting, bankerwinning
}

func (m *Manager) robotPlayerPlaceBet(rp *RobotPlayer) {
	rp.Lock()
	defer rp.Unlock()

	if m.gameStatus != "betting" {
		return
	}

	var robotBetTotal int64 = 0
	for k := range rp.currentBettings {
		robotBetTotal += rp.currentBettings[k]
	}
	if robotBetTotal > 2000000 {
		return
	}

	var key = randomRange(0, 7) // random bet choice
	// var amount = int64(randomRange(0, rp.gameCoin))
	// if rp.gameCoin > 100000 {
	// 	amount = int64(randomRange(1, 100)) * 1000
	// }

	var bettingCoins = []int64{1000, 10000, 100000, 1000000, 5000000}
	var amount = bettingCoins[randomRange(0, 2)]
	betkey := CurrentBetting{
		key: key,
	}
	if val, ok := rp.currentBettings[betkey]; ok {
		val += amount
		rp.currentBettings[betkey] = val
	} else {
		rp.currentBettings[betkey] = amount
	}
	m.Lock()
	m.betZones[key].Total += amount
	m.group.Broadcast("updateZone", &protocol.PlaceBetResponse{
		Key:   key,
		Total: m.betZones[key].Total,
		MyBet: rp.currentBettings[betkey],
		Uid:   -1,
	})
	m.Unlock()
}

// end robot player business logic

type (
	RobotPlayer struct {
		userName        string
		currentBettings map[CurrentBetting]int64
		awarded         int64
		gameCoin        int64
		sync.RWMutex
	}
)
