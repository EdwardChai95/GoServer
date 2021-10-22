package game

import (
	"fmt"
	"strconv"

	"github.com/lonng/nano/session"
	"gitlab.com/wolfplus/gamespace-lhd/db"
)

type (
	Player struct {
		uid             int64  // 用户ID
		faceUri         string // 头像地址
		userName        string // 玩家名字
		session         *session.Session
		room            *Room
		currentBettings map[int]int64 // [gambler key] amount
		awarded         int64
		CurrentWinnings map[int]int64 // [gambler key] amount
	}
)

// business func

func (p *Player) calcPlayerResult() int64 {
	r := p.room
	p.awarded = 0
	var bankerWinnings int64 = 0

	var totalbetting int64 = 0 // for logging
	bet := "[押注："              // for logging
	i := 0

	for key, amount := range p.currentBettings { // key refers to gamblers index
		i++

		totalbetting += amount
		singleBet := fmt.Sprintf("%v｜%v", r.gamblers[key].Title, strconv.Itoa(int(amount)))
		bet += singleBet
		if i != len(p.currentBettings) {
			bet += "，"
		}
		var odds int64 = 1
		var winlose int64 = 0
		if r.gamblers[key].IsWin { // player wins
			if r.roomType == 0 {
				odds = r.gamblers[key].Odds
			}
			winlose += odds * amount // 改了这边
			p.CurrentWinnings[key] = odds * amount
		} else {
			if r.roomType == 0 {
				odds = r.banker.Odds
			}
			winlose -= odds * amount // 改了这边
			p.CurrentWinnings[key] = -(r.banker.Odds * amount)
		}

		p.awarded += winlose // 改了这边
		r.gamblers[key].Totalbetting += amount
		r.gamblers[key].WinLose += winlose
	}

	var tax int64 = 0
	bankerWinnings = -p.awarded
	if p.awarded > 0 {
		tax = int64(0.05 * float64(p.awarded))
		p.awarded = int64(0.95 * float64(p.awarded))
	}

	game_coin := db.GetGameCoinByUid(p.Uid())

	if p.awarded != 0 {
		db.UpdateGameCoinByUid(p.Uid(), p.awarded) // update gamecoin of player in db
		//add 1007
		db.UpdateWinGameCoinByUid(p.Uid(), p.awarded) // update wingamecoin of player in db
		go db.NewGameCoinTransaction(p.Uid(), p.awarded)
	}

	if totalbetting > 0 {
		bet += "]" // for logging
		// [开奖：" + winningBetZone.Name + "｜" + strconv.Itoa(winningBetZone.Odds) + "，" +
		// winningTextBetZone.Name + "｜" + strconv.Itoa(winningTextBetZone.Odds) + "]
		otherInforStr := ""
		otherInforStr += "[总押注：" + strconv.FormatInt(totalbetting, 10) + "，输赢：" + strconv.FormatInt(p.awarded, 10) +
			"，税收：" + strconv.FormatInt(tax, 10) + "] "
		otherInforStr += bet

		//New add by kelvin 0701
		if tax > 0 {
			go db.CollectTax(tax)
		}

		//		game_coin := db.GetGameCoinByUid(p.Uid())
		level := fmt.Sprintf(r.title + " " + r.label)

		logInformation := map[string]string{
			"uid":       strconv.FormatInt(p.Uid(), 10),
			"game":      "拼十",
			"betTotal":  strconv.FormatInt(totalbetting, 10),
			"result":    strconv.FormatInt(p.awarded, 10), // player bet
			"level":     level,
			"otherInfo": otherInforStr,
			"before":    strconv.FormatInt(game_coin, 10),
			"used":      strconv.FormatInt(p.awarded, 10),
			"after":     strconv.FormatInt(game_coin+p.awarded, 10),
			"tax":       strconv.FormatInt(tax, 10),
		}

		go r.NewLogInformation(logInformation)
	}

	return bankerWinnings
}

func (p *Player) GetTotalBetting() int64 {
	var totalBetting int64
	for _, amount := range p.currentBettings {
		totalBetting += amount
	}
	return totalBetting
}

func (p *Player) GetBet(key int) int64 {
	return p.currentBettings[key]
}

func (p *Player) PlaceBet(key int, amount int64) (bool, string) {
	gamecoin := db.GetGameCoinByUid(p.uid) // get current gamecoin first
	currentRoom := p.room

	if gamecoin >= amount && currentRoom.gameStatus == "betting" {

		if (amount + p.GetTotalBetting()) > 2000000 {
			//return false, "当局最多押注1000万"
			return false, "Lên đến 10 triệu cược"
		}

		if currentRoom.roomType == 1 {
			if p.GetTotalBetting() > gamecoin || amount > gamecoin {
				//return false, "剩余数额不足"
				return false, "Không đủ tiền"
			}
		} else {
			if p.GetTotalBetting()*10 > gamecoin ||
				amount*10 > gamecoin {
				//return false, "剩余数额不足"
				return false, "Không đủ tiền"
			}
		}

		if val, ok := p.currentBettings[key]; ok { // key refers to gamblers index
			if currentRoom.roomType == 1 {
				if p.currentBettings[key] > gamecoin {
					//return false, "剩余数额不足"
					return false, "Không đủ tiền"
				}
			} else {
				if (amount+p.currentBettings[key])*10 > gamecoin {
					//return false, "剩余数额不足"
					return false, "Không đủ tiền"
				}
			}
			if (amount+p.currentBettings[key])*10 > int64(currentRoom.bankerCoins) {
				//return false, "所押注金额不能大于庄家赔付最大金额"
				return false, "Nhà cái không có đủ tiền"
			}
			val += amount
			p.currentBettings[key] = val
		} else {
			p.currentBettings[key] = amount
		}
		currentRoom.gamblers[key].Totalbetting += amount
		return true, ""
	} else {
		return false, ""
	}
}

func newPlayer(s *session.Session, uid int64, name string, faceUri string, room *Room) *Player {
	p := &Player{
		uid:             uid,
		userName:        name,
		faceUri:         faceUri,
		currentBettings: map[int]int64{},
		CurrentWinnings: map[int]int64{},
	}

	p.bindSession(s)
	p.bindRoom(room)

	return p
}

func (p *Player) bindRoom(room *Room) {
	p.room = room
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
