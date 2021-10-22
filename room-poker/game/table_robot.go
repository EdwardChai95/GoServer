package game

import (
	"fmt"
	"math/rand"

	"gitlab.com/wolfplus/gamespace-lhd/db"
	"gitlab.com/wolfplus/gamespace-lhd/protocol"
)

// robot options
var robotNames = []string{"错迎海", "乐寒梅", "示曜", "南宫忆秋", "玄绍祺",
	"荣芳泽", "南门朗", "幸承福", "丰瑜蓓", "实书双"}

func NewRobotPlayer() *protocol.RobotPlayer {
	return &protocol.RobotPlayer{
		Uid:      int64(rand.Intn(9-1) + 1),
		FaceUri:  db.Int64ToString(int64(rand.Intn(10-1) + 1)),
		UserName: robotNames[randomRange(0, len(robotNames)-1)],
		GameCoin: int64(randomRange(300000000, 900000000)),
	}
}

func (t *Table) robotPlayerTakeSeat(i int) {
	rp := t.robotPlayers[i]
	if t.seatedPlayers[i] == nil {
		t.seatedPlayers[i] = &SeatedPlayer{
			Uid:             rp.Uid,
			FaceUri:         rp.FaceUri,
			UserName:        rp.UserName,
			OverallGameCoin: rp.GameCoin,
			UseableGameCoin: rp.GameCoin / 2,
			BetAmount:       0,
			TotalBetAmount:  0,
			CardRank:        0,
			IsInGame:        false,
			HasAllIn:        false,
			HasFolded:       false,
			IsMyTurn:        false,
			IsBanker:        false,
			HasPlayed:       false,
			Status:          2,
			cards:           []protocol.Card{},
			ShowCards:       []protocol.Card{},
			HighestHand:     []protocol.Card{},
			PhaseBetting:    map[string]int64{},
		}
		t.seatUpdateBroadcast()
	}
	if t.gameStatus == GAMESTATUS_WAITING {
		numOfSeatedPlayers := t.numOfSeatedPlayers()
		if numOfSeatedPlayers >= NUMBER_OF_MINIMUM_PLAYERS {
			go t.preflopPhase()
			return
		}
	}
}

func (t *Table) robotPlayerPlaceBet(rp *SeatedPlayer) {
	robotSeatInfo, _ := t.getSeatedPlayerByUid(rp.Uid)
	if !robotSeatInfo.IsMyTurn {
		fmt.Println("not your turn")
		return
	}

	// 0 CHECK, -1 FOLD, -2 FOLLOW
	amount := 0
	if len(robotSeatInfo.cards) > 0 {
		card1 := robotSeatInfo.cards[0]
		card2 := robotSeatInfo.cards[1]
		amount = robotPlayerCheckCard(card1.Value, card2.Value, int(t.bettingAmount))
	}

	if amount == 0 {
		robotSeatInfo.Status = 6
	} else if amount == -1 {
		robotSeatInfo.Fold()
	} else {
		amount = int(t.bettingAmount) - int(robotSeatInfo.BetAmount)
		robotSeatInfo.Status = 4
	}
	t.myTurnChan <- true
}

func robotPlayerCheckCard(card1, card2, totalbet int) int {
	var amount, card int
	amount = 0
	card = 0
	if card1 > card2 {
		card = card1
	} else {
		card = card2
	}
	if card == 2 && totalbet <= 2000 {
		amount = -2
	} else if card == 2 && totalbet > 2000 {
		amount = -1
	}
	if card == 3 && totalbet <= 3000 {
		amount = -2
	} else if card == 3 && totalbet > 3000 {
		amount = -1
	}
	if card == 4 && totalbet <= 4000 {
		amount = -2
	} else if card == 4 && totalbet > 4000 {
		amount = -1
	}
	if card == 5 && totalbet <= 5000 {
		amount = -2
	} else if card == 5 && totalbet > 5000 {
		amount = -1
	}
	if card == 6 && totalbet <= 6000 {
		amount = -2
	} else if card == 6 && totalbet > 6000 {
		amount = -1
	}
	if card == 7 && totalbet <= 7000 {
		amount = -2
	} else if card == 7 && totalbet > 7000 {
		amount = -1
	}
	if card == 8 && totalbet <= 8000 {
		amount = -2
	} else if card == 8 && totalbet > 8000 {
		amount = -1
	}
	if card == 9 && totalbet <= 9000 {
		amount = -2
	} else if card == 9 && totalbet > 9000 {
		amount = -1
	}
	if card == 10 && totalbet <= 10000 {
		amount = -2
	} else if card == 10 && totalbet > 10000 {
		amount = -1
	}
	if card == 11 && totalbet <= 10000 {
		amount = -2
	} else if card == 11 && totalbet > 10000 {
		amount = -1
	}
	if card == 12 && totalbet <= 10000 {
		amount = -2
	} else if card == 12 && totalbet > 10000 {
		amount = -1
	}
	if card == 13 && totalbet <= 10000 {
		amount = -2
	} else if card == 13 && totalbet > 10000 {
		amount = -1
	}
	if card == 14 && totalbet <= 10000 {
		amount = -2
	} else if card == 14 && totalbet > 10000 {
		amount = -1
	}

	return amount
}
