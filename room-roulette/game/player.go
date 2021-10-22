package game

import (
	"sync"

	"github.com/lonng/nano/session"
	log "github.com/sirupsen/logrus"
	"gitlab.com/wolfplus/gamespace-lhd/db"
	"gitlab.com/wolfplus/gamespace-lhd/protocol"
)

type (
	Player struct {
		uid             int64  // 用户ID
		faceUri         string // 头像地址
		userName        string // 玩家名字
		session         *session.Session
		logger          *log.Entry // 日志
		room            *Room
		club            *Club
		currentBettings map[protocol.CurrentBetting]int64
		animalBet       protocol.BetResult
		textBet         protocol.BetResult
		awarded         int64
		seatIndex       int
		bonus           int64
		level           int
		logInformation  map[string]string
		sync.RWMutex
	}
)

func (p *Player) leaveSeat() {
	if p.seatIndex != -1 {
		r := p.room
		r.seats[p.seatIndex] = nil
		r.seatsUpdate()
		p.seatIndex = -1
	}
}

func (p *Player) joinSeat(seatIndex int) {
	r := p.room
	gamecoin := db.GetGameCoinByUid(p.Uid()) // get current gamecoin
	if gamecoin > r.minSeatCoin {            // minimum need to have 3000000 gamecoin
		if r.seats[seatIndex] == nil && p.seatIndex == -1 {
			r.seats[seatIndex] = p
			p.seatIndex = seatIndex
		}
	}
}

// old function
// func (p *Player) calcPlayerAdditionalResult(zone string, key int) {
// 	// key is animal
// 	// zone is color
// 	r := p.room
// 	var awarded int64
// 	for k := range p.currentBettings {
// 		if (k.Zone == zone || k.Key == key) && k != r.winningBetZone {
// 			zoneRow := r.getBetZoneByColor(k.Zone)
// 			zoneRowChild := zoneRow[k.Key]
// 			animalwinning := p.currentBettings[k] * int64(zoneRowChild.Odds)
// 			p.animalBet = protocol.BetResult{
// 				Zone:      k.Zone,
// 				Key:       k.Key,
// 				Odds:      zoneRowChild.Odds,
// 				PlacedBet: p.currentBettings[k],
// 				Reward:    animalwinning,
// 			}
// 			awarded += animalwinning
// 		}
// 	}
// 	p.awarded += awarded
// 	p.bonus = awarded
// 	if awarded > 0 {
// 		db.UpdateGameCoinByUid(p.Uid(), awarded) // update gamecoin of player in db
// 		go db.NewGameCoinTransaction(p.Uid(), awarded)
// 	}
// }

func (p *Player) clearAllBets() {

	for currentBetting, currentBettingAmount := range p.currentBettings {
		zone := p.room.betZones.Red
		if currentBetting.Zone == "Green" {
			zone = p.room.betZones.Green
		} else if currentBetting.Zone == "Yellow" {
			zone = p.room.betZones.Yellow
		}
		zone[currentBetting.Key].Total -= currentBettingAmount
		p.room.group.Broadcast("updateZone", &protocol.PlaceBetResponse{
			Zone:  currentBetting.Zone,
			Key:   currentBetting.Key,
			Total: zone[currentBetting.Key].Total,
			MyBet: 0,
			Uid:   p.uid,
		})
	}
	p.currentBettings = map[protocol.CurrentBetting]int64{}
}

func (p *Player) getAllBetsTotal() int64 {
	var total int64 = 0
	for k := range p.currentBettings {
		total += p.currentBettings[k]
	}
	return total
}

func (p *Player) getBet(zone string, key int) int64 {
	return p.currentBettings[protocol.CurrentBetting{
		Zone: zone,
		Key:  key,
	}]
}

func (p *Player) placeBet(zone string, key int, amount int64) bool {

	betkey := protocol.CurrentBetting{
		Zone: zone,
		Key:  key,
	}
	if amount+p.currentBettings[betkey] > p.room.maxGameCoin {
		return false
	}
	if val, ok := p.currentBettings[betkey]; ok {
		val += amount
		p.currentBettings[betkey] = val
	} else {
		p.currentBettings[betkey] = amount
	}
	return true
}

func newPlayer(s *session.Session, uid int64, name string, faceUri string,
	room *Room, club *Club, level int) *Player {
	p := &Player{
		uid:             uid,
		userName:        name,
		faceUri:         faceUri,
		club:            club,
		logger:          log.WithField("player", uid),
		currentBettings: map[protocol.CurrentBetting]int64{},
		animalBet:       protocol.BetResult{},
		textBet:         protocol.BetResult{},
		seatIndex:       -1,
		level:           level,
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
