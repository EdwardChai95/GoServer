package game

import (
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lonng/nano"
	"gitlab.com/wolfplus/gamespace-lhd/db"
	"gitlab.com/wolfplus/gamespace-lhd/pokersolver"
	"gitlab.com/wolfplus/gamespace-lhd/protocol"
)

type (
	Table struct {
		gameStatus    string
		deadline      time.Time                      // when to go to the next phase
		seatedPlayers [NUMBER_OF_SEATS]*SeatedPlayer // room 的所有入座玩家
		players       map[int64]*Player              // room 的所有玩家
		cards         []protocol.Card                // 场次使用的扑克牌
		publicCards   []protocol.Card                // 公牌
		prizePool     int64                          // 底池
		bettingAmount int64
		myTurnChan    chan bool // 同步玩家押注
		// limitedConfig
		SmallRaise int64 // 前2轮加注金额
		BigRaise   int64 // 后2轮加注金额
		IsLimited  bool  // 是否有限
		// to config:
		Title           string // 场次名字 e.g. 初级场
		GameCoinToEnter int64  // 准入金额
		SmallBlind      int64  // 小盲注金额
		BigBlind        int64  // 大盲注金额
		group           *nano.Group
		manager         *Manager
		// robot
		robotPlayers []*protocol.RobotPlayer // robot players for this table
	}
)

// room phases cycle function

func (t *Table) showdownPhase() {
	// var earlyShowdown bool = false
	// if t.gameStatus != GAMESTATUS_RIVER {
	// 	earlyShowdown = true
	// }

	t.gameStatus = GAMESTATUS_SHOWDOWN

	// var highestHands [][]*protocol.Card
	// var ranks []int
	// var seatNumbers []int
	var playerHands []pokersolver.PlayerHand = nil
	var foldTotalBetAmount int64 = 0
	var betTotal int64 = t.prizePool // for logging
	var bankerWinTotal int64 = 0     // for logging
	var taxTotal int64 = 0           // for logging

	var wg sync.WaitGroup
	for _, p := range t.seatedPlayers {
		if p != nil && p.IsInGame {
			_, seatNumber := t.getSeatedPlayerByUid(p.Uid)
			if p.HasFolded {
				foldTotalBetAmount += p.TotalBetAmount
			} else {
				wg.Add(1)
				go func(p *SeatedPlayer, seatNumber int) {
					// if player, ok := r.players[p.Uid]; ok {
					// 	if player.session != nil {

					highestHand, rank := pokersolver.GetHighestHand(p.cards, t.publicCards)
					// highestHands = append(highestHands, highestHand)
					// ranks = append(ranks, rank)
					// seatNumbers = append(seatNumbers, seatNumber)
					playerHands = append(playerHands, pokersolver.PlayerHand{
						Hand:       highestHand,
						Rank:       rank,
						SeatNumber: seatNumber,
					})
					p.CardRank = rank
					p.CardRankTitle = pokersolver.RANK_TITLE[rank]
					p.ShowCards = []protocol.Card{
						p.cards[0],
						p.cards[1],
					}
					p.HighestHand = highestHand
					// 	}
					// }
					wg.Done()
				}(p, seatNumber)
			}
		}
	}
	wg.Wait()

	var showdownWinners []protocol.ShowdownWinners = nil

	for {
		if len(playerHands) == 0 || t.prizePool <= 0 {
			break
		}

		winners := pokersolver.CompareHands(playerHands) // only get seatnumbers of winners

		tmpPlayerHands := make([]pokersolver.PlayerHand, len(playerHands))
		copy(tmpPlayerHands, playerHands)

		for _, seatNumber := range winners {
			showdownWinner := t.seatedPlayers[seatNumber]

			// logger.Printf("showdownWinner.TotalBetAmount: %v", showdownWinner.TotalBetAmount)
			// logger.Printf("int64(len(playerHands): %v", int64(len(playerHands)))
			// logger.Printf("int64(len(winners): %v", int64(len(winners)))

			// all in situation
			winnerWinAmount := ((foldTotalBetAmount/int64(len(playerHands)) +
				showdownWinner.TotalBetAmount) * int64(len(playerHands))) / int64(len(winners))
			// logger.Printf("winnerWinAmount: %v", winnerWinAmount)
			// logger.Printf("t.prizePool: %v", t.prizePool)
			if winnerWinAmount > t.prizePool {
				winnerWinAmount = t.prizePool
			}

			showdownWinner.UseableGameCoin += winnerWinAmount
			t.prizePool -= winnerWinAmount
			// TO DO UPDATE USER GAME COIN

			showdownWinners = append(showdownWinners, protocol.ShowdownWinners{
				SeatNumber: seatNumber,
				WinAmount:  winnerWinAmount,
			})

			showdownWinner.WinAmount = winnerWinAmount // for logging

			for i, tmpPlayerHand := range tmpPlayerHands {
				if tmpPlayerHand.SeatNumber == seatNumber {
					tmpPlayerHands = append(tmpPlayerHands[:i], tmpPlayerHands[i+1:]...)
					break
				}
			}
		}
		playerHands = tmpPlayerHands
	}

	// log can update game coin
	var log_publicCards string = "["
	for _, c := range t.publicCards {
		log_publicCards += c.Face
	}
	log_publicCards += "]"
	// logger.Println(log_publicCards)

	for _, p := range t.seatedPlayers {
		if p != nil && p.IsInGame {
			log_playerCards := "["
			for _, c := range p.cards {
				log_playerCards += c.Face
			}
			log_playerCards += "]"
			// logger.Println(log_playerCards)

			log_playerRank := "[" + db.Int64ToString(int64(p.CardRank)) + "]"

			log_phaseBetting := "["
			if val, ok := p.PhaseBetting[GAMESTATUS_PREFLOP]; ok {
				log_phaseBetting += db.Int64ToString(val) + "："
			} else {
				log_phaseBetting += "0" + "："
			}
			if val, ok := p.PhaseBetting[GAMESTATUS_FLOP]; ok {
				log_phaseBetting += db.Int64ToString(val) + "："
			} else {
				log_phaseBetting += "0" + "："
			}
			if val, ok := p.PhaseBetting[GAMESTATUS_TURN]; ok {
				log_phaseBetting += db.Int64ToString(val) + "："
			} else {
				log_phaseBetting += "0" + "："
			}
			if val, ok := p.PhaseBetting[GAMESTATUS_RIVER]; ok {
				log_phaseBetting += db.Int64ToString(val)
			} else {
				log_phaseBetting += "0"
			}
			log_phaseBetting += "]"

			playerlog := log_publicCards + " " + log_playerCards + " " + log_playerRank + log_phaseBetting
			// logger.Println(playerlog)

			currentPlayerGameCoin := db.GetGameCoinByUid(p.Uid)

			var used int64 = -p.TotalBetAmount
			var tax int64 = 0
			if p.WinAmount > 0 {
				tax = int64(float64(p.WinAmount-p.TotalBetAmount) * 0.05)
				used = int64(float64(p.WinAmount-p.TotalBetAmount) * 0.95)
			}

			if tax > 0 {
				playerlog += " [税收：" + db.Int64ToString(tax) + "]"
				taxTotal += tax
			}

			bankerWinTotal -= used

			logInformation := map[string]string{
				"uid":        db.Int64ToString(p.Uid),
				"game":       LOG_GAMETITLE,
				"level":      t.Title,
				"other_info": playerlog,
				"bet_total":  db.Int64ToString(p.TotalBetAmount),
				"win_total":  db.Int64ToString(p.WinAmount),
				"before":     db.Int64ToString(currentPlayerGameCoin),
				"after":      db.Int64ToString(currentPlayerGameCoin + used),
				"used":       db.Int64ToString(used),
				"tax":        db.Int64ToString(tax),
			}
			go db.NewLogInformation(logInformation)
			go db.UpdateGameCoinByUid(p.Uid, used, LOG_GAMETITLE)
			p.OverallGameCoin = db.GetGameCoinByUid(p.Uid)
		}
	} // end of for seated players

	if taxTotal > 0 {
		log_publicCards += " [税收：" + db.Int64ToString(taxTotal) + "]"
	}

	logInformation := map[string]string{
		"uid":              "0",
		"game":             LOG_GAMETITLE,
		"level":            t.Title,
		"other_info":       log_publicCards,
		"bet_total":        strconv.Itoa(int(betTotal)),
		"win_total":        strconv.Itoa(int(bankerWinTotal)),
		"banker_win_total": strconv.Itoa(int(bankerWinTotal)),
		"tax":              db.Int64ToString(taxTotal),
	}
	go db.InsertAllLogInformations(logInformation)

	// logger.Printf("earlyShowdown: %v", earlyShowdown)

	// if earlyShowdown {
	// 	t.deadline = time.Now().UTC().Add(time.Second *
	// 		time.Duration(10))
	// } else {
	// t.deadline = time.Now().UTC().Add(time.Second *
	// 	time.Duration(len(showdownWinners)*ANIMATION_TIME_SHOWDOWN))
	t.deadline = time.Now().UTC().Add(time.Second *
		time.Duration(ANIMATION_TIME_SHOWDOWN))
	// len(showdownWinners)*ANIMATION_TIME_SHOWDOWN
	// logger.Println(len(showdownWinners) * ANIMATION_TIME_SHOWDOWN)
	// }

	t.seatUpdateBroadcast()

	// prizepool is betTotal
	t.group.Broadcast(GAMESTATUS_SHOWDOWN, &protocol.ShowdownBroadcast{
		Winners:             showdownWinners,
		PublicCards:         t.publicCards,
		Deadline:            t.deadline,
		SeatedPlayersUpdate: t.seatedPlayers,
		PrizePool:           betTotal,
	})

	sleeptime := t.deadline.Sub(time.Now().UTC()).Seconds()
	time.Sleep(time.Duration(sleeptime+1) * time.Second) // sleep 1 more second to prevent overlap

	logger.Println(GAMESTATUS_SHOWDOWN + " animation end")
	go t.waitingPhase()
}

func (t *Table) riverPhase() {
	if t.gameStatus != GAMESTATUS_TURN {
		return
	}

	t.bettingAmount = 0
	t.gameStatus = GAMESTATUS_RIVER
	t.deadline = time.Now().UTC().Add(time.Second * time.Duration(ANIMATION_TIME_1CARD))

	t.seatUpdateBroadcast()

	t.group.Broadcast(GAMESTATUS_RIVER, &protocol.RiverBroadcast{
		PublicCards: t.publicCards,
		Deadline:    t.deadline,
	})

	sleeptime := t.deadline.Sub(time.Now().UTC()).Seconds()
	time.Sleep(time.Duration(sleeptime+1) * time.Second) // sleep 1 more second to prevent overlap

	logger.Println("river animation end")

	// numOfInGamePlayers := r.numOfInGamePlayers()
	bankerIndex := 0
	for i, p := range t.seatedPlayers {
		if t.seatedPlayerIsInGame(i) {
			if p.IsBanker {
				bankerIndex = i
			}
		}
	}
	currentPlayer, currentSeat := t.nextPlayableSeatedPlayer(bankerIndex)

	t.playersTurnSequence(currentPlayer, currentSeat)

	go t.showdownPhase()
}

func (t *Table) turnPhase() {
	if t.gameStatus != GAMESTATUS_FLOP {
		return
	}

	t.bettingAmount = 0
	t.gameStatus = GAMESTATUS_TURN
	t.deadline = time.Now().UTC().Add(time.Second * time.Duration(ANIMATION_TIME_1CARD))

	t.seatUpdateBroadcast()

	t.group.Broadcast(GAMESTATUS_TURN, &protocol.TurnBroadcast{
		PublicCards: []protocol.Card{
			t.publicCards[0],
			t.publicCards[1],
			t.publicCards[2],
			t.publicCards[3],
		},
		Deadline: t.deadline,
	})

	sleeptime := t.deadline.Sub(time.Now().UTC()).Seconds()
	time.Sleep(time.Duration(sleeptime+1) * time.Second) // sleep 1 more second to prevent overlap

	logger.Println("flop animation end")

	// numOfInGamePlayers := r.numOfInGamePlayers()
	bankerIndex := 0
	for i, p := range t.seatedPlayers {
		if t.seatedPlayerIsInGame(i) {
			if p.IsBanker {
				bankerIndex = i
			}
		}
	}
	currentPlayer, currentSeat := t.nextPlayableSeatedPlayer(bankerIndex)

	if t.playersTurnSequence(currentPlayer, currentSeat) < 2 {
		t.seatUpdateBroadcast()
		go t.showdownPhase()
	} else {
		go t.riverPhase()
	}

}

func (t *Table) flopPhase() {
	if t.gameStatus != GAMESTATUS_PREFLOP {
		return
	}

	t.bettingAmount = 0
	t.gameStatus = GAMESTATUS_FLOP
	t.deadline = time.Now().UTC().Add(time.Second * time.Duration(ANIMATION_TIME))

	t.seatUpdateBroadcast()

	t.group.Broadcast(GAMESTATUS_FLOP, &protocol.FlopBroadcast{
		PublicCards: []protocol.Card{
			t.publicCards[0],
			t.publicCards[1],
			t.publicCards[2],
		},
		Deadline: t.deadline,
	})

	sleeptime := t.deadline.Sub(time.Now().UTC()).Seconds()
	time.Sleep(time.Duration(sleeptime+1) * time.Second) // sleep 1 more second to prevent overlap

	logger.Println("flop animation end")

	// numOfInGamePlayers := r.numOfInGamePlayers()
	bankerIndex := 0
	for i, p := range t.seatedPlayers {
		if t.seatedPlayerIsInGame(i) {
			if p.IsBanker {
				bankerIndex = i
			}
		}
	}
	currentPlayer, currentSeat := t.nextPlayableSeatedPlayer(bankerIndex)

	if t.playersTurnSequence(currentPlayer, currentSeat) < 2 {
		t.seatUpdateBroadcast()
		go t.showdownPhase()
	} else {
		go t.turnPhase()
	}

}

func (t *Table) preflopPhase() {
	if t.gameStatus != GAMESTATUS_WAITING {
		return
	}

	// logger.Println("preflopPhase start")

	for i, p := range t.seatedPlayers {
		// if p != nil && (p.UseableGameCoin <= t.BigBlind || p.session == nil) {
		// 	t.seatedPlayers[i] = nil
		// }
		if p != nil { // 判断玩家
			if t.seatedPlayers[i].Uid > 10 &&
				(p.UseableGameCoin <= t.BigBlind || p.session == nil) {
				t.seatedPlayers[i] = nil
			}
		}
	}

	// setup table for a new round
	t.prizePool = 0                   // reset
	t.publicCards = []protocol.Card{} // reset

	t.seatUpdateBroadcast() // clean up losers

	t.gameStatus = GAMESTATUS_PREFLOP
	numOfSeatedPlayers := t.numOfSeatedPlayers()
	logger.Printf("numOfSeatedPlayers: %v", numOfSeatedPlayers)
	if numOfSeatedPlayers < NUMBER_OF_MINIMUM_PLAYERS {
		if numOfSeatedPlayers == 0 {
			t.gameStatus = GAMESTATUS_NOGAME
			return
		}
		go t.waitingPhase()
		return
	}

	prevBankerIndex := randomRange(0, len(t.seatedPlayers)-1)
	for i, p := range t.seatedPlayers {
		if p != nil {
			p.IsInGame = true
			if p.IsBanker {
				p.IsBanker = false
				prevBankerIndex = i
			}
		}
	}

	newBanker, newBankerSeatNumber := t.nextPlayableSeatedPlayer(prevBankerIndex)
	smallBlind, smallBlindSeatNumber := t.nextPlayableSeatedPlayer(newBankerSeatNumber)
	bigBlind, bigBlindSeatNumber := t.nextPlayableSeatedPlayer(smallBlindSeatNumber)
	starter, starterSeatNumber := t.nextPlayableSeatedPlayer(bigBlindSeatNumber)

	newBanker.IsBanker = true
	smallBlind.BetAmount = t.SmallBlind      // determine bet amount for the turn
	smallBlind.TotalBetAmount = t.SmallBlind // update total bet amount
	smallBlind.UseableGameCoin -= t.SmallBlind
	smallBlind.Status = 0
	smallBlind.PhaseBetting[t.gameStatus] = t.SmallBlind // for logging

	bigBlind.BetAmount = t.BigBlind      // determine bet amount for the turn
	bigBlind.TotalBetAmount = t.BigBlind // update total bet amount
	bigBlind.UseableGameCoin -= t.BigBlind
	bigBlind.Status = 1
	bigBlind.PhaseBetting[t.gameStatus] = t.BigBlind // for logging

	t.bettingAmount = t.BigBlind

	currentPlayer := newBanker // the person to be dealt cards to first
	currentSeat := newBankerSeatNumber

	// ***** TEST 测试 MODE
	// var gongPai = []string{"cA", "c2", "s6", "d6", "c8"}
	// var wanjiaPai = [][]string{
	// 	{"c4", "hK"}, // 0
	// 	{"c6", "s9"}, // 1 w
	// 	{"d6", "h3"}, // 2 w
	// 	{"c5", "s5"}, // 3
	// 	{"h6", "s3"}, // 4 w
	// 	{"c7", "s7"}, // 5
	// }

	// var publicCards = []protocol.Card{}
	// for _, p := range gongPai {
	// 	publicCards = append(publicCards, _getCard(p))
	// }
	// t.publicCards = publicCards
	// for seatNumber, wp := range wanjiaPai {
	// 	if t.seatedPlayers[seatNumber] != nil {
	// 		var cards = []protocol.Card{}
	// 		for _, p := range wp {
	// 			cards = append(cards, _getCard(p))
	// 		}
	// 		t.seatedPlayers[seatNumber].cards = cards
	// 	}
	// }
	// ***** TEST 测试 MODE

	randomCardsIndexes := rand.Perm(len(t.cards)) // shuffle

	for i, index := range randomCardsIndexes {
		if numOfSeatedPlayers*2 <= i { // means the players have all the cards
			if len(t.publicCards) == 5 {
				break
			}
			t.publicCards = append(t.publicCards, t.cards[index])
			continue
		}
		if len(currentPlayer.cards) == 2 {
			currentPlayer, currentSeat = t.nextPlayableSeatedPlayer(currentSeat)
		}
		currentPlayer.cards = append(currentPlayer.cards, t.cards[index]) // give each player 2 cards
	}
	// setup done

	// first round is a individual push because we dont want them to see each others cards
	t.deadline = time.Now().UTC().Add(time.Second * time.Duration(ANIMATION_PLAYER1CARD*numOfSeatedPlayers))

	var wg sync.WaitGroup

	for _, p := range t.seatedPlayers {
		if p != nil {
			wg.Add(1)
			go func(p *SeatedPlayer,
				seatedPlayersUpdate [6]*SeatedPlayer, deadline time.Time) {
				if p.session != nil {
					p.session.Push(GAMESTATUS_PREFLOP, &protocol.PreflopBroadcast{
						SeatedPlayersUpdate: seatedPlayersUpdate,
						MyCards:             p.cards,
						Deadline:            deadline,
					})
				}
				wg.Done()
			}(p, t.seatedPlayers, t.deadline)
		}
	}
	wg.Wait()

	// for the spectators
	t.group.Broadcast(GAMESTATUS_PREFLOP, &protocol.PreflopBroadcast{
		SeatedPlayersUpdate: t.seatedPlayers,
		Deadline:            t.deadline,
	})

	sleeptime := t.deadline.Sub(time.Now().UTC()).Seconds()
	time.Sleep(time.Duration(sleeptime+1) * time.Second) // sleep 1 more second to prevent overlap

	logger.Println("preflopPhase animation end")

	t.prizePool = t.SmallBlind + t.BigBlind

	currentPlayer = starter
	currentSeat = starterSeatNumber

	if t.playersTurnSequence(currentPlayer, currentSeat) < 2 {
		t.seatUpdateBroadcast()
		go t.showdownPhase()
	} else {
		go t.flopPhase()
	}
}

func (t *Table) waitingPhase() {
	// no deadline because waiting for other players to connect and not a countdown
	if t.gameStatus != GAMESTATUS_NOGAME && t.gameStatus != GAMESTATUS_SHOWDOWN &&
		t.gameStatus != GAMESTATUS_PREFLOP {
		return
	}
	t.gameStatus = GAMESTATUS_WAITING
	t.prizePool = 0
	logger.Println("waiting phase now")

	for _, p := range t.seatedPlayers {
		if p != nil {
			p.IsInGame = false  // reset
			p.HasAllIn = false  // reset
			p.HasFolded = false // reset
			p.IsMyTurn = false  // reset
			// p.IsBanker = false // reset
			p.BetAmount = 0                   // reset
			p.TotalBetAmount = 0              // reset
			p.WinAmount = 0                   // reset
			p.cards = []protocol.Card{}       // reset
			p.ShowCards = []protocol.Card{}   // reset
			p.HighestHand = []protocol.Card{} // reset
			p.Status = 2                      // set to waiting
			p.CardRank = 0                    // reset
			p.CardRankTitle = ""
			p.PhaseBetting = map[string]int64{} // reset
		}
	}

	t.group.Broadcast(GAMESTATUS_WAITING, &protocol.WaitingBroadcast{
		SeatedPlayersUpdate: t.seatedPlayers,
	})

	x := 0
	for i, p := range t.seatedPlayers { //判断玩家是否还在
		if p != nil {
			if t.seatedPlayers[i].Uid > 10 {
				x = 0 // 有玩家
			} else {
				x = 1 // 没有玩家
			}
		}
	}

	numOfSeatedPlayers := t.numOfSeatedPlayers()
	if numOfSeatedPlayers >= NUMBER_OF_MINIMUM_PLAYERS {
		go t.preflopPhase()
	} else if numOfSeatedPlayers < NUMBER_OF_MINIMUM_PLAYERS &&
		numOfSeatedPlayers != 0 && x == 0 { // 添加机器人
		for i := range t.seatedPlayers {
			time.Sleep(time.Duration(4) * time.Second) // 机器人加入时间
			num := t.numOfSeatedPlayers()
			if t.seatedPlayers[i] == nil && num < NUMBER_OF_MINIMUM_PLAYERS {
				t.robotPlayerTakeSeat(i) // 机器人入座
			}
		}
	} else if x == 1 { // 玩家退出，机器人跟着退出
		for i, p := range t.seatedPlayers {
			if p != nil {
				t.seatedPlayers[i] = nil
				numseat := t.numOfSeatedPlayers()
				if numseat == 0 {
					t.gameStatus = GAMESTATUS_NOGAME
					return
				}
			}
		}
	}

	// for {
	// 	time.Sleep(time.Duration(2) * time.Second) // sleep 2 second check every second

	// 	numOfSeatedPlayers := t.numOfSeatedPlayers()

	// 	// if numOfSeatedPlayers < NUMBER_OF_MINIMUM_PLAYERS && numOfSeatedPlayers != 0 {
	// 	// here means nobody is in the room or the game yet
	// 	// init the room
	// 	// go t.waitingPhase()
	// 	// } else
	// 	if numOfSeatedPlayers >= NUMBER_OF_MINIMUM_PLAYERS {
	// 		go t.preflopPhase()
	// 		break
	// 	} else {
	// 		// t.seatUpdateBroadcast() // this broadcast is just to clean up the empty seats
	// 	}
	// }
}

// business logic
func (t *Table) TakeSeat(uid int64, req *protocol.TakeSeatRequest) (bool, string) {
	if p, ok := t.players[uid]; ok {
		gamecoin := db.GetGameCoinByUid(uid)

		if t.seatedPlayers[req.SeatNumber] == nil &&
			gamecoin >= req.UseableGameCoin &&
			req.UseableGameCoin >= t.GameCoinToEnter {
			for _, checkp := range t.seatedPlayers {
				if checkp != nil && checkp.Uid == p.uid && checkp.session != nil {
					return false, "already seated"
				}
			}

			t.seatedPlayers[req.SeatNumber] = &SeatedPlayer{
				Uid:             p.uid,
				FaceUri:         p.faceUri,
				UserName:        p.userName,
				session:         p.session,
				OverallGameCoin: gamecoin,
				UseableGameCoin: req.UseableGameCoin,
				BetAmount:       0,     // def
				TotalBetAmount:  0,     // def
				CardRank:        0,     // default
				IsInGame:        false, // default
				HasAllIn:        false,
				HasFolded:       false, // default
				IsMyTurn:        false, // default
				IsBanker:        false, // default
				HasPlayed:       false, // default
				Status:          2,
				cards:           []protocol.Card{},
				ShowCards:       []protocol.Card{},
				HighestHand:     []protocol.Card{},
				PhaseBetting:    map[string]int64{}, // default
			}
			t.seatUpdateBroadcast()

			//todo
			if t.gameStatus == GAMESTATUS_WAITING {
				numOfSeatedPlayers := t.numOfSeatedPlayers()
				if numOfSeatedPlayers >= NUMBER_OF_MINIMUM_PLAYERS {
					go t.preflopPhase()

				}
			}

			return true, "成功"
		} else {
			return false, "already taken" // unavailable
		}
	}
	return false, ""
}

// helpers:

func _getCard(strPai string) protocol.Card { // TEST USE ONLY
	var str = strings.Split(strPai, "")
	var suit = str[0]
	var face = str[1]
	value, err := strconv.Atoi(face)
	if err != nil {
		if face == "A" {
			value = 14
		} else if face == "K" {
			value = 13
		} else if face == "Q" {
			value = 12
		} else if face == "J" {
			value = 11
		}
	} else if face == "1" {
		value = 10
		face = "10"
	}

	return protocol.Card{
		Value: value,
		Suit:  suit,
		Face:  face,
	}
}

func (t *Table) playersTurnSequence(currentPlayer *SeatedPlayer, currentSeat int) int {
	if numOfPlayablePlayers := t.numOfPlayablePlayers(); numOfPlayablePlayers < 2 {
		return numOfPlayablePlayers
	}

	if currentPlayer.HasFolded { // leaver possible
		currentPlayer, currentSeat = t.nextPlayableSeatedPlayer(currentSeat)
	}

	for {
		currentPlayer.IsMyTurn = true // they start, if 3 players this will be newbanker
		currentPlayer.Status = 8      // is betting
		t.deadline = time.Now().UTC().Add(time.Second * time.Duration(PLAY_TIME))
		t.seatUpdateBroadcast()

		// turn base sync and update
		t.myTurnChan = make(chan bool)

		sleepTime := t.deadline.Sub(time.Now().UTC()).Seconds()

		// TEST
		// go func(currentPlayer *SeatedPlayer) {
		// 	time.Sleep(time.Duration(sleepTime+1) * time.Second)

		// 	logger.Println(currentPlayer.Uid)
		// 	if currentPlayer.IsMyTurn {
		// 		t.players[currentPlayer.Uid].PlaceBet(&protocol.PlaceBetRequest{
		// 			Amount: -2, // -1 for fold
		// 		})
		// 	}
		// }(currentPlayer)
		// TEST
		// logger.Printf("sleeptime: %v", sleeptime)

		// TEST ROBOT
		go func() {
			time.Sleep(time.Duration(sleepTime+1) * time.Second)
			if currentPlayer.Uid >= 0 && currentPlayer.Uid <= 10 {
				robotSeatInfo, _ := t.getSeatedPlayerByUid(currentPlayer.Uid)
				t.robotPlayerPlaceBet(robotSeatInfo) // 机器人自动跟注
			}
		}()
		// TEST ROBOT

		select {
		case <-time.After(time.Duration(sleepTime+1) * time.Second): // player run out of turn
			currentPlayer.Fold()
			break
		case <-t.myTurnChan:
			break
		}

		t.seatUpdateBroadcast()

		currentPlayer.HasPlayed = true // player has played used for determining proceed condition!!!
		currentPlayer.IsMyTurn = false // end current player turn

		if currentPlayer.HasAllIn {
			// sleepTime := 3 // 2+1
			time.Sleep(time.Duration(3) * time.Second)
		}

		if t.proceedCondition() { // this is for proceeding to next round
			logger.Println("finally we proceed")

			// time.Sleep(time.Duration(1) * time.Second)
			break
		}

		// currentPlayer.Status = -1 // current player turn finished
		currentPlayer, currentSeat = t.nextPlayableSeatedPlayer(currentSeat)

		close(t.myTurnChan)
		if currentPlayer.Uid >= 0 && currentPlayer.Uid <= 10 {
			break // 关闭机器人循环
		}
	}

	return t.numOfPlayablePlayers()
}

func (t *Table) proceedCondition() bool {
	if t.numOfPlayablePlayers() == 1 {
		player, _ := t.nextPlayableSeatedPlayer(0)
		// logger.Printf("player: %v", player)
		if player.BetAmount >= t.bettingAmount &&
			!player.HasPlayed && t.gameStatus != GAMESTATUS_PREFLOP {
			return true
		}
	}

	numOfInGamePlayers := t.numOfInGamePlayers() // do not use num of playable players, number may be incorrect because of players who all in
	// logger.Printf("numOfInGamePlayers: %v", numOfInGamePlayers)
	if numOfInGamePlayers <= 1 { // 1. check if there is at least 2 players (not folded and is in game)
		return true
	} else {
		numOfHasPlayed := 0
		for i, p := range t.seatedPlayers {
			if p != nil && t.seatedPlayerIsInGame(i) && (p.HasPlayed || p.HasAllIn ||
				p.UseableGameCoin == 0 ||
				p.UseableGameCoin == -1) { // 2. check all in game player have played HasPlayed
				numOfHasPlayed += 1
			}
		}

		logger.Printf("numOfHasPlayed 1: %v", numOfHasPlayed)

		if numOfHasPlayed == numOfInGamePlayers { // 3. check all in game player BetAmount is the same
			// t.bettingAmount
			toProceed := true
			for i, p := range t.seatedPlayers {
				if p != nil && t.seatedPlayerIsInGame(i) && p.HasPlayed {
					if p.BetAmount != t.bettingAmount && p.UseableGameCoin > 0 {
						// this means someone did not raise enough so need to go back to players turn
						toProceed = false
						// logger.Printf("p.UserName: %v", p.UserName)
						// logger.Printf("p.BetAmount: %v", p.BetAmount)
						// logger.Printf("t.bettingAmount: %v", t.bettingAmount)
						// logger.Printf("p.UseableGameCoin: %v", p.UseableGameCoin)
					}
					// if t.seatedPlayerIsInGame(i) && !t.seatedPlayers[i].HasAllIn &&
					// 	t.seatedPlayers[i].UseableGameCoin > 0 {
					// 	p.Status = -1
					// }
				}
			}

			if toProceed { // new round restart attribute
				t.seatUpdateBroadcast()
				time.Sleep(time.Duration(1) * time.Second)

				t.bettingAmount = 0
				for _, p := range t.seatedPlayers {
					if p != nil && p.IsInGame {
						p.HasPlayed = false
						p.IsMyTurn = false // just in case

						// if p.UseableGameCoin > 0 {
						p.BetAmount = 0
						// }

						if !p.HasAllIn &&
							p.UseableGameCoin > 0 {
							p.Status = -1
						}
					}
				}
			}
			return toProceed
		}
	}

	return false
}

func (t *Table) getSeatedPlayerByUid(uid int64) (*SeatedPlayer, int) {
	for i, p := range t.seatedPlayers {
		if p != nil && p.Uid == uid {
			return p, i
		}
	}

	return nil, -1
}

func (t *Table) getPlayingPlayer() (*SeatedPlayer, int) {
	for i, p := range t.seatedPlayers {
		if p != nil && p.IsMyTurn {
			return p, i
		}
	}
	return nil, -1
}

func (t *Table) seatedPlayerIsInGame(seatNumber int) bool {
	return t.seatedPlayers[seatNumber] != nil &&
		t.seatedPlayers[seatNumber].IsInGame &&
		!t.seatedPlayers[seatNumber].HasFolded
}

func (t *Table) prevInGameSeatedPlayer(index int) (*SeatedPlayer, int) {
	i := index
	for {
		i--

		if i < 0 {
			i = len(t.seatedPlayers) - 1
		}
		if t.seatedPlayerIsInGame(i) {
			return t.seatedPlayers[i], i
		}
	}
}

func (t *Table) nextPlayableSeatedPlayer(index int) (*SeatedPlayer, int) {
	i := index
	for {
		i++

		if i >= len(t.seatedPlayers) {
			i = 0
		}
		if t.seatedPlayerIsInGame(i) && !t.seatedPlayers[i].HasAllIn &&
			t.seatedPlayers[i].UseableGameCoin > 0 {
			return t.seatedPlayers[i], i
		}
	}
}

func (t *Table) numOfPlayablePlayers() int {
	numOfSeatedPlayers := 0
	for i := range t.seatedPlayers {
		if t.seatedPlayerIsInGame(i) &&
			t.seatedPlayers[i].UseableGameCoin > 0 {
			numOfSeatedPlayers++
		}
	}
	return numOfSeatedPlayers
}

func (t *Table) numOfInGamePlayers() int {
	numOfSeatedPlayers := 0
	for i := range t.seatedPlayers {
		if t.seatedPlayerIsInGame(i) {
			numOfSeatedPlayers++
		}
	}
	return numOfSeatedPlayers
}

func (t *Table) numOfSeatedPlayers() int {
	numOfSeatedPlayers := 0
	for _, p := range t.seatedPlayers {
		if p != nil {
			numOfSeatedPlayers++
		}
	}
	return numOfSeatedPlayers
}

func (t *Table) seatUpdateBroadcast() {
	// logger.Printf("t.gameStatus : %v", t.gameStatus)
	if t.gameStatus == GAMESTATUS_NOGAME {
		numOfSeatedPlayers := t.numOfSeatedPlayers()
		logger.Printf("numOfSeatedPlayers: %v", numOfSeatedPlayers)
		if numOfSeatedPlayers < NUMBER_OF_MINIMUM_PLAYERS && numOfSeatedPlayers != 0 {
			// here means nobody is in the room or the game yet
			// init the room
			go t.waitingPhase()
		}
	}
	// else if t.gameStatus == GAMESTATUS_WAITING && numOfSeatedPlayers >= NUMBER_OF_MINIMUM_PLAYERS {
	// 	go t.preflopPhase()
	// }

	bettingAmount := t.bettingAmount
	if bettingAmount == 0 {
		bettingAmount = t.SmallBlind
	}

	t.group.Broadcast(GAMESTATUS_SEAT_UPDATE, &protocol.SeatUpdateBroadcast{
		SeatedPlayersUpdate: t.seatedPlayers,
		Deadline:            t.deadline,
		GameStatus:          t.gameStatus,
		PrizePool:           t.prizePool,
		BettingAmount:       bettingAmount,
	}) // tell front end whats going on for the seatings
}

// init:

func newLimitedTable(title string, gameCoinToEnter, smallBlind, bigBlind,
	smallRaise, bigRaise int64) *Table {
	// 有限加注金额是固定的
	table := newTable(title, gameCoinToEnter, smallBlind, bigBlind)

	table.IsLimited = true
	table.SmallRaise = smallRaise
	table.BigRaise = bigRaise

	return table
}

func newTable(title string, gameCoinToEnter, smallBlind, bigBlind int64) *Table {
	numRobotPlayers := 4
	var robotPlayers []*protocol.RobotPlayer
	for i := 0; i < numRobotPlayers; i++ {
		robotPlayers = append(robotPlayers, NewRobotPlayer())
	}

	return &Table{
		gameStatus:      GAMESTATUS_NOGAME,
		seatedPlayers:   [NUMBER_OF_SEATS]*SeatedPlayer{},
		players:         map[int64]*Player{},
		cards:           newCards(),
		Title:           title,
		GameCoinToEnter: gameCoinToEnter,
		SmallBlind:      smallBlind,
		BigBlind:        bigBlind,
		IsLimited:       false,
		group:           nano.NewGroup(title),
		manager:         defaultManager,
		robotPlayers:    robotPlayers,
	}
}

func newCards() []protocol.Card {
	var cards []protocol.Card
	cards = append(cards, newSuit("s")...)
	cards = append(cards, newSuit("h")...)
	cards = append(cards, newSuit("c")...)
	cards = append(cards, newSuit("d")...)
	// cards = append(cards, &protocol.Card{
	// 	Value: 10,
	// 	Suit:  "s",
	// 	Face:  "Jo",
	// })
	// cards = append(cards, &protocol.Card{
	// 	Value: 10,
	// 	Suit:  "h",
	// 	Face:  "Jo",
	// })
	return cards
}

func randomRange(min int, max int) int {
	// return rand.Intn(max-min) + min

	return min + rand.Intn(max-min+1)
}

func newSuit(s string) []protocol.Card {
	return []protocol.Card{
		{
			Value: 14,
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
			Value: 11,
			Suit:  s,
			Face:  "J",
		},
		{
			Value: 12,
			Suit:  s,
			Face:  "Q",
		},
		{
			Value: 13,
			Suit:  s,
			Face:  "K",
		},
	}
}

func (t *Table) sessionCount() int {
	return len(t.players)
}

func (t *Table) setPlayer(uid int64, p *Player) {
	if _, ok := t.players[uid]; ok {
		logger.Warnf("玩家已经存在，正在覆盖玩家， UID=%d", uid)
	}
	t.players[uid] = p
}

func (t *Table) offline(uid int64) {
	for i, checkp := range t.seatedPlayers {
		if checkp != nil && checkp.Uid == uid {
			t.seatedPlayers[i].UseableGameCoin = -1
			t.seatedPlayers[i].Fold()
			t.seatedPlayers[i].session = nil
			t.seatedPlayers[i].BetAmount = 0

			if t.seatedPlayers[i].IsMyTurn {
				t.myTurnChan <- true // end my turn!
			}

			if t.gameStatus == GAMESTATUS_WAITING || !t.seatedPlayers[i].IsInGame {
				t.seatedPlayers[i] = nil
			}

			t.seatUpdateBroadcast()
		}
	}

	delete(t.players, uid) // golang func
	logger.Infof("ROOM 玩家: %d从在线列表中删除, 剩余：%d", uid, len(t.players))
}
