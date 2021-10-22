package game

import (
	"fmt"
	"strconv"

	"gitlab.com/wolfplus/gamespace-lhd/protocol"
)

func NewRobotPlayer() *protocol.RobotPlayer {
	return &protocol.RobotPlayer{
		UserName:        robotNames[randomRange(0, len(robotNames)-1)],
		CurrentBettings: map[int]int64{},
		Awarded:         0,
		GameCoin:        randomRange(300000000, 900000000),
	}
}

// robot players business logic
func (r *Room) robotPlayerCalcResult(rp *protocol.RobotPlayer) int64 {
	var totalbetting int64 = 0 // forlogging
	bet := "[押注："              // for logging
	i := 0

	var bankerWinnings int64 = 0
	rp.Awarded = 0
	for key, amount := range rp.CurrentBettings { // key refers to gamblers index
		i++

		totalbetting += amount
		singleBet := fmt.Sprintf("%v｜%v", r.gamblers[key].Title, strconv.Itoa(int(amount)))
		bet += singleBet
		if i != len(rp.CurrentBettings) {
			bet += "，"
		}

		var odds int64 = 1
		var winlose int64 = 0
		if r.gamblers[key].IsWin { // player wins
			if r.roomType == 0 {
				odds = r.gamblers[key].Odds
			}
			winlose = -odds * amount
		} else {
			if r.roomType == 0 {
				odds = r.banker.Odds
			}
			winlose = -odds * amount
		}

		rp.Awarded = winlose
		r.gamblers[key].Totalbetting += amount
		r.gamblers[key].WinLose += winlose
	}
	bankerWinnings = -rp.Awarded
	var tax int64 = 0
	if rp.Awarded > 0 {
		tax = int64(0.05 * float64(rp.Awarded))
		rp.Awarded = int64(0.95 * float64(rp.Awarded))
	}
	rp.CurrentBettings = map[int]int64{} // reset current bettings for the robot

	if totalbetting > 0 {
		bet += "]" // for logging
		// [开奖：" + winningBetZone.Name + "｜" + strconv.Itoa(winningBetZone.Odds) + "，" +
		// winningTextBetZone.Name + "｜" + strconv.Itoa(winningTextBetZone.Odds) + "]
		otherInforStr := ""
		otherInforStr += "[总押注：" + strconv.FormatInt(totalbetting, 10) + "，输赢：" + strconv.FormatInt(rp.Awarded, 10) +
			"，税收：" + strconv.FormatInt(tax, 10) + "] "
		otherInforStr += bet

		logInformation := map[string]string{
			"uid":       strconv.FormatInt(-1, 10),
			"game":      "拼十",
			"betTotal":  strconv.FormatInt(totalbetting, 10),
			"result":    strconv.FormatInt(rp.Awarded, 10), // player bet
			"level":     r.label,
			"otherInfo": otherInforStr,
			"before":    strconv.FormatInt(int64(rp.GameCoin), 10),
			"used":      strconv.FormatInt(rp.Awarded, 10),
			"after":     strconv.FormatInt(int64(rp.GameCoin)+rp.Awarded, 10),
			"tax":       strconv.FormatInt(tax, 10),
		}

		go r.NewLogInformation(logInformation)
	}

	return bankerWinnings
}

func (r *Room) robotPlayerPlaceBet(rp *protocol.RobotPlayer) {
	if r.gameStatus != "betting" { // 1. make sure room status is betting
		return
	}

	var robotBetTotal int64 = 0
	for k := range rp.CurrentBettings {
		if val, ok := rp.CurrentBettings[k]; ok {
			robotBetTotal += val
		}
	}
	if robotBetTotal > 500000 { // 1.1 make sure total bet is less than 2000000
		return
	}

	var key = randomRange(0, len(r.gamblers)-1)                            // 2. random bet choice
	var bettingCoins = []int64{100, 1000, 10000, 100000, 1000000, 5000000} // 3. available bet coins
	var amount = bettingCoins[randomRange(0, len(bettingCoins)-4)]         // 4. random bet amount  10000
	if val, ok := rp.CurrentBettings[key]; ok {                            // 5. update current betting
		val += amount
		rp.CurrentBettings[key] = val
	} else {
		rp.CurrentBettings[key] = amount
	}
	r.gamblers[key].Totalbetting += amount                         // 6. update bet zone
	r.group.Broadcast("updateGambler", &protocol.PlaceBetResponse{ // 7. broadcast to the other players
		Key:    key,
		Total:  r.gamblers[key].Totalbetting,
		MyBet:  rp.CurrentBettings[key],
		Uid:    -1,
		Amount: amount,
	})
}
