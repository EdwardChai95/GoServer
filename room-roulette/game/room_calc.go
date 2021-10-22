package game

import (
	"strconv"

	"gitlab.com/wolfplus/gamespace-lhd/db"
	"gitlab.com/wolfplus/gamespace-lhd/protocol"
)

func (r *Room) calcParticipantResult(currentBettings map[protocol.CurrentBetting]int64) (int64, int64, int64, protocol.BetResult, protocol.BetResult, int64, map[string]string) { // both robot and player
	var awarded int64 // to return
	var loss int64
	var totalbetting int64 // to return
	var winningBet int64   // to return

	animalBet := protocol.BetResult{} // to return
	textBet := protocol.BetResult{}   // to return
	var bonus int64 = 0               // to return

	logInformation := map[string]string{}

	bet := "[押注：" // for logging
	i := 0
	for k := range currentBettings {
		i++
		zoneRow := r.getBetZoneByColor(k.Zone)
		zoneRowChild := zoneRow[k.Key]
		bet += "" + zoneRowChild.Name + "｜" + strconv.Itoa(int(currentBettings[k]))
		if i != len(currentBettings) {
			bet += "，"
		}
		if k == r.winningBetZone ||
			(r.special == "bigfour" && k.Zone == r.winningBetZone.Zone) || // 大四喜 same color bigfour
			(r.special == "bigthree" && k.Key == r.winningBetZone.Key) { // 大三元 same animal bigthree
			animalwinning := currentBettings[k] * int64(zoneRowChild.Odds)
			animalBet = protocol.BetResult{
				Zone:      k.Zone,
				Key:       k.Key,
				Odds:      zoneRowChild.Odds,
				PlacedBet: currentBettings[k],
				Reward:    animalwinning,
			}
			awarded += animalwinning
			if k != r.winningBetZone {
				bonus += animalwinning
			}
			winningBet += currentBettings[k]
		} else if r.winningTextBetZone == k {
			textwinning := currentBettings[k] * int64(zoneRowChild.Odds)
			textBet = protocol.BetResult{
				Zone:      k.Zone,
				Key:       k.Key,
				Odds:      zoneRowChild.Odds,
				PlacedBet: currentBettings[k],
				Reward:    textwinning,
			}
			awarded += textwinning
			winningBet += currentBettings[k]
		}
		loss += currentBettings[k]
		totalbetting += currentBettings[k] // this is bet_total in log
	}
	bet += "]" // for logging
	win_total := awarded - loss

	if r.special == "double" {
		bonus = awarded
		win_total += awarded
		awarded += awarded
	}

	if totalbetting > 0 {
		otherInforStr := ""
		otherInforStr += "[总押注：" + strconv.FormatInt(totalbetting, 10) + "，输赢：" +
			strconv.FormatInt(win_total, 10) + "] "
		otherInforStr += bet

		// game_coin := db.GetGameCoinByUid(p.Uid())

		logInformation = map[string]string{
			// "uid":       strconv.FormatInt(p.Uid(), 10),
			"game":      "动物乐园",
			"betTotal":  strconv.FormatInt(totalbetting, 10),
			"result":    strconv.FormatInt(win_total, 10), // player bet
			"level":     r.level,
			"otherInfo": otherInforStr,
			// "before":    strconv.FormatInt(game_coin, 10),
			"used": strconv.FormatInt(win_total, 10),
			// "after":     strconv.FormatInt(game_coin+win_total, 10),
		}
	}

	return totalbetting, awarded, winningBet, animalBet, textBet, bonus, logInformation
}

func (r *Room) getParticipantCut(totalWinningBet int64, currentBettings map[protocol.CurrentBetting]int64) float64 {
	// robot and player both
	var participantcut float64
	var playerbetting int64
	for k := range currentBettings {
		if k == r.winningBetZone || r.winningTextBetZone == k {
			// 押注总额 （total）
			playerbetting += currentBettings[k] // 中奖押注额
			participantcut += float64(currentBettings[k]) /
				float64(totalWinningBet) * r.special_prize

			// log.Infof("current bettings: %v", currentBettings[k])
			// log.Infof("total winning bet: %v", totalWinningBet)
		}
	}
	if participantcut > float64(playerbetting*40) {
		participantcut = float64(playerbetting * 40)
	}
	return participantcut
}

func (p *Player) calcPlayerResult(winningBetZone protocol.BetItem, winningTextBetZone protocol.BetItem) (int64, int64, int64) {
	r := p.room
	totalbetting, awarded, winningBet,
		animalBet, textBet, bonus, logInformation := r.calcParticipantResult(p.currentBettings)

	p.awarded = awarded
	p.animalBet = animalBet
	p.textBet = textBet
	p.bonus = bonus
	p.logInformation = logInformation
	win_total, _ := strconv.ParseInt(p.logInformation["used"], 10, 64)

	if totalbetting > 0 {
		game_coin := db.GetGameCoinByUid(p.Uid())
		p.logInformation["uid"] = strconv.FormatInt(p.Uid(), 10)
		p.logInformation["before"] = strconv.FormatInt(game_coin, 10)
		p.logInformation["after"] = strconv.FormatInt(game_coin+win_total, 10)

		go r.NewLogInformation(p.logInformation)
	}

	db.UpdateGameCoinByUid(p.Uid(), (win_total)) // update gamecoin of player in db
	//add 1007
	db.UpdateWinGameCoinByUid(p.Uid(), (win_total)) // update wingamecoin of player in db
	go db.NewGameCoinTransaction(p.Uid(), win_total)

	return totalbetting, awarded, winningBet
}
