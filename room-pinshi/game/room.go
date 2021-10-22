package game

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lonng/nano"
	log "github.com/sirupsen/logrus"
	"gitlab.com/wolfplus/gamespace-lhd/db"
	"gitlab.com/wolfplus/gamespace-lhd/define"
	"gitlab.com/wolfplus/gamespace-lhd/protocol"
)

type (
	Room struct {
		group                  *nano.Group
		title                  string
		label                  string
		minGameCoin            int64
		roomType               int // 1 for 通比
		gamblers               []*protocol.Gambler
		banker                 *protocol.Gambler // DOES NOT HANDLE BANKER COINS AND ROUNDS
		chip                   []int64
		players                map[int64]*Player // room 的所有玩家
		gameStatus             string            // phase to facilitate game
		deadline               time.Time         // when to go to the next phase
		winners                []protocol.Winner
		bankerCoins            int
		bankerImg              int
		bankerRounds           int
		totalPlayerWinnings    int64                   // per round
		totalBankerWinnings    int64                   // per round
		historyList            [][]*protocol.Gambler   // history
		robotPlayers           []*protocol.RobotPlayer // robot players for this room
		collection_amount      int64
		collection_betting     int64
		collection_lastupdated time.Time
		logInformations        []map[string]string
	}
)

// robot options
var robotNames = []string{"Minh Huệ", "Ngọc Thanh", "Lý Mỹ Kỳ", "Hồ Vĩnh Khoa", "Nguyễn Kim Hồng",
	"Phạm Gia Chi Bảo", "Ngoc Trinh", "Nguyễn Hoàng Bích", "Đặng Thu Thảo", "Nguyen Thanh Tung"}

var cards = newCards()

// phases

func (r *Room) resultPhase() {
	if r.gameStatus != "animation" {
		return
	}

	r.gameStatus = "result"
	r.deadline = time.Now().UTC().Add(time.Second * time.Duration(6)) // 15 deadline change

	r.bankerRounds++
	r.bankerCoins += int(r.totalBankerWinnings)

	if r.bankerCoins <= 0 || r.bankerRounds >= 9 {
		r.bankerRounds = 0
		r.bankerImg = randomRange(1, 10)
		r.bankerCoins = randomRange(300000000, 900000000)
	}

	var x = 0
	for uid := range r.players {
		go func(uid int64, x int) {
			updateRobotWinners := false
			if x == 0 {
				updateRobotWinners = true
			}
			if p, ok := r.players[uid]; ok {
				gamecoin := db.GetGameCoinByUid(uid) // get current gamecoin
				if p.session != nil {
					p.session.Push("resultPhase", &protocol.ResultPhase{
						GameCoin:           gamecoin,
						Deadline:           r.deadline,
						TotalBetting:       p.GetTotalBetting(),
						Winnings:           p.awarded,
						Winners:            r.winners,
						BankerWinnings:     r.totalBankerWinnings,
						BankerCoins:        r.bankerCoins,
						BankerImg:          r.bankerImg,
						Gamblers:           r.gamblers,
						Banker:             r.banker,
						CurrentWinnings:    p.CurrentWinnings,
						UpdateRobotWinners: updateRobotWinners,
					})
				}

				// reset
				p.currentBettings = map[int]int64{}
				p.CurrentWinnings = map[int]int64{}
			}
		}(uid, x)
		x++
	}

	for i := range r.robotPlayers {
		// reset robot players
		go func(rp *protocol.RobotPlayer) {
			if rp.GameCoin <= 1000 {
				rp = NewRobotPlayer()
			}
		}(r.robotPlayers[i])
	}

	//if r.sessionCount() == 0 {
	//	r.gameStatus = "nogame"
	//} else {
	//	go func() {
	s := r.deadline.Sub(time.Now()).Seconds()
	time.Sleep(time.Duration(s) * time.Second)
	r.bettingPhase()
	//	}()
	//}

}

func (r *Room) animationPhase() {
	if r.gameStatus != "betting" {
		return
	}
	r.gameStatus = "animation"
	r.deadline = time.Now().UTC().Add(time.Second * time.Duration(6)) // 18 deadline change
	level := fmt.Sprintf(r.title + " " + r.label)

	// 分牌
	randCardIndexes := rand.Perm(len(cards))[:25]
	for i, index := range randCardIndexes {
		if i < 5 {
			r.banker.RoundCards = append(r.banker.RoundCards, cards[index])
		} else if i < 10 {
			r.gamblers[0].RoundCards = append(r.gamblers[0].RoundCards, cards[index])
		} else if i < 15 {
			r.gamblers[1].RoundCards = append(r.gamblers[1].RoundCards, cards[index])
		} else if i < 20 {
			r.gamblers[2].RoundCards = append(r.gamblers[2].RoundCards, cards[index])
		} else if i < 25 {
			r.gamblers[3].RoundCards = append(r.gamblers[3].RoundCards, cards[index])
		}
	}

	// r.collection_amount = 当日累计输金额（扣除税）
	// r.collection_betting = 当日总投注金额
	if r.collection_amount > 0 && r.collection_betting > 0 {
		if int64(float64(r.collection_amount)/float64(r.collection_betting)*100.0) >
			int64(define.CONFIG_A) ||
			int64(r.collection_amount) >
				int64(define.CONFIG_B) {
			if randomRange(1, 100) < define.CONFIG_X {
				// 确保庄家是拼十
				for {
					r.comboChecker(r.banker)

					if r.banker.Rank >= define.CONFIG_C {
						break
					}
					r.banker.RoundCards = []*protocol.Card{}
					r.gamblers[0].RoundCards = []*protocol.Card{}
					r.gamblers[1].RoundCards = []*protocol.Card{}
					r.gamblers[2].RoundCards = []*protocol.Card{}
					r.gamblers[3].RoundCards = []*protocol.Card{}

					randCardIndexes := rand.Perm(len(cards))[:25]
					for i, index := range randCardIndexes {
						if i < 5 {
							r.banker.RoundCards = append(r.banker.RoundCards, cards[index])
						} else if i < 10 {
							r.gamblers[0].RoundCards = append(r.gamblers[0].RoundCards, cards[index])
						} else if i < 15 {
							r.gamblers[1].RoundCards = append(r.gamblers[1].RoundCards, cards[index])
						} else if i < 20 {
							r.gamblers[2].RoundCards = append(r.gamblers[2].RoundCards, cards[index])
						} else if i < 25 {
							r.gamblers[3].RoundCards = append(r.gamblers[3].RoundCards, cards[index])
						}
					}
				}
			}
		}
	}

	// ======TEST CARDS
	// r.banker.RoundCards = []*protocol.Card{}
	// r.gamblers[0].RoundCards = []*protocol.Card{}
	// r.gamblers[1].RoundCards = []*protocol.Card{}
	// r.gamblers[2].RoundCards = []*protocol.Card{}
	// r.gamblers[3].RoundCards = []*protocol.Card{}
	// testCardIndexes := []int64{5, 13, 4, 20, 7, 49, 12, 10, 8, 44, 50, 51, 48, 37, 27, 2, 41, 42, 16, 46, 19, 45, 22, 17, 6}
	// for i, index := range testCardIndexes {
	// 	if i < 5 {
	// 		r.banker.RoundCards = append(r.banker.RoundCards, cards[index])
	// 	} else if i < 10 {
	// 		r.gamblers[0].RoundCards = append(r.gamblers[0].RoundCards, cards[index])
	// 	} else if i < 15 {
	// 		r.gamblers[1].RoundCards = append(r.gamblers[1].RoundCards, cards[index])
	// 	} else if i < 20 {
	// 		r.gamblers[2].RoundCards = append(r.gamblers[2].RoundCards, cards[index])
	// 	} else if i < 25 {
	// 		r.gamblers[3].RoundCards = append(r.gamblers[3].RoundCards, cards[index])
	// 	}
	// }
	// ======END TEST CARDS

	r.comboCheckers(r.gamblers, r.banker)

	go db.RecordWinningItem(r.gamblers[0], r.gamblers[1], r.gamblers[2], r.gamblers[3], level) // level

	if len(r.historyList) >= 10 {
		copy(r.historyList[0:], r.historyList[1:])
		r.historyList[len(r.historyList)-1] = []*protocol.Gambler{} // Erase last element (write zero value).
		r.historyList = r.historyList[:len(r.historyList)-1]        // Truncate slice.
	}
	r.historyList = append(r.historyList, r.newSingleHistoryList())

	animationPhaseRes := &protocol.AnimationPhaseResponse{
		Deadline: r.deadline,
		Gamblers: r.gamblers,
		Banker:   r.banker,
	}

	r.group.Broadcast("animationPhase", animationPhaseRes)

	var wg sync.WaitGroup
	r.totalPlayerWinnings = 0
	r.totalBankerWinnings = 0

	var roundBetTotal int64 = 0

	var collectionUsed int64 = 0     // for human players only
	var collectionBetTotal int64 = 0 // for human players only

	for _, p := range r.players {
		wg.Add(1)
		go func(p *Player) {
			playerTotalBetting := p.GetTotalBetting()
			bankerWinnings := p.calcPlayerResult()
			r.totalPlayerWinnings += p.awarded
			r.totalBankerWinnings += bankerWinnings

			roundBetTotal += playerTotalBetting
			collectionUsed += p.awarded
			collectionBetTotal += playerTotalBetting
			wg.Done()
		}(p)
	}
	wg.Wait()

	// calc robot winnings
	for i := range r.robotPlayers {
		wg.Add(1)
		go func(robotPlayer *protocol.RobotPlayer) {
			var robotBetTotal int64 = 0
			for k := range robotPlayer.CurrentBettings {
				robotBetTotal += robotPlayer.CurrentBettings[k]
			}
			// roundBetTotal += robotBetTotal
			r.robotPlayerCalcResult(robotPlayer) // bankerWinnings :=
			// r.totalPlayerWinnings += robotPlayer.Awarded
			// r.totalBankerWinnings += bankerWinnings
			wg.Done()
		}(r.robotPlayers[i])
	}
	wg.Wait()

	var tax int64 = 0
	if r.totalBankerWinnings > 0 {
		tax = int64(0.05 * float64(r.totalBankerWinnings))
		r.totalBankerWinnings = int64(0.95 * float64(r.totalBankerWinnings)) // 税5%，庄家赢扣税
	}

	// animationOut, _ := json.Marshal(animationPhaseRes)

	// [青龙：A10B7C5A5C4，押10343100，赢-10000][白虎：A10B7C5A5C4，押10343100，赢-10000]
	gamblersLog := ""
	systemGamblersLog := ""
	for _, gambler := range r.gamblers {
		gamblerLog := fmt.Sprintf("[%v]", gambler.Rank)
		gamblersLog += gamblerLog

		systemGamblerLog := fmt.Sprintf("[%v：%v，押%v，赢%v]", gambler.Title,
			gambler.CardValues, gambler.Totalbetting, gambler.WinLose)
		systemGamblersLog += systemGamblerLog
	}

	gamblerLog := fmt.Sprintf("[%v]", r.banker.Rank)
	gamblersLog += gamblerLog

	systemGamblerLog := fmt.Sprintf("[庄家：%v]", r.banker.CardValues)
	systemGamblersLog += systemGamblerLog

	otherInfo := "开奖：" + systemGamblersLog
	otherInfo += fmt.Sprintf("[总押注：%v， 总输赢：%v，税：%v]", roundBetTotal, -r.totalPlayerWinnings, tax)
	logInformation := map[string]string{
		"uid":            "0",
		"game":           "拼十",
		"level":          level,
		"otherInfo":      otherInfo,
		"betTotal":       strconv.Itoa(int(roundBetTotal)),
		"winTotal":       strconv.Itoa(int(-r.totalPlayerWinnings)),
		"bankerWinTotal": strconv.Itoa(int(r.totalBankerWinnings)),
		"before":         strconv.Itoa(int(r.bankerCoins)),
		"used":           strconv.Itoa(int(r.totalBankerWinnings)),
		"after":          strconv.Itoa(int(r.bankerCoins + int(r.totalBankerWinnings))),
		"tax":            strconv.Itoa(int(tax)),
	}
	go r.InsertAllLogInformations(logInformation, gamblersLog)

	r.collection_update(collectionUsed, collectionBetTotal)

	// sort by the winnings of each players
	var awards = []int{}
	keys := map[int]int64{}
	for k, p := range r.players {
		keys[int(p.awarded)] = k
		awards = append(awards, int(p.awarded))
	}
	for i, rp := range r.robotPlayers { // robot players added to sort
		keys[int(rp.Awarded)] = int64(i) // this is not a real uid
		awards = append(awards, int(rp.Awarded))
	}
	sort.Sort(sort.Reverse(sort.IntSlice(awards))) // big to small
	// populate winners with the sorted result
	r.winners = []protocol.Winner{}
	for i := range awards {
		if len(r.winners) >= 5 {
			break
		}
		key := keys[awards[i]] // uid
		if p, ok := r.players[key]; ok {
			r.winners = append(r.winners, protocol.Winner{
				Username: p.userName,
				Awarded:  awards[i],
				IsRobot:  false,
			})
		} else {
			r.winners = append(r.winners, protocol.Winner{
				Username: r.robotPlayers[key].UserName,
				Awarded:  awards[i],
				IsRobot:  true,
			})
		}
	}

	go func() {
		s := r.deadline.Sub(time.Now()).Seconds()
		time.Sleep(time.Duration(s) * time.Second)
		r.resultPhase()
	}()
}

func (r *Room) bettingPhase() {

	// reset for betting
	for i := range r.gamblers {
		r.gamblers[i].Totalbetting = 0
		r.gamblers[i].RoundCards = []*protocol.Card{}
	}
	r.banker.RoundCards = []*protocol.Card{}

	r.gameStatus = "betting"

	r.deadline = time.Now().UTC().Add(time.Second * time.Duration(10)) // 27 deadline change
	// log.Infof("BETTING DEADLINE: %d", r.deadline)
	// fmt.Println(r.deadline.String())

	r.group.Broadcast("bettingPhase", &protocol.BettingPhaseResponse{
		Deadline: r.deadline,
	})

	// count down to next game status
	go func() {
		s := r.deadline.Sub(time.Now()).Seconds() + 1
		time.Sleep(time.Duration(s) * time.Second)
		r.animationPhase()
	}()

	for i := range r.robotPlayers {
		var betTimes = randomRange(3, 10)
		for x := 0; x < betTimes; x++ {
			go func(robotPlayer *protocol.RobotPlayer) {
				s := r.deadline.Sub(time.Now()).Seconds() + 1
				n := randomRange(3000, int(s)*1000)
				time.Sleep(time.Duration(n) * time.Millisecond)
				r.robotPlayerPlaceBet(robotPlayer)
			}(r.robotPlayers[i])
		}
	}
}

// utility functions
func (r *Room) collection_update(collectionUsed, total int64) {
	// now := db.GetCurrentShanghaiTime()
	// // if collectionUsed > 0 {
	// if now.Day() != r.collection_lastupdated.Day() {
	// 	r.collection_amount = collectionUsed
	// 	r.collection_betting = 0
	// } else {
	r.collection_amount += collectionUsed
	// }
	// }
	r.collection_betting += total
	// r.collection_lastupdated = now
	level := fmt.Sprintf(r.title + " " + r.label)
	db.NewGameCollection("拼十", level, r.collection_amount, r.collection_betting)
}

func (r *Room) newSingleHistoryList() []*protocol.Gambler {
	g := NewGamblers()
	for i, gambler := range g {
		*gambler = *r.gamblers[i]
	}
	return g
}

func (r *Room) comboCheckers(gamblers []*protocol.Gambler, banker *protocol.Gambler) {
	r.comboChecker(banker)             // determine odds AND rank
	for _, gambler := range gamblers { // by default a gambler loses
		r.comboChecker(gambler) // determine odds AND rank
		if gambler.Odds > banker.Odds {
			gambler.IsWin = true
		} else if gambler.Odds == banker.Odds {
			// here we determine if gambler and banker has the same odds
			// 点数一样按照散排情况比大小；
			if gambler.Rank > banker.Rank {
				// 点数不一样按照牌型依次比大小：五花>四炸>葫芦>双十>十代九>十代八>十代七>十代六>十代五>十代四>十代三>十代二>十代一>散牌；
				gambler.IsWin = true
				continue
			}
			if banker.Rank > gambler.Rank {
				// you lose so NO NEED to carry on compare flow
				continue // continue comparing other gamblers
			}

			// 散排情况
			same5cards := true       // assumption before comparing
			for i := 0; i < 5; i++ { // 先比最大的牌的点数
				if gambler.OrderedCardValues[i].FaceValue > banker.OrderedCardValues[i].FaceValue {
					gambler.IsWin = true // already determined a winner so can break
					same5cards = false
					break // break out of 散排情况
				} else if gambler.OrderedCardValues[i].FaceValue < banker.OrderedCardValues[i].FaceValue { // 如果一样则比第二大的牌的点数
					same5cards = false // gambler lose
					break
				}
			}

			if same5cards { // 5张牌点数都一样大
				gamblerTempValue := gambler.OrderedCardValues[0].FaceValue
				if gambler.OrderedCardValues[0].Suit == "c" {
					gamblerTempValue += 1
				} else if gambler.OrderedCardValues[0].Suit == "h" {
					gamblerTempValue += 2
				} else if gambler.OrderedCardValues[0].Suit == "s" {
					gamblerTempValue += 3
				}
				bankerTempValue := banker.OrderedCardValues[0].FaceValue
				if banker.OrderedCardValues[0].Suit == "c" {
					bankerTempValue += 1
				} else if banker.OrderedCardValues[0].Suit == "h" {
					bankerTempValue += 2
				} else if banker.OrderedCardValues[0].Suit == "s" {
					bankerTempValue += 3
				}

				if gamblerTempValue > bankerTempValue {
					gambler.IsWin = true
				}
			}
		}
	}
}

func (r *Room) comboChecker(gambler *protocol.Gambler) { // determine odds AND rank
	gambler.Totalbetting = 0
	gambler.WinLose = 0

	orderedCardValues := []int{} // for comparing in case of a tie in odds
	tempMaps := map[int]*protocol.Card{}

	for _, gcard := range gambler.RoundCards {
		tempValue := gcard.Value
		// tempValue will calculate also the color of the card to really determine best card
		// i.e. spade jack will be top card if your hand also consist of diamond jack
		gcard.FaceValue = gcard.Value
		if gcard.Face == "J" {
			gcard.FaceValue += 1
			tempValue += 1
		} else if gcard.Face == "Q" {
			gcard.FaceValue += 2
			tempValue += 2
		} else if gcard.Face == "K" {
			gcard.FaceValue += 3
			tempValue += 3
		} else if gcard.Face == "Jo" {
			gcard.FaceValue += 4
			tempValue += 4
		}

		orderedCardValues = append(orderedCardValues, tempValue)
		tempMaps[tempValue] = gcard
	}
	sort.Sort(sort.Reverse(sort.IntSlice(orderedCardValues))) // big to small
	gambler.OrderedCardValues = []*protocol.Card{}
	for i := range orderedCardValues {
		tempCard := tempMaps[orderedCardValues[i]]
		gambler.OrderedCardValues = append(gambler.OrderedCardValues, tempCard)
	}

	// FOR LOGGING
	cardValues := ""
	for _, card := range gambler.RoundCards {
		cardValues += strings.ToUpper(card.Suit)
		face := card.Face
		if face == "J" {
			face = "11"
		} else if face == "Q" {
			face = "12"
		} else if face == "K" {
			face = "13"
		}
		cardValues += face
	}
	gambler.CardValues = cardValues

	gambler.Odds = 1
	gambler.Rank = 0
	gambler.Combo = "meiniu"
	gambler.IsWin = false // default value ( for banker)

	if (gambler.RoundCards[0].Face == "J" ||
		gambler.RoundCards[0].Face == "Q" ||
		gambler.RoundCards[0].Face == "K") &&
		(gambler.RoundCards[1].Face == "J" ||
			gambler.RoundCards[1].Face == "Q" ||
			gambler.RoundCards[1].Face == "K") &&
		(gambler.RoundCards[2].Face == "J" ||
			gambler.RoundCards[2].Face == "Q" ||
			gambler.RoundCards[2].Face == "K") &&
		(gambler.RoundCards[3].Face == "J" ||
			gambler.RoundCards[3].Face == "Q" ||
			gambler.RoundCards[3].Face == "K") &&
		(gambler.RoundCards[4].Face == "J" ||
			gambler.RoundCards[4].Face == "Q" ||
			gambler.RoundCards[4].Face == "K") {
		// 五花
		if r.roomType == 0 {
			gambler.Odds = 10
		}
		gambler.Rank = 13
		gambler.Combo = "wuhuaniu"
		return
	}

	// check for 四带一
	fcards := combinationUtil(gambler.RoundCards,
		[]map[int]*protocol.Card{}, map[int]*protocol.Card{}, 0, len(gambler.RoundCards)-1, 0, 4)
	for _, combi := range fcards {
		if combi[0].Face == combi[1].Face &&
			combi[0].Face == combi[2].Face &&
			combi[0].Face == combi[3].Face {
			gambler.Combo = "zhadan"
			if r.roomType == 0 {
				gambler.Odds = 8
			}
			gambler.Rank = 12
			return
		}
	}

	pinshis := combinationUtil(gambler.RoundCards,
		[]map[int]*protocol.Card{}, map[int]*protocol.Card{}, 0, len(gambler.RoundCards)-1, 0, 3)
	dai := map[int][]*protocol.Card{}

	for j, combi := range pinshis { // combi is the 3 cards option
		if combi[0].Face == combi[1].Face &&
			combi[0].Face == combi[2].Face { // potential 三带二
			daiCombo := getDaiCombo(gambler, combi) // get the remaining two cards

			if len(daiCombo) == 2 && daiCombo[0].Face == daiCombo[1].Face { // checking for hulu
				gambler.RoundCards = []*protocol.Card{}
				for _, card := range combi {
					gambler.RoundCards = append(gambler.RoundCards, card)
				}
				for _, card := range daiCombo {
					gambler.RoundCards = append(gambler.RoundCards, card)
				}
				gambler.Combo = "hulu"
				if r.roomType == 0 {
					gambler.Odds = 6
				}
				gambler.Rank = 11
				return
			}
		}

		combiTotal := 0
		for _, card := range combi {
			combiTotal += card.Value
		}
		if combiTotal == 10 || combiTotal%10 == 0 { // valid pinshi

			daiCombo := getDaiCombo(gambler, combi) // get the remaining two cards

			if len(daiCombo) == 2 {
				dai[j] = daiCombo
			}
		}
	} // end of for loop

	// EDIT HERE for 10 + 2 cards with the sum of 10

	if len(dai) > 0 {
		var totals = []int{}
		keys := map[int]int{}

		for i, combi := range dai {
			total := 0
			for _, card := range combi {
				// log.Infof("dai cards: %v %v", x, card)
				total += card.Value
			}
			// for x, card := range pinshis[i] {
			// 	log.Infof("pinshi cards: %v %v", x, card)
			// }
			keys[total] = i
			totals = append(totals, total)
		}

		sort.Sort(sort.Reverse(sort.IntSlice(totals))) // big to small
		key := keys[totals[0]]

		// rearrange the card position
		gambler.RoundCards = []*protocol.Card{}
		for _, card := range pinshis[key] {
			gambler.RoundCards = append(gambler.RoundCards, card)
			// log.Infof("best pinshi cards: %v %v", x, card)
		}
		for _, card := range dai[key] {
			gambler.RoundCards = append(gambler.RoundCards, card)
			// log.Infof("best dai cards: %v %v", x, card)
		}

		if totals[0] > 10 {
			totals[0] -= 10
		}
		if totals[0] == 10 || totals[0] == 0 {
			gambler.Combo = "niuniu"
			if r.roomType == 0 {
				gambler.Odds = 5
			}
			gambler.Rank = 10
		} else {
			gambler.Combo = "niu" + strconv.Itoa(totals[0])
			gambler.Rank = totals[0]
			if r.roomType == 0 {
				switch totals[0] {
				case 9:
					gambler.Odds = 4
				case 8:
					gambler.Odds = 3
				case 7:
					gambler.Odds = 2
				}
			}
		}

	}

}

func getDaiCombo(gambler *protocol.Gambler, combi map[int]*protocol.Card) []*protocol.Card {
	daiCombo := []*protocol.Card{}
	for _, card := range combi {
		for _, gcard := range gambler.RoundCards {
			if gcard.Suit != card.Suit || gcard.Face != card.Face {
				inCombi := false
				for _, checkcard := range combi { // not contain in pinshi
					if gcard == checkcard {
						inCombi = true
						break
					}
				}
				for _, checkcard := range daiCombo { // (not contain) not already added to daicombo
					if gcard == checkcard {
						inCombi = true
						break
					}
				}
				if inCombi {
					continue
				}
				daiCombo = append(daiCombo, gcard)
			}
		}
	}
	return daiCombo
}

func combinationUtil(roundCards []*protocol.Card, pinshis []map[int]*protocol.Card,
	data map[int]*protocol.Card, start, end, index, r int) []map[int]*protocol.Card { // where r = Size of a combination
	if index == r {
		pinshi := map[int]*protocol.Card{}
		for j, val := range data {
			pinshi[j] = val
		}
		pinshis = append(pinshis, pinshi)
		return pinshis
	}

	for i := start; i <= end && end-i+1 >= r-index; i++ {
		data[index] = roundCards[i]
		pinshis = combinationUtil(roundCards, pinshis, data, i+1, end, index+1, r)
	}

	return pinshis
}

// helper func
func randomRange(min int, max int) int {
	// return rand.Intn(max-min) + min
	return min + rand.Intn(max-min+1)
}

func randRangeFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func (r *Room) setPlayer(uid int64, p *Player) {
	if _, ok := r.players[uid]; ok {
		log.Warnf("玩家已经存在，正在覆盖玩家， UID=%d", uid)
	}
	log.Infof("room setPlayer")
	r.players[uid] = p
}

func (r *Room) sessionCount() int {
	return len(r.players)
}

func (r *Room) offline(uid int64) {
	delete(r.players, uid) // golang func
	log.Infof("ROOM 玩家: %d从在线列表中删除, 剩余：%d", uid, len(r.players))
}

func newCards() []*protocol.Card {
	var cards []*protocol.Card
	cards = append(cards, newSuit("s")...)
	cards = append(cards, newSuit("h")...)
	cards = append(cards, newSuit("c")...)
	cards = append(cards, newSuit("d")...)
	return cards
}

func newSuit(s string) []*protocol.Card {
	return []*protocol.Card{
		{
			Value: 1,
			Suit:  s,
			Face:  "A",
		},
		{
			Value: 2,
			Suit:  s,
			Face:  "2",
		},
		{
			Value: 3,
			Suit:  s,
			Face:  "3",
		},
		{
			Value: 4,
			Suit:  s,
			Face:  "4",
		},
		{
			Value: 5,
			Suit:  s,
			Face:  "5",
		},
		{
			Value: 6,
			Suit:  s,
			Face:  "6",
		},
		{
			Value: 7,
			Suit:  s,
			Face:  "7",
		},
		{
			Value: 8,
			Suit:  s,
			Face:  "8",
		},
		{
			Value: 9,
			Suit:  s,
			Face:  "9",
		},
		{
			Value: 10,
			Suit:  s,
			Face:  "10",
		},
		{
			Value: 10,
			Suit:  s,
			Face:  "J",
		},
		{
			Value: 10,
			Suit:  s,
			Face:  "Q",
		},
		{
			Value: 10,
			Suit:  s,
			Face:  "K",
		},
	}
}
