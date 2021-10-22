package game

import (
	"strconv"
	"sync"

	"gitlab.com/wolfplus/gamespace-lhd/protocol"
)

// robot player logic function

func (m *Manager) setRobotsLoggingInfo(rp *RobotPlayer, robotWinnings int64, totalbetting int64, bet string) {
	if totalbetting > 0 {
		otherInforStr := "" //"[开奖：" + winningItems + m.specialPrizeChinese + "]" //+ "，"
		otherInforStr += "[总押注：" + strconv.FormatInt(totalbetting, 10) +
			"，输赢：" + strconv.FormatInt(robotWinnings, 10) + "] "
		otherInforStr += bet
		logInformation := map[string]string{
			"uid":       strconv.FormatInt(-1, 10),
			"game":      "水果",
			"betTotal":  strconv.FormatInt(totalbetting, 10),
			"result":    strconv.FormatInt(robotWinnings, 10), // robot bet
			"level":     "",
			"otherInfo": otherInforStr,
			"before":    strconv.FormatInt(int64(rp.gameCoin), 10),
			"used":      strconv.FormatInt(robotWinnings, 10),
			"after":     strconv.FormatInt(int64(rp.gameCoin)+robotWinnings, 10),
		}
		m.NewLogInformation(logInformation)
	}
	m.Lock()
	m.totalWinLose += robotWinnings
	m.Unlock()
}

func (m *Manager) calcRobotSpecialResult(rp *RobotPlayer) {
	var awarded int64 // raw winnings
	var totalbetting int64
	var winningBet int64
	var robotWinnings int64 = 0 // amount robot win or lose

	bet := "[押注：" // for logging
	i := 0
	for k := range rp.currentBettings {
		// isWin := false
		i++

		robotsBetZoneName := m.betZones[k].Name
		bet += "" + robotsBetZoneName + "｜" + strconv.Itoa(int(rp.currentBettings[k]))
		if i != len(rp.currentBettings) {
			bet += "，"
		}
		for _, selectedWinningItem := range m.selectedWinningItems {
			if k == selectedWinningItem.WinningBetZoneIndex {
				// if yes, add to award
				// isWin = true
				k := selectedWinningItem.WinningBetZoneIndex
				reward := rp.currentBettings[k] * int64(selectedWinningItem.Odds)
				awarded += reward
				winningBet += rp.currentBettings[k]
			}
		} // end of for
		// if !isWin {
		// 	// if not minus from award
		// 	awarded -= rp.currentBettings[k]
		// }
		totalbetting += rp.currentBettings[k]
	}
	bet += "]" // for logging
	// if awarded > 0 {
	// robotWinnings = awarded - winningBet
	robotWinnings = awarded - totalbetting
	// } else {
	// 	robotWinnings = awarded
	// }
	rp.awarded = awarded //robotWinnings
	m.setRobotsLoggingInfo(rp, robotWinnings, totalbetting, bet)
}

func (m *Manager) robotPlayerCalcResult(rp *RobotPlayer) {
	var awarded int64 // raw winnings
	var totalbetting int64
	var winningBet int64
	var robotWinnings int64 = 0 // amount robot win or lose

	bet := "[押注：" // for logging
	i := 0
	for k := range rp.currentBettings {
		i++

		robotsBetZoneName := m.betZones[k].Name
		bet += "" + robotsBetZoneName + "｜" + strconv.Itoa(int(rp.currentBettings[k]))
		if i != len(rp.currentBettings) {
			bet += "，"
		}
		if k == m.selectedWinningItem.WinningBetZoneIndex { //selected winnerItem
			reward := rp.currentBettings[k] * int64(m.selectedWinningItem.Odds)
			awarded += reward
			winningBet += rp.currentBettings[k]
		}
		// else {
		// 	awarded -= rp.currentBettings[k]
		// }
		totalbetting += rp.currentBettings[k]
	}
	bet += "]" // for logging
	// if awarded > 0 {
	// robotWinnings = awarded - winningBet
	robotWinnings = awarded - totalbetting // wk add
	// } else {
	// 	robotWinnings = awarded
	// }
	rp.awarded = awarded // robotWinnings
	m.setRobotsLoggingInfo(rp, robotWinnings, totalbetting, bet)
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
	if robotBetTotal > 500000 { //99990000
		return
	}

	var key = randomRange(0, 7) // random bet index // CANEDIT
	var bettingCoins = []int64{10, 100, 1000, 10000, 100000, 1000000}
	var amount = bettingCoins[randomRange(0, 2)] // CANEDIT can make more dynamic

	if val, ok := rp.currentBettings[key]; ok {
		val += amount
		rp.currentBettings[key] = val
	} else {
		rp.currentBettings[key] = amount
	}
	m.Lock()
	m.betZones[key].Total += amount
	m.group.Broadcast("updateZone", &protocol.PlaceBetResponse{
		BetZoneKey: key,
		TotalBet:   m.betZones[key].Total,
		MyBet:      rp.currentBettings[key],
		Uid:        -1,
	})
	m.Unlock()
}

// robotPlayer
func newRobotPlayer() *RobotPlayer {
	return &RobotPlayer{
		userName:        robotNames[randomRange(0, len(robotNames)-1)],
		currentBettings: map[int]int64{},
		awarded:         0,
		gameCoin:        randomRange(300000000, 900000000),
		faceUri:         randomRange(1, 10),
		level:           randomRange(0, 20),
	}
}

type (
	RobotPlayer struct {
		userName        string
		currentBettings map[int]int64
		awarded         int64
		gameCoin        int
		faceUri         int
		level           int
		sync.Mutex
	}
)
