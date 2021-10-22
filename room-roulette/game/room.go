package game

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
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
		level                  string
		name                   string
		minGameCoin            int64
		maxGameCoin            int64
		minSeatCoin            int64
		maxRobotBetting        int64 // restrict robot betting
		chip                   []int64
		gameStatus             string // facilitate game
		icon                   string
		bg                     string
		betZones               protocol.BetZones
		deadline               time.Time
		players                map[int64]*Player // 所有room的玩家
		winningBetZone         protocol.CurrentBetting
		winningTextBetZone     protocol.CurrentBetting
		betItems               map[float64]protocol.CurrentBetting
		winners                map[int]string
		seats                  [6]*Player
		collection_amount      int64
		collection_betting     int64
		collection_lastupdated time.Time
		special                string // bigthree, bigfour, double
		specialChinese         string // bigthree, bigfour, double
		special_prize          float64
		randAngle              int // animation var
		randAnimals            [24]string
		randLights             [24]string
		historyList            []protocol.HistoryItem  // history
		robotPlayers           []*protocol.RobotPlayer // robot players for this room
		LogInformations        []map[string]string
		randColors             []string
		sync.RWMutex
	}
)

var offsetDec float64 = 0.45

// var randColors = []string{"Red", "Green", "Yellow"}

func (r *Room) collection_update(collected, total int64) {
	// now := db.GetCurrentShanghaiTime()
	// // if collected > 0 {
	// if now.Day() != r.collection_lastupdated.Day() {
	// 	r.collection_amount = collected
	// 	r.collection_betting = 0
	// } else {
	r.collection_amount += collected
	// }
	// }
	r.collection_betting += total
	// r.collection_lastupdated = now

	db.NewGameCollection("动物乐园", r.level, r.collection_amount, r.collection_betting)
	db.DailyCollection("动物乐园", r.level, r.collection_amount, r.collection_betting)
	// log.Println("collection_update")
	// log.Println(collected)
	// log.Println(r.collection_amount)
	// file, err := os.Create("static/" + r.name + ".txt")
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	// defer file.Close()

	// file.WriteString("服务器累计赢得数量：" +
	// 	strconv.FormatInt(r.collection_amount, 10) +
	// 	"\n回收数量：" +
	// 	strconv.FormatInt(r.collection_betting, 10) + "\n" + r.collection_lastupdated.String())
}

func (r *Room) seatsUpdate() {
	currentSeats := [6]*protocol.Player{}
	for i := range currentSeats {
		// log.Infof("player: %+v", r.seats[i])
		if r.seats[i] != nil {
			currentSeats[i] = &protocol.Player{
				FaceUri:  r.seats[i].faceUri,
				UserName: r.seats[i].userName,
				GameCoin: db.GetGameCoinByUid(r.seats[i].uid),
				Level:    r.seats[i].level,
			}
		}
	}

	r.group.Broadcast("seatsUpdate", &protocol.RoomSeatResponse{Seats: currentSeats})
}

func (r *Room) resultPhase() {
	if r.gameStatus != "animation" {
		return
	}

	var wg sync.WaitGroup

	r.gameStatus = "result"
	r.deadline = time.Now().UTC().Add(time.Second * time.Duration(15)) // 15 deadline change

	winners := []protocol.Winner{}
	for k, w := range r.winners {
		winners = append(winners, protocol.Winner{
			Username: w,
			Awarded:  k,
		})
	}

	var x = 0 // just to get the first player
	for uid := range r.players {
		wg.Add(1)
		go func(uid int64, x int) {
			updateRobotWinners := false
			if x == 0 {
				updateRobotWinners = true
			}
			if p, ok := r.players[uid]; ok {
				gamecoin := db.GetGameCoinByUid(uid) // get current gamecoin
				if p.session != nil {
					p.session.Push("resultPhase", &protocol.ResultPhase{
						AnimalBet:          p.animalBet,
						TextBet:            p.textBet,
						Awarded:            p.awarded,
						Bonus:              p.bonus,
						GameCoin:           gamecoin,
						Deadline:           r.deadline,
						Winners:            winners,
						Special:            r.special,
						UpdateRobotWinners: updateRobotWinners, // whether to send broadcast
					})
					x++
				}
				// reset
				p.animalBet = protocol.BetResult{}
				p.textBet = protocol.BetResult{}
				p.currentBettings = map[protocol.CurrentBetting]int64{}
				p.awarded = 0
				p.bonus = 0
			}
			wg.Done()
		}(uid, x)
	}

	wg.Wait()

	go r.seatsUpdate()

	for i := range r.robotPlayers {
		// reset robot players
		wg.Add(1)
		go func(rp *protocol.RobotPlayer) {
			// reset robot
			rp.CurrentBettings = map[protocol.CurrentBetting]int64{}
			rp.Awarded = 0
			if rp.GameCoin <= 1000 {
				rp = NewRobotPlayer()
			}
			wg.Done()
		}(r.robotPlayers[i])
	}
	wg.Wait()
	//if r.sessionCount() == 0 {
	//	r.gameStatus = "nogame"
	//} else {
	//go func() {
	s := r.deadline.Sub(time.Now()).Seconds()
	time.Sleep(time.Duration(s) * time.Second)
	r.bettingPhase()
	//}()
	//}
}

func randRanageFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

// func (r *Room) getProbability(betItem *protocol.BetItem, isOffset bool) float64 {
// 	probability := 1 / float64(betItem.Odds)
// 	if isOffset {
// 		probability *= offsetDec
// 	}
// 	if _, ok := r.betItems[probability]; ok {
// 		probability += math.SmallestNonzeroFloat64
// 	}
// 	return probability
// }

func (r *Room) animationPhase() {
	if r.gameStatus != "betting" {
		return
	}
	r.gameStatus = "animation"
	r.deadline = time.Now().UTC().Add(time.Second * time.Duration(18)) // 18 deadline change

	zone := r.getBetZoneByColor(r.winningBetZone.Zone) // r.winningBetZone is determined in bettingphase
	winningBetZone := zone[r.winningBetZone.Key]

	randomTextNumber := randomRange(1, 100) // TEST PLEASE REMOVE!
	// z red 43 h green 14 x yellow 43
	winningTextBetZone := r.betZones.Yellow[4]
	winningTextColor := "Yellow" // xian
	switch {
	case randomTextNumber <= 14:
		// green
		winningTextBetZone = r.betZones.Green[4]
		winningTextColor = "Green" // he
		break
	case randomTextNumber <= 57:
		// red
		winningTextBetZone = r.betZones.Red[4]
		winningTextColor = "Red" // zhuang
		break
	}

	r.winningTextBetZone = protocol.CurrentBetting{
		Zone: winningTextColor,
		Key:  4,
	}

	// save to db winningBetZone.Icon "lion" winningBetZone.Bg "red" winningTextBetZone.Icon "zhuang"
	go db.RecordWinningItem(winningBetZone.Icon, winningBetZone.Bg, winningTextBetZone.Icon, r.level) // new level
	var wg sync.WaitGroup

	r.special = ""
	r.specialChinese = ""
	randomSpecialNumber := randomRange(1, 200) // TEST PLEASE REMOVE!

	var total int64 = 0
	var collected int64 = 0
	var totalWinningBet int64 = 0

	for _, p := range r.players {
		wg.Add(1)
		go func(p *Player) {
			total += p.getAllBetsTotal()
			wg.Done()
		}(p)
	}
	wg.Wait()

	// for _, rp := range r.robotPlayers {
	//	wg.Add(1)
	//	go func(rp *protocol.RobotPlayer) {
	//		for k := range rp.CurrentBettings {
	//			total += rp.CurrentBettings[k]
	//		}
	//		wg.Done()
	//	}(rp)
	// }
	// wg.Wait()

	switch {
	case randomSpecialNumber <= 2: // 大四喜 same color
		r.special = "bigfour"
		r.specialChinese = "大四喜"
		break
	case randomSpecialNumber <= 5: // 大三元 same animal
		r.special = "bigthree"
		r.specialChinese = "大三元"
		break
	case randomSpecialNumber <= 10: // 双倍 double
		// 	//最少押注总额500万才触发彩金
		//if total >= 5000000 {
		r.special = "double"
		r.specialChinese = "双倍"
		//}
		break
	case randomSpecialNumber <= 20: // 彩金
		if r.collection_betting > 0 {
			if r.collection_amount > int64(define.MaxCollectionAmt) &&
				total >= int64(define.MaxTotalBetting) &&
				int64(float64(r.collection_amount)/float64(r.collection_betting)*100.0) > int64(define.CollectedPercentage) {
				// 	// 押注总额大于500万 AND 服务器累计当日回收数量/累计押注数量*100% 大于 8
				randomRange1 := randomRange(define.MinRangePrize, define.MaxRangePrize)
				r.special_prize = float64(randomRange1) / 100.0 * float64(r.collection_amount)

				// logger.Print("randomRange1 ", randomRange1)
				// logger.Print("r.collection_amount ", r.collection_amount)
				// logger.Print("r.special_prize ", r.special_prize)

				r.group.Broadcast("specialPrize", &protocol.SpecialPrize{
					SpecialPrize: int64(r.special_prize),
				})
				// 	// 彩金数量(special_prize)为服务器当日回收数量的百分之1~8
				// logger.Infof("Special Prize: %v", r.special_prize)
				r.special = "special_prize"
				r.specialChinese = "彩金"
			}
		}
		break
	}

	for _, p := range r.players {
		wg.Add(1)
		go func(p *Player, winningBetZone protocol.BetItem, winningTextBetZone protocol.BetItem) {
			betting, awarded, winningBet := p.calcPlayerResult(winningBetZone, winningTextBetZone)
			collected += betting - awarded
			totalWinningBet += winningBet
			wg.Done()
		}(p, winningBetZone, winningTextBetZone)
	}

	wg.Wait()

	for i := range r.robotPlayers {
		wg.Add(1)
		go func(robotPlayer *protocol.RobotPlayer, winningBetZone protocol.BetItem, winningTextBetZone protocol.BetItem) {
			r.Lock()
			_, _, winningBet := r.robotPlayerCalcResult(robotPlayer, winningBetZone, winningTextBetZone)
			//collected += betting - awarded
			totalWinningBet += winningBet
			r.Unlock()
			wg.Done()
		}(r.robotPlayers[i], winningBetZone, winningTextBetZone)
	}
	wg.Wait()

	r.collection_update(collected, total)

	// r.betZones *protocol.BetZones
	resultInfo := winningBetZone.Name + "｜" + strconv.Itoa(winningBetZone.Odds) + "，" +
		winningTextBetZone.Name + "｜" + strconv.Itoa(winningTextBetZone.Odds)
	// if r.specialChinese != "" {
	// 	resultInfo += "，" + r.specialChinese
	// }

	if r.special != "" {
		betZone := ""
		switch r.special {
		case "bigfour": // [开奖：大四喜：黄狮｜35，黄兔|4，黄猴|18，黄猫|23，闲｜2] bigfour
			resultInfo = "大四喜："
			zoneRow := r.getBetZoneByColor(r.winningBetZone.Zone)
			for i, v := range zoneRow {
				if i == 4 {
					break
				}
				betZone += fmt.Sprintf("%v｜%v", v.Name, v.Odds)
				betZone += "，"
			}
			resultInfo += betZone +
				winningTextBetZone.Name + "｜" + strconv.Itoa(winningTextBetZone.Odds)
		case "bigthree": // [开奖：大三元：黄狮｜35，红狮|25，绿狮|36，闲｜2] bigthree
			resultInfo = "大三元："
			v := r.betZones.Red[r.winningBetZone.Key]
			betZone += fmt.Sprintf("%v｜%v", v.Name, v.Odds)
			betZone += "，"
			v = r.betZones.Green[r.winningBetZone.Key]
			betZone += fmt.Sprintf("%v｜%v", v.Name, v.Odds)
			betZone += "，"
			v = r.betZones.Yellow[r.winningBetZone.Key]
			betZone += fmt.Sprintf("%v｜%v", v.Name, v.Odds)
			betZone += "，"
			resultInfo += betZone +
				winningTextBetZone.Name + "｜" + strconv.Itoa(winningTextBetZone.Odds)
		case "double": // [开奖：双倍：黄狮|35，闲｜2] double
			resultInfo = "双倍：" + resultInfo
		case "special_prize": // [开奖：彩金：12345，黄狮｜35，闲｜2] special_prize
			resultInfo = fmt.Sprintf("彩金：%v，"+resultInfo, int64(r.special_prize))

			r.LogInformations = []map[string]string{} // reset all the LogInformations
			for _, p := range r.players {
				wg.Add(1)
				go func(p *Player) {
					playercut := r.getParticipantCut(totalWinningBet, p.currentBettings)
					p.bonus = int64(playercut)
					if p.bonus > 0 {
						win_total, _ := strconv.ParseInt(p.logInformation["used"], 10, 64)
						after, _ := strconv.ParseInt(p.logInformation["after"], 10, 64)
						win_total += p.bonus
						after += p.bonus
						p.logInformation["used"] = strconv.FormatInt(win_total, 10)
						p.logInformation["after"] = strconv.FormatInt(after, 10)
						db.UpdateGameCoinByUid(p.Uid(), p.bonus)
						db.UpdateWinGameCoinByUid(p.Uid(), p.bonus)
						db.NewGameCoinTransaction(p.Uid(), p.bonus)
						p.awarded += p.bonus
					}

					p.logInformation["otherInfo"] += "[彩金：" + strconv.FormatInt(p.bonus, 10) + "]"
					// logger.Infof("Bonus: %v", p.bonus)
					r.NewLogInformation(p.logInformation)

					wg.Done()
				}(p)
			}
			wg.Wait()

			for _, rp := range r.robotPlayers {
				wg.Add(1)
				go func(rp *protocol.RobotPlayer) {
					playercut := r.getParticipantCut(totalWinningBet, rp.CurrentBettings)
					bonusCut := int64(playercut)

					if bonusCut > 0 {
						win_total, _ := strconv.ParseInt(rp.LogInformation["used"], 10, 64)
						after, _ := strconv.ParseInt(rp.LogInformation["after"], 10, 64)
						win_total += bonusCut
						after += bonusCut
						rp.LogInformation["used"] = strconv.FormatInt(win_total, 10)
						rp.LogInformation["after"] = strconv.FormatInt(after, 10)
						rp.Awarded += bonusCut
					}
					rp.LogInformation["otherInfo"] += "[彩金：" + strconv.FormatInt(bonusCut, 10) + "]"
					r.NewLogInformation(rp.LogInformation)
					rp.Awarded += bonusCut
					rp.GameCoin += bonusCut
					wg.Done()
				}(rp)
			}
			wg.Wait()

			// =============
			// 彩金的金额系统记录就显示该局总彩金数额，玩家记录显示该玩家获得的彩金数额
		}
	}

	otherInfo := "开奖：[" + resultInfo + "]"
	otherInfo += fmt.Sprintf("[总押注：%v， 总输赢：%v]", total, collected)

	logInformation := map[string]string{
		"uid":            "0",
		"game":           "动物乐园",
		"level":          r.level,
		"otherInfo":      otherInfo,
		"result":         strconv.Itoa(r.winningBetZone.Key),
		"rate":           strconv.Itoa(winningBetZone.Odds),
		"betTotal":       strconv.Itoa(int(total)),
		"winTotal":       strconv.Itoa(int(collected)),
		"bankerWinTotal": strconv.Itoa(int(collected)),
	}
	r.InsertAllLogInformations(logInformation, resultInfo)

	if len(r.historyList) >= 50 {
		r.historyList = r.historyList[:len(r.historyList)-1]
	}
	r.historyList = append([]protocol.HistoryItem{{
		Color:   winningBetZone.IconBg,
		Animal:  winningBetZone.Icon,
		Text:    winningTextBetZone.Icon,
		Special: r.special,
	},
	}, r.historyList...)

	r.group.Broadcast("animationPhase", &protocol.AnimationPhaseResponse{
		Deadline:           r.deadline,
		WinningBetZone:     winningBetZone,
		WinningTextBetZone: winningTextBetZone,
		RandAngle:          r.randAngle,
		RandAnimals:        r.randAnimals,
		RandLights:         r.randLights,
		HistoryList:        r.historyList,
	})

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

	sort.Sort(sort.Reverse(sort.IntSlice(awards)))
	r.winners = map[int]string{}
	for i := range awards {
		if len(r.winners) >= 3 {
			break
		}
		key := keys[awards[i]] // uid
		if p, ok := r.players[key]; ok && awards[i] > 0 {
			r.winners[awards[i]] = p.userName
		} else if awards[i] > 0 {
			r.winners[awards[i]] = r.robotPlayers[key].UserName
		}
	}

	// go func() {
	s := r.deadline.Sub(time.Now()).Seconds()
	time.Sleep(time.Duration(s) * time.Second)
	r.resultPhase()
	// }()
}

func setForBetting(color []protocol.BetItem) {
	for i := range color {
		color[i].Total = 0
	}
}

func randomAnimalOdds(total int, min int, max int) []int {
	index1 := randomRange(min, max)
	index2 := randomRange(min, total-max)
	index3 := total - index1 - index2
	if index3 < min {
		return randomAnimalOdds(total, min, max)
	}
	arr := []int{index1, index2, index3}
	sort.Sort(sort.Reverse(sort.IntSlice(arr))) // big to small
	return arr
}

func randomRange(min int, max int) int {
	// return rand.Intn(max-min) + min
	return min + rand.Intn(max-min+1)
}

func (r *Room) getBetZoneByColor(color string) []protocol.BetItem {
	switch color {
	case "Red":
		return r.betZones.Red
	case "Green":
		return r.betZones.Green
	default:
		return r.betZones.Yellow
	}
}

func (r *Room) setBettingOdds() {
	// reset them to zero
	setForBetting(r.betZones.Red)
	setForBetting(r.betZones.Green)
	setForBetting(r.betZones.Yellow)

	r.betZones.Red[4].Odds = 2    // z
	r.betZones.Green[4].Odds = 7  // h
	r.betZones.Yellow[4].Odds = 2 // x

	rand.Shuffle(len(r.randColors), func(i, j int) { r.randColors[i], r.randColors[j] = r.randColors[j], r.randColors[i] })

	// lion
	rand1 := randomRange(101, 109)              // max for a single animal
	randArr1 := randomAnimalOdds(rand1, 25, 52) // big to small
	// panda
	rand2 := randomRange(65, 69)                // max for a single animal
	randArr2 := randomAnimalOdds(rand2, 15, 32) // big to small
	// monkey
	rand3 := randomRange(32, 35)               // max for a single animal
	randArr3 := randomAnimalOdds(rand3, 8, 16) // big to small
	// rabit
	rand4 := 16                               // max for a single animal
	randArr4 := randomAnimalOdds(rand4, 4, 8) // big to small

	betZone1 := r.getBetZoneByColor(r.randColors[0])
	betZone1[0].Odds = randArr1[0]
	betZone1[1].Odds = randArr2[0]
	betZone1[2].Odds = randArr3[0]
	betZone1[3].Odds = randArr4[0]

	betZone2 := r.getBetZoneByColor(r.randColors[1])
	betZone2[0].Odds = randArr1[1]
	betZone2[1].Odds = randArr2[1]
	betZone2[2].Odds = randArr3[1]
	betZone2[3].Odds = randArr4[1]

	betZone3 := r.getBetZoneByColor(r.randColors[2])
	betZone3[0].Odds = randArr1[2]
	betZone3[1].Odds = randArr2[2]
	betZone3[2].Odds = randArr3[2]
	betZone3[3].Odds = randArr4[2]
}

var oddsProbReferences = [][][]*protocol.OddsProbReferenceRow{
	{
		{
			&protocol.OddsProbReferenceRow{Odds: 46, Prob: 0.0199},
			&protocol.OddsProbReferenceRow{Odds: 23, Prob: 0.0398},
			&protocol.OddsProbReferenceRow{Odds: 13, Prob: 0.0704},
			&protocol.OddsProbReferenceRow{Odds: 8, Prob: 0.1145},
		},
		{
			&protocol.OddsProbReferenceRow{Odds: 40, Prob: 0.0229},
			&protocol.OddsProbReferenceRow{Odds: 20, Prob: 0.0458},
			&protocol.OddsProbReferenceRow{Odds: 11, Prob: 0.0832},
			&protocol.OddsProbReferenceRow{Odds: 7, Prob: 0.1308},
		},
		{
			&protocol.OddsProbReferenceRow{Odds: 25, Prob: 0.0366},
			&protocol.OddsProbReferenceRow{Odds: 12, Prob: 0.0763},
			&protocol.OddsProbReferenceRow{Odds: 7, Prob: 0.1308},
			&protocol.OddsProbReferenceRow{Odds: 4, Prob: 0.2289},
		},
	},
	{
		{
			&protocol.OddsProbReferenceRow{Odds: 40, Prob: 0.0230},
			&protocol.OddsProbReferenceRow{Odds: 20, Prob: 0.0460},
			&protocol.OddsProbReferenceRow{Odds: 11, Prob: 0.0837},
			&protocol.OddsProbReferenceRow{Odds: 7, Prob: 0.1315},
		},
		{
			&protocol.OddsProbReferenceRow{Odds: 35, Prob: 0.0263},
			&protocol.OddsProbReferenceRow{Odds: 17, Prob: 0.0541},
			&protocol.OddsProbReferenceRow{Odds: 10, Prob: 0.0920},
			&protocol.OddsProbReferenceRow{Odds: 6, Prob: 0.1534},
		},
		{
			&protocol.OddsProbReferenceRow{Odds: 31, Prob: 0.0297},
			&protocol.OddsProbReferenceRow{Odds: 15, Prob: 0.0613},
			&protocol.OddsProbReferenceRow{Odds: 8, Prob: 0.1150},
			&protocol.OddsProbReferenceRow{Odds: 5, Prob: 0.1840},
		},
	},
	{
		{
			&protocol.OddsProbReferenceRow{Odds: 28, Prob: 0.0329},
			&protocol.OddsProbReferenceRow{Odds: 14, Prob: 0.0659},
			&protocol.OddsProbReferenceRow{Odds: 8, Prob: 0.1153},
			&protocol.OddsProbReferenceRow{Odds: 5, Prob: 0.1844},
		},
		{
			&protocol.OddsProbReferenceRow{Odds: 35, Prob: 0.0263},
			&protocol.OddsProbReferenceRow{Odds: 20, Prob: 0.0461},
			&protocol.OddsProbReferenceRow{Odds: 11, Prob: 0.0838},
			&protocol.OddsProbReferenceRow{Odds: 8, Prob: 0.1153},
		},
		{
			&protocol.OddsProbReferenceRow{Odds: 31, Prob: 0.0297},
			&protocol.OddsProbReferenceRow{Odds: 17, Prob: 0.0542},
			&protocol.OddsProbReferenceRow{Odds: 10, Prob: 0.0922},
			&protocol.OddsProbReferenceRow{Odds: 6, Prob: 0.1537},
		},
	},
}

func (r *Room) bettingPhase() {
	r.gameStatus = "betting"

	r.deadline = time.Now().UTC().Add(time.Second * time.Duration(27)) // 27 deadline change
	// log.Infof("BETTING DEADLINE: %d", r.deadline)
	// fmt.Println(r.deadline.String())

	// r.setBettingOdds()

	// this is always constant
	r.betZones.Red[4].Odds = 2    // z
	r.betZones.Green[4].Odds = 7  // h
	r.betZones.Yellow[4].Odds = 2 // x

	// reset totals to zero
	setForBetting(r.betZones.Red)
	setForBetting(r.betZones.Green)
	setForBetting(r.betZones.Yellow)

	r.betItems = map[float64]protocol.CurrentBetting{}
	var probabilities = []float64{}

	// 3 out of 1 oddsprob reference
	oddsProbReference := oddsProbReferences[randomRange(0, len(oddsProbReferences)-1)]

	// random decide which color is first, second or third
	rand.Shuffle(len(r.randColors), func(i, j int) { r.randColors[i], r.randColors[j] = r.randColors[j], r.randColors[i] })

	betZone1 := r.getBetZoneByColor(r.randColors[0])
	betZone2 := r.getBetZoneByColor(r.randColors[1])
	betZone3 := r.getBetZoneByColor(r.randColors[2])

	for i := 0; i < 4; i++ {
		betZone1[i].Odds = oddsProbReference[0][i].Odds
		betZone2[i].Odds = oddsProbReference[1][i].Odds
		betZone3[i].Odds = oddsProbReference[2][i].Odds

		for x := 0; x < 3; x++ {
			probability := oddsProbReference[x][i].Prob
			if len(probabilities) > 0 {
				probability += probabilities[len(probabilities)-1]
			}
			r.betItems[probability] = protocol.CurrentBetting{
				Zone: r.randColors[x],
				Key:  i,
			}
			probabilities = append(probabilities, probability)
		}

	}

	// for i := 0; i < 4; i++ { // 0 0.026 lion, 1 0.044 panda, 2 0.28 monkey, 3 0.65 rabit
	// 	fixedOffset := 0.65
	// 	if i == 0 {
	// 		fixedOffset = 0.026
	// 	} else if i == 1 {
	// 		fixedOffset = 0.044
	// 	} else if i == 2 {
	// 		fixedOffset = 0.28
	// 	}

	// 	oddsX := float64(r.betZones.Red[i].Odds)
	// 	oddsY := float64(r.betZones.Green[i].Odds)
	// 	oddsZ := float64(r.betZones.Yellow[i].Odds)

	// 	probability := fixedOffset / (oddsX * (1/oddsX + 1/oddsY + 1/oddsZ))
	// 	if len(probabilities) > 0 {
	// 		probability += probabilities[len(probabilities)-1]
	// 	}
	// 	r.betItems[probability] = protocol.CurrentBetting{
	// 		Zone: "Red",
	// 		Key:  i,
	// 	}
	// 	probabilities = append(probabilities, probability)

	// 	probability = fixedOffset / (oddsY * (1/oddsX + 1/oddsY + 1/oddsZ))
	// 	if len(probabilities) > 0 {
	// 		probability += probabilities[len(probabilities)-1]
	// 	}
	// 	r.betItems[probability] = protocol.CurrentBetting{
	// 		Zone: "Green",
	// 		Key:  i,
	// 	}
	// 	probabilities = append(probabilities, probability)

	// 	probability = fixedOffset / (oddsZ * (1/oddsX + 1/oddsY + 1/oddsZ))
	// 	if len(probabilities) > 0 {
	// 		probability += probabilities[len(probabilities)-1]
	// 	}
	// 	r.betItems[probability] = protocol.CurrentBetting{
	// 		Zone: "Yellow",
	// 		Key:  i,
	// 	}
	// 	probabilities = append(probabilities, probability)
	// } // end of for loop

	sort.Float64s(probabilities) // small to big now
	// log.Infof("Probabilities：%b", probabilities)
	randomZoneNumber := randRanageFloat(0.0, probabilities[len(probabilities)-1])
	// log.Infof("randomZoneNumber %b", randomZoneNumber)

	r.winningBetZone = r.betItems[probabilities[len(probabilities)-1]] // current betting
	for _, v := range probabilities {
		if randomZoneNumber <= v {
			r.winningBetZone = r.betItems[v]
			break
		}
	}
	// r.winningBetZone = protocol.CurrentBetting{ // TEST PLEASE REMOVE!
	// 	Zone: "Yellow",
	// 	Key:  0,
	// }

	zone := r.getBetZoneByColor(r.winningBetZone.Zone) // r.winningBetZone is determined in bettingphase
	winningBetZone := zone[r.winningBetZone.Key]       // protocol BetItem

	// r.randAngle = -1 // rand.Intn(24-1) + 1
	zoneAnimal := [4]string{"lion", "panda", "monkey", "rabit"}
	zoneColor := [3]string{"red", "green", "yellow"}
	// 0 monkey,  1 rabit, 2 lion, 3 monkey, 4 rabit, 5 monkey, 6 panda, 7 rabit, 8 monkey, 9 lion,
	// 10 panda, 11 monkey, 12 rabit
	// 13 monkey, 14 lion, 15 rabit, 16 monkey, 17 rabit, 18 panda, 19 rabit,
	// 20 monkey, 21 lion, 22 panda, 23 rabit

	in := []int{}

	for i := 0; i < 24; i++ {
		switch i {
		case 2, 9, 14, 21:
			r.randAnimals[i] = zoneAnimal[0] // lion
			break
		case 6, 10, 18, 22:
			r.randAnimals[i] = zoneAnimal[1] // panda
			break
		case 1, 4, 7, 12, 15, 17, 19, 23:
			r.randAnimals[i] = zoneAnimal[3] // rabit
			break
		case 0, 3, 5, 8, 11, 13, 16, 20:
			r.randAnimals[i] = zoneAnimal[2] // monkey
			break
		}
		if r.randAnimals[i] == winningBetZone.Icon {
			in = append(in, i)
		}
	}

	r.randAngle = in[rand.Intn(len(in))] //

	for i := 0; i < 24; i++ {
		if i == r.randAngle {
			// r.randAnimals[i] = winningBetZone.Icon
			r.randLights[i] = winningBetZone.Bg
		} else {
			// randIndex1 := rand.Intn(4)
			randIndex2 := rand.Intn(3)
			// r.randAnimals[i] = zoneAnimal[randIndex1]
			r.randLights[i] = zoneColor[randIndex2]
		}
	}

	r.group.Broadcast("bettingPhase", &protocol.BettingPhaseResponse{
		Deadline:    r.deadline,
		BetZones:    r.betZones,
		RandAnimals: r.randAnimals,
		RandLights:  r.randLights,
	})

	var wg sync.WaitGroup
	for i := range r.robotPlayers {
		var betTimes = randomRange(3, 10)
		for x := 0; x < betTimes; x++ {
			wg.Add(1)
			go func(robotPlayer *protocol.RobotPlayer) {
				s := r.deadline.Sub(time.Now()).Seconds() - 1
				n := randomRange(3000, int(s)*1000)
				time.Sleep(time.Duration(n) * time.Millisecond)
				r.robotPlayerPlaceBet(robotPlayer)
				wg.Done()
			}(r.robotPlayers[i])
		}
	}
	wg.Wait()

	// count down to next game status
	s := r.deadline.Sub(time.Now()).Seconds() + 1
	time.Sleep(time.Duration(s) * time.Second)
	r.animationPhase()
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
