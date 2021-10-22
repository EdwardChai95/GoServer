package game

import (
	// "github.com/ethereum/go-ethereum/log"
	"github.com/lonng/nano/session"
	"gitlab.com/wolfplus/gamespace-lhd/protocol"
)

type (
	Player struct {
		uid      int64  // 用户ID
		faceUri  string // 头像地址
		level    int
		userName string // 玩家名字
		session  *session.Session
		table    *Table
		manager  *Manager
	}

	SeatedPlayer struct {
		Uid      int64
		FaceUri  string // 头像地址
		UserName string // 玩家名字
		session  *session.Session

		// match var
		PhaseBetting    map[string]int64
		OverallGameCoin int64
		UseableGameCoin int64
		BetAmount       int64
		TotalBetAmount  int64
		WinAmount       int64
		cards           []protocol.Card // do not broadcast
		CardRank        int
		CardRankTitle   string
		ShowCards       []protocol.Card // only showdown broadcast
		HighestHand     []protocol.Card // only showdown broadcast

		// match status
		HasAllIn  bool
		HasFolded bool
		IsMyTurn  bool
		IsBanker  bool
		IsInGame  bool
		Status    int  // for frontend display
		HasPlayed bool // for determining proceed condition
	}
)

// business func

func (mySeatInfo *SeatedPlayer) Fold() {
	mySeatInfo.Status = 3
	mySeatInfo.HasFolded = true
}

func (p *Player) PlaceBet(req *protocol.PlaceBetRequest) bool {
	if p == nil {
		return false
	}
	myTable := p.table

	mySeatInfo, _ := myTable.getSeatedPlayerByUid(p.uid)

	if !mySeatInfo.IsMyTurn {
		return false
	}

	amount := req.Amount // 0 CHECK, -1 FOLD, -2 FOLLOW
	logger.Println(amount)

	if amount == 0 ||
		(myTable.bettingAmount == 0 && amount == -2) ||
		(myTable.bettingAmount == mySeatInfo.BetAmount && amount == -2) { // check
		// logger.Printf("mySeatInfo.BetAmount < myRoom.bettingAmount: %v", mySeatInfo.BetAmount < myRoom.bettingAmount)
		if mySeatInfo.BetAmount < myTable.bettingAmount &&
			myTable.bettingAmount != 0 {
			return false
		}
		if amount == 0 || mySeatInfo.BetAmount == 0 {
			mySeatInfo.Status = 6
		} else {
			mySeatInfo.Status = -1
		}

	} else if amount == -1 { // fold
		mySeatInfo.Fold()
	} else { // raise or follow
		// logger.Printf("amount: %v", amount)
		// logger.Printf("mySeatInfo.BetAmount: %v", mySeatInfo.BetAmount)
		// logger.Printf("myTable.bettingAmount: %v", myTable.bettingAmount)
		// logger.Printf("follow condition: %v", amount <= -2 ||
		// 	amount < myTable.bettingAmount)

		if amount <= -2 ||
			amount < myTable.bettingAmount { // follow, treat it as such even if not
			amount = myTable.bettingAmount - mySeatInfo.BetAmount
			mySeatInfo.Status = 4 // follow
		}

		if amount > 0 { // sufficient amount

			if mySeatInfo.UseableGameCoin <= amount { // all in
				amount = mySeatInfo.UseableGameCoin //- mySeatInfo.BetAmount
				mySeatInfo.HasAllIn = true
				mySeatInfo.Status = 7
			} else if mySeatInfo.Status != 4 {
				mySeatInfo.Status = 5
			}

			mySeatInfo.BetAmount += amount
			mySeatInfo.TotalBetAmount += amount
			mySeatInfo.UseableGameCoin -= amount

			myTable.prizePool += amount

			if myPhaseBetting, ok := mySeatInfo.PhaseBetting[myTable.gameStatus]; ok {
				myPhaseBetting += amount
				mySeatInfo.PhaseBetting[myTable.gameStatus] = myPhaseBetting
			} else {
				mySeatInfo.PhaseBetting[myTable.gameStatus] = amount
			}

			if amount > myTable.bettingAmount {
				myTable.bettingAmount = mySeatInfo.BetAmount // bet standard for other players to follow
			}
		} else {
			mySeatInfo.Fold()
		}

	}

	// mySeatInfo.IsMyTurn = false
	// mySeatInfo.HasPlayed = true

	myTable.myTurnChan <- true // end my turn!

	return true
}

// helper functions

// init functions

func newPlayer(s *session.Session, uid int64, name string, faceUri string, level int,
	m *Manager, selectedRoom *Table) *Player {
	p := &Player{
		uid:      uid,
		userName: name,
		faceUri:  faceUri,
		level:    level,
		manager:  m,
	}

	p.bindSession(s, selectedRoom)

	return p
}

func (p *Player) bindSession(s *session.Session, selectedRoom *Table) {
	p.session = s
	p.table = selectedRoom
	p.session.Set(kCurPlayer, p)
}

func (p *Player) removeSession() {
	p.session.Remove(kCurPlayer)
	p.session = nil
}

func (p *Player) Uid() int64 {
	return p.uid
}
