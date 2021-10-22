package game

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"sync"

	// "sync"
	"time"

	log "github.com/sirupsen/logrus"

	// "gitlab.com/wolfplus/gamespace-lhd/db"
	"gitlab.com/wolfplus/gamespace-lhd/db"
	"gitlab.com/wolfplus/gamespace-lhd/protocol"
)

var totalPlayerWinnings int64 = 0
var totalPlayerTaxes int64 = 0
var totalPlayerBetting int64 = 0
var totalBankerWinnings int64 = 0

//banker details
var bankerRounds int = -1

func (m *Manager) resultPhase() {
	if m.gameStatus != "animation" {
		return
	}

	m.gameStatus = "result"
	m.deadline = time.Now().UTC().Add(time.Second * time.Duration(8)) // 8 deadline change
	bankerWinLoseAmt := totalBankerWinnings

	if int64(m.bankerAmt)+bankerWinLoseAmt <= 0 || bankerRounds >= 9 || m.bankerImg == 0 {
		// new robot banker
		m.bankerImg = randomRange(1, 10)
		m.bankerAmt = randomRange(300000000, 900000000)
		bankerRounds = 0
	} else {
		// m.bankerImg = 0 // means tell the front end dont trigger change banker
		m.bankerAmt += int(bankerWinLoseAmt)
		bankerRounds += 1
	}

	// winners := []protocol.Winner{}
	// for k, w := range m.winners {
	// 	winners = append(winners, protocol.Winner{
	// 		Username: w,
	// 		Awarded:  k,
	// 	})
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
			if p, ok := players[uid]; ok {
				gamecoin := db.GetGameCoinByUid(uid) // get current gamecoin
				if p.session != nil {
					p.session.Push("resultPhase", &protocol.ResultPhase{
						PersonalBetResult:  p.betResult,
						Awarded:            p.awarded,
						GameCoin:           gamecoin,
						Deadline:           m.deadline,
						Winners:            m.winners,
						BankerWinLoseAmt:   bankerWinLoseAmt,
						WinningBetZone:     m.winningBetZone,
						BankerImg:          m.bankerImg,
						BankerAmt:          m.bankerAmt,
						BankerRounds:       bankerRounds,
						UpdateRobotWinners: updateRobotWinners,
					})
				}
				// reset
				p.betResult = protocol.BetResult{}
				p.currentBettings = map[CurrentBetting]int64{}
				p.awarded = 0
			}
			wg.Done()
		}(uid, x)
		x++
	}
	wg.Wait()

	for i := range RoomRobotPlayers {
		wg.Add(1)
		// reset robot players
		go func(rp *RobotPlayer) {
			if rp.gameCoin <= 1000 {
				rp = newRobotPlayer()
			} else {
				rp.currentBettings = map[CurrentBetting]int64{}
				rp.awarded = 0
			}
			wg.Done()
		}(RoomRobotPlayers[i])
	}
	wg.Wait()

	//if m.sessionCount() == 0 {
	//	m.gameStatus = "nogame"
	//} else {
	//	go func() {
	s := m.deadline.Sub(time.Now()).Seconds() + 1
	time.Sleep(time.Duration(s) * time.Second)
	m.bettingPhase()
	//	}()
	//}
}

func randRanageFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func (m *Manager) animationPhase() {
	if m.gameStatus != "betting" {
		return
	}

	rand.Seed(time.Now().UTC().UnixNano())

	m.gameStatus = "animation"
	m.deadline = time.Now().UTC().Add(time.Second * time.Duration(12)) // 18 deadline change

	maxProbability := 0.0
	var probabilityRange = []float64{}
	var probabilities = []float64{
		0.024793388, // 0
		0.033057851, // 1
		0.049586777, // 2
		0.099173554, // 3
		0.198347107, // 4
		0.198347107, // 5
		0.198347107, // 6
		0.198347107, // 7
	}

	// sort.Float64s(probabilities) // small to big now

	for i := 0; i < len(probabilities); i++ {
		maxProbability += probabilities[i]
		probabilityRange = append(probabilityRange, maxProbability)
	}
	randProbability := randRanageFloat(0.0, maxProbability)
	// log.Infof("probabilityRange: %v", probabilityRange)
	// log.Infof("randProbability: %v", randProbability)

	winningKey := 7
	winningProbability := probabilities[7]    // should have been commented out by wk.
	m.winningBetZone = m.betZones[winningKey] //m.betZones[randomRange(0, 7)]
	for i, v := range probabilityRange {
		if randProbability <= v {
			m.winningBetZone = m.betZones[i]
			winningKey = i
			winningProbability = probabilities[i] //  wk
			break
		}
	}
	// m.winningBetZone = m.betZones[0] // TEST PLEASE REMOVE

	// add to db winningBetZone
	go db.RecordWinningItem(winningKey)

	// TODO add to history
	// m.historyList
	if len(m.historyList) >= 50 {
		m.historyList = m.historyList[:len(m.historyList)-1]
	}
	m.historyList = append([]*protocol.BetItem{m.winningBetZone}, m.historyList...)

	randAngle := 0
	in := []int{}
	for i := 0; i < len(m.winningItems); i++ {
		if m.winningItems[i].NickName == m.winningBetZone.NickName {
			in = append(in, i)
		}
	}
	randAngle = in[rand.Intn(len(in))]

	m.group.Broadcast("animationPhase", &protocol.AnimationPhaseResponse{
		Deadline:       m.deadline,
		WinningBetZone: m.winningBetZone,
		RandAngle:      randAngle,
	})

	var wg sync.WaitGroup
	totalPlayerWinnings = 0
	totalPlayerTaxes = 0
	totalPlayerBetting = 0
	totalBankerWinnings = 0

	players := m.getPlayers()

	for _, p := range players {
		wg.Add(1)
		go func(p *Player, winningBetZone *protocol.BetItem) {
			m.Lock()
			defer m.Unlock()
			playerWinnings, playerTax, totalBetting, bankerWinning := p.calcPlayerResult(winningBetZone)
			totalPlayerWinnings += playerWinnings
			if playerWinnings > 0 {
				totalPlayerTaxes += playerTax
			}
			totalPlayerBetting += totalBetting
			totalBankerWinnings += bankerWinning
			wg.Done()
		}(p, m.winningBetZone)
	}

	wg.Wait()

	for i := range RoomRobotPlayers {
		wg.Add(1)
		go func(robotPlayer *RobotPlayer, winningBetZone *protocol.BetItem) {
			m.Lock()
			defer m.Unlock()
			playerWinnings, playerTax, totalBetting, bankerWinning := m.robotPlayerCalcResult(robotPlayer, winningBetZone)
			totalPlayerWinnings += playerWinnings
			if playerWinnings > 0 {
				//	totalPlayerTaxes += playerTax
				fmt.Println(playerTax)
			}
			totalPlayerBetting += totalBetting
			totalBankerWinnings += bankerWinning
			wg.Done()
		}(RoomRobotPlayers[i], m.winningBetZone)
	}

	wg.Wait()

	// save to db for the day totalPlayerTaxes
	// 开奖：[红兔｜7，和｜7] [总押注：230， 总输赢：-97] [参数：43126]
	// [开奖：" + winningBetZone.Name + "｜" + strconv.Itoa(winningBetZone.Odds) + "]
	prizes := fmt.Sprintf("%v｜%v",
		m.winningBetZone.Name, m.winningBetZone.Odds)

	otherInfo := fmt.Sprintf("开奖：["+prizes+"][总押注：%v， 总输赢：%v, 税: %v]", totalPlayerBetting, totalBankerWinnings, totalPlayerTaxes)
	if totalPlayerTaxes > 0 {
		go db.CollectTax(totalPlayerTaxes)
	}

	logInformation := map[string]string{
		"uid":            "0",
		"game":           "豪车汇",
		"result":         strconv.Itoa(winningKey),
		"rate":           strconv.FormatFloat(winningProbability, 'E', -1, 64),
		"betTotal":       strconv.Itoa(int(totalPlayerBetting)),
		"winTotal":       strconv.Itoa(int(totalBankerWinnings)),
		"bankerWinTotal": strconv.Itoa(int(-totalPlayerWinnings)),
		"otherInfo":      otherInfo,
		"tax":            strconv.Itoa(int(totalPlayerTaxes)),
	}
	db.InsertAllLogInformations(logInformation, prizes)

	// sort by the winnings of each players
	var awards = []int{}
	keys := map[int]int64{}
	for k, p := range players {
		keys[int(p.awarded)] = k
		awards = append(awards, int(p.awarded))
	}
	for i, rp := range RoomRobotPlayers { // robot players added to sort
		keys[int(rp.awarded)] = int64(i) // this is not a real uid
		awards = append(awards, int(rp.awarded))
	}
	sort.Sort(sort.Reverse(sort.IntSlice(awards))) // big to small
	// populate winners with the sorted result
	// m.winners = map[int]string{}
	m.winners = []protocol.Winner{}
	for i := range awards {
		key := keys[awards[i]] // uid
		if p, ok := players[key]; ok && awards[i] > 0 {
			// m.winners[awards[i]] = p.userName
			m.winners = append(m.winners, protocol.Winner{
				Username: p.userName,
				Awarded:  awards[i],
				IsRobot:  false,
			})
		} else if awards[i] > 0 {
			// m.winners[awards[i]] = RoomRobotPlayers[key].userName
			m.winners = append(m.winners, protocol.Winner{
				Username: RoomRobotPlayers[key].userName,
				Awarded:  awards[i],
				IsRobot:  true,
			})
		}
	}

	// go func() {
	s := m.deadline.Sub(time.Now()).Seconds() + 1
	time.Sleep(time.Duration(s) * time.Second)
	m.resultPhase()
	// }()
}

func randomRange(min int, max int) int {
	// return rand.Intn(max-min) + min
	return min + rand.Intn(max-min+1)
}

func (m *Manager) bettingPhase() {
	m.gameStatus = "betting"

	// reset to zero
	for i := range m.betZones {
		m.betZones[i].Total = 0
	}

	m.deadline = time.Now().UTC().Add(time.Second * time.Duration(20)) // 30 deadline change
	log.Infof("BETTING DEADLINE: %d", m.deadline)
	fmt.Println(m.deadline.String())

	m.group.Broadcast("bettingPhase", &protocol.BettingPhaseResponse{
		Deadline: m.deadline,
	})

	var wg sync.WaitGroup
	for i := range RoomRobotPlayers {

		var betTimes = randomRange(3, 10)
		for x := 0; x < betTimes; x++ {
			wg.Add(1)
			go func(robotPlayer *RobotPlayer) {
				s := m.deadline.Sub(time.Now()).Seconds() - 1
				n := randomRange(3000, int(s)*1000)
				time.Sleep(time.Duration(n) * time.Millisecond)
				m.robotPlayerPlaceBet(robotPlayer)
				wg.Done()
			}(RoomRobotPlayers[i])
		}
	}
	wg.Wait()

	// count down to next game status
	go func() {
		s := m.deadline.Sub(time.Now()).Seconds() + 1
		time.Sleep(time.Duration(s) * time.Second)
		m.animationPhase()
	}()
}
