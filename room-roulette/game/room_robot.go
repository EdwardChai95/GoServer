package game

import (
	"strconv"

	"gitlab.com/wolfplus/gamespace-lhd/protocol"
)

// robot options
var robotNames = []string{"Minh Huệ", "Ngọc Thanh", "Lý Mỹ Kỳ", "Hồ Vĩnh Khoa", "Nguyễn Kim Hồng",
	"Phạm Gia Chi Bảo", "Ngoc Trinh", "Nguyễn Hoàng Bích", "Đặng Thu Thảo", "Nguyen Thanh Tung"}

func NewRobotPlayer() *protocol.RobotPlayer {
	return &protocol.RobotPlayer{
		UserName:        robotNames[randomRange(0, len(robotNames)-1)],
		CurrentBettings: map[protocol.CurrentBetting]int64{},
		Awarded:         0,
		GameCoin:        int64(randomRange(300000000, 900000000)),
	}
}

func (r *Room) robotPlayerCalcResult(rp *protocol.RobotPlayer,
	winningBetZone protocol.BetItem, winningTextBetZone protocol.BetItem) (int64, int64, int64) {

	totalbetting, awarded, winningBet,
		_, _, _, logInformation := r.calcParticipantResult(rp.CurrentBettings)

	rp.Awarded = awarded

	rp.LogInformation = logInformation
	win_total, _ := strconv.ParseInt(logInformation["used"], 10, 64)
	if totalbetting > 0 {
		rp.LogInformation["uid"] = strconv.FormatInt(-1, 10)
		rp.LogInformation["before"] = strconv.FormatInt(rp.GameCoin, 10)
		rp.LogInformation["after"] = strconv.FormatInt(rp.GameCoin+win_total, 10)

		r.NewLogInformation(rp.LogInformation)
	}
	rp.GameCoin += win_total

	return totalbetting, awarded, winningBet
}

// robot player business logic

func (r *Room) robotPlayerPlaceBet(rp *protocol.RobotPlayer) {
	rp.Lock()
	defer rp.Unlock()

	if r.gameStatus != "betting" { // 1. make sure room status is betting
		return
	}

	var robotBetTotal int64 = 0
	for k := range rp.CurrentBettings {
		robotBetTotal += rp.CurrentBettings[k]
	}

	if robotBetTotal > r.maxRobotBetting { // 1.1 make sure total bet is less than xxx
		return
	}

	robotBetZone := r.randColors[randomRange(0, len(r.randColors)-1)]
	robotBetKey := randomRange(0, 4)
	var key = protocol.CurrentBetting{
		Zone: robotBetZone,
		Key:  robotBetKey,
	} // 2. random bet choice

	var bettingCoins = r.chip                                      // 3. available bet coins
	var amount = bettingCoins[randomRange(0, len(bettingCoins)-1)] // 4. random bet amount

	if val, ok := rp.CurrentBettings[key]; ok { // 5. update current betting
		val += amount
		rp.CurrentBettings[key] = val
	} else {
		rp.CurrentBettings[key] = amount
	}

	r.Lock()
	zone := r.getBetZoneByColor(robotBetZone)
	zone[robotBetKey].Total += amount                           // 6. update bet zone
	r.group.Broadcast("updateZone", &protocol.PlaceBetResponse{ // 7. broadcast to the other players
		Zone:  robotBetZone,
		Key:   robotBetKey,
		Total: zone[robotBetKey].Total,
		MyBet: rp.CurrentBettings[key],
		Uid:   -1,
	})
	r.Unlock()
}
