package game

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/lonng/nano"
	log "github.com/sirupsen/logrus"
	"gitlab.com/wolfplus/gamespace-lhd/db"
	"gitlab.com/wolfplus/gamespace-lhd/protocol"
)

type (
	Room struct {
		group           *nano.Group
		level           string
		name            string
		minGameCoin     int64
		gameStatus      string // facilitate game
		icon            string
		deadline        time.Time
		players         map[int64]*Player // 所有room的玩家
		total           int64
		index           int64
		auto            int64
		winners         int64
		winIndex        int64
		historyList     []protocol.HistoryItem // history
		angpaoList      []protocol.AngPao
		robotPlayers    []*protocol.RobotPlayer // robot players for this room
		logInformations []map[string]string
		sync.RWMutex
	}
)

func (r *Room) resultPhase() {
	if r.gameStatus != "animation" {
		return
	}

	// var wg sync.WaitGroup

	r.gameStatus = "result"
	r.deadline = time.Now().UTC().Add(time.Second * time.Duration(10)) // 10 deadline change

	for uid := range r.players {
		if p, ok := r.players[uid]; ok {
			gamecoin := db.GetGameCoinByUid(p.uid)
			r.group.Broadcast("resultPhase", &protocol.ResultPhaseResponse{
				Deadline:   r.deadline,
				GameCoin:   gamecoin,
				Index:      r.index,
				Winners:    r.winIndex,
				Min:        r.minGameCoin,
				AngPaoList: r.angpaoList,
			})
		}
	}

	s := r.deadline.Sub(time.Now()).Seconds() + 1
	time.Sleep(time.Duration(s) * time.Second)
	r.sendAngPaoPhase(r.winners)
}

func (r *Room) animationPhase() {
	if r.gameStatus != "playing" {
		return
	}
	r.gameStatus = "animation"
	r.deadline = time.Now().UTC().Add(time.Second * time.Duration(18)) // 15 deadline change

	var maxAmount, playerNum, robotNum, autoNum, playerWinnings, playerTax int64
	playerNum = 0
	robotNum = 0
	autoNum = 0

	maxAmount = r.angpaoList[0].Amount
	r.winners = r.angpaoList[0].Uid
	r.winIndex = r.angpaoList[0].Index

	index := r.randSender()
	r.index = index
	for i := 0; i < int(r.total); i++ {
		if maxAmount < r.angpaoList[i].Amount {
			maxAmount = r.angpaoList[i].Amount
			r.winIndex = r.angpaoList[i].Index
		}
		if r.index == int64(i) {
			r.winners = r.angpaoList[i].Uid
		}
	}

	for uid := range r.players {
		if p, ok := r.players[uid]; ok {
			if p.auto == 1 {
				autoNum += 1
			}
		}
	}

	for uid := range r.players {
		if p, ok := r.players[uid]; ok {
			gamecoin := db.GetGameCoinByUid(uid)

			for j := 0; j < int(r.total); j++ {
				if r.angpaoList[j].Uid > 0 && r.angpaoList[j].Uid < 100 {
					robotNum += 1
				} else {
					playerNum += 1
				}
				if r.angpaoList[j].Uid == p.uid {
					playerWinnings = r.angpaoList[j].Amount
					if playerWinnings > 0 {
						playerTax = int64(float64(playerWinnings) * 0.05)
						playerWinnings -= playerTax
					}
				}
			}
			otherInfo := fmt.Sprintf("[玩家人数：%v，机器人数量：%v，自动领取：%v，税收: %v]", playerNum, robotNum, autoNum, playerTax)

			for k := 0; k < int(r.total); k++ {
				if r.angpaoList[k].Uid == p.uid {
					logInformation := map[string]string{
						"uid":       db.Int64ToString(p.uid),
						"game":      "抢红包",
						"level":     r.level,
						"otherInfo": otherInfo,
						"before":    strconv.FormatInt(gamecoin, 10),
						"used":      strconv.FormatInt(playerWinnings, 10),
						"after":     strconv.FormatInt(gamecoin+playerWinnings, 10),
						"tax":       strconv.Itoa(int(playerTax)),
					}
					r.NewLogInformation(logInformation)
				}
			}

			logInformation1 := map[string]string{
				"uid":       "0",
				"game":      "抢红包",
				"level":     r.level,
				"otherInfo": otherInfo,
			}
			r.InsertAllLogInformations(logInformation1, otherInfo)
		}
		db.UpdateGameCoinByUid(uid, playerWinnings)
		//add 1012
		db.UpdateWinGameCoinByUid(uid, playerWinnings)
	}

	if len(r.historyList) >= 20 {
		r.historyList = r.historyList[:len(r.historyList)-1]
	}
	r.historyList = append([]protocol.HistoryItem{{
		Id:     r.winners,
		Amount: maxAmount,
	},
	}, r.historyList...)

	r.group.Broadcast("animationPhase", &protocol.AnimationPhaseResponse{
		Deadline:    r.deadline,
		Index:       r.index,
		AngPaoList:  r.angpaoList,
		HistoryList: r.historyList,
	})

	s := r.deadline.Sub(time.Now()).Seconds() + 1
	time.Sleep(time.Duration(s) * time.Second)
	r.resultPhase()
}

func (r *Room) playingPhase() {
	r.gameStatus = "playing"

	r.deadline = time.Now().UTC().Add(time.Second * time.Duration(15))

	r.angpaoList = []protocol.AngPao{}
	r.redPackage(int(r.total), int(r.minGameCoin))
	for uid := range r.players {
		if p, ok := r.players[uid]; ok {
			gamecoin := db.GetGameCoinByUid(uid)
			if p.session != nil {
				if gamecoin >= r.minGameCoin {
					if p.auto == 1 {
						r.getAngPao(p.uid, p.userName, p.faceUri)
					} else {
						r.waitingPhase()
					}
					fmt.Println("auto:", p.auto)
					r.group.Broadcast("playingPhase", &protocol.PlayingPhaseResponse{
						Deadline: r.deadline,
						Auto:     p.auto,
					})
				} else {
					r.group.Broadcast("kickPlayer", &protocol.KickPlayerResponse{
						Uid:   uid,
						Error: "游戏币不足，自动退出房间",
					})
				}
			}
		}
	}

	go func() {
		//a := r.deadline.Sub(time.Now()).Seconds()
		time.Sleep(time.Duration(5) * time.Second)
		for i := 0; i < int(r.total); i++ {
			if r.angpaoList[i].Uid == 0 {
				time.Sleep(time.Duration(1) * time.Second)
				r.robotGetAngPao(i)
			}
		}
	}()

	// count down to next game status
	s := r.deadline.Sub(time.Now()).Seconds() + 1
	time.Sleep(time.Duration(s) * time.Second)
	r.waitingPhase()
}

func (r *Room) waitingPhase() {
	var last int
	last = int(r.total) - 1
	for i := 0; i < int(r.total); i++ {
		if r.angpaoList[last].Uid != 0 {
			r.animationPhase()
		}
	}
}

func (r *Room) getAngPao(uid int64, userName, faceUri string) {
	var last int
	var left int64
	last = int(r.total) - 1
	left = int64(0)

	for i := 0; i < int(r.total); i++ {
		if r.angpaoList[last].Uid != 0 {
			r.waitingPhase()
		} else {
			if left == 0 {
				left = r.minGameCoin - r.angpaoList[i].Amount
			} else {
				left = left - r.angpaoList[i].Amount
			}

			if r.angpaoList[i].Uid == 0 {
				r.angpaoList[i].Uid = uid
				r.angpaoList[i].FaceUri = faceUri
				r.angpaoList[i].UserName = userName
				r.group.Broadcast("angpaoPhase", &protocol.GetAngPaoResponse{
					Uid:        uid,
					Total:      r.minGameCoin,
					Left:       left,
					AngPaoList: r.angpaoList,
				})
				fmt.Println("1")
				r.waitingPhase()
				return
			}
		}
	}
}

func (r *Room) sendAngPaoPhase(uid int64) {
	if uid > 0 && uid < 100 {
		r.playingPhase()
	} else {
		if p, ok := r.players[uid]; ok {
			gamecoin := db.GetGameCoinByUid(uid)

			otherInfo := fmt.Sprintf("[%v发放了%v红包]", uid, r.minGameCoin)
			if p.uid == uid {
				logInformation := map[string]string{
					"uid":       db.Int64ToString(p.uid),
					"game":      "发红包",
					"level":     r.level,
					"otherInfo": otherInfo,
					"before":    strconv.FormatInt(gamecoin, 10),
					"used":      strconv.FormatInt(-r.minGameCoin, 10),
					"after":     strconv.FormatInt(gamecoin-r.minGameCoin, 10),
				}
				r.NewLogInformation(logInformation)
			}
			logInformation1 := map[string]string{
				"uid":       "0",
				"game":      "发红包",
				"level":     r.level,
				"otherInfo": otherInfo,
			}
			r.InsertAllLogInformations(logInformation1, otherInfo)
		}
		db.UpdateGameCoinByUid(uid, -r.minGameCoin)
		db.UpdateWinGameCoinByUid(uid, -r.minGameCoin)
		r.playingPhase()
	}
}

func (r *Room) autoAngPao(uid int64) {
	if p, ok := r.players[uid]; ok {
		gamecoin := db.GetGameCoinByUid(uid)
		if p.session != nil {
			if gamecoin >= r.minGameCoin {
				if p.auto == 0 {
					p.auto = 1 // 自动领取
					r.group.Broadcast("autoAngPao", &protocol.AutoAngPaoResponse{
						Uid:  uid,
						Auto: p.auto,
						Tips: "已开启自动抢宝藏功能",
					})
					r.waitingPhase()
				} else {
					p.auto = 0 // 取消自动领取
					r.group.Broadcast("autoAngPao", &protocol.AutoAngPaoResponse{
						Uid:  uid,
						Auto: p.auto,
						Tips: "已关闭自动抢宝藏功能",
					})
					r.waitingPhase()
				}
			} else {
				r.group.Broadcast("kickPlayer", &protocol.KickPlayerResponse{
					Uid:   uid,
					Error: "游戏币不足，自动退出房间",
				})
			}
		}
	}
}

func (r *Room) randSender() int64 {
	var index int64
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	index = int64(r1.Intn(int(r.total)))

	return index
}

func randAngPao(remainCount, remainMoney int) int {
	if remainCount == 1 {
		return remainMoney
	}
	rand.Seed(time.Now().UnixNano())

	var min = 1
	max := remainMoney / remainCount * 2
	money := rand.Intn(max) + min
	return money
}

func (r *Room) redPackage(count, money int) {
	for i := 0; i < count; i++ {
		m := randAngPao(count-i, money)
		money -= m
		r.angpaoList = append(r.angpaoList, protocol.AngPao{
			Index:    int64(i),
			Uid:      int64(0),
			UserName: "",
			FaceUri:  "",
			Amount:   int64(m),
		})
	}

}

func randomRange(min int, max int) int {
	// return rand.Intn(max-min) + min

	return min + rand.Intn(max-min+1)
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
