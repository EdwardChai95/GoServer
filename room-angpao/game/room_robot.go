package game

import (
	"math/rand"

	"gitlab.com/wolfplus/gamespace-lhd/db"
	"gitlab.com/wolfplus/gamespace-lhd/protocol"
)

// robot options
var robotNames = []string{"Minh Huệ", "Ngọc Thanh", "Lý Mỹ Kỳ", "Hồ Vĩnh Khoa", "Nguyễn Kim Hồng",
	"Phạm Gia Chi Bảo", "Ngoc Trinh", "Nguyễn Hoàng Bích", "Đặng Thu Thảo", "Nguyen Thanh Tung"}

func NewRobotPlayer() *protocol.RobotPlayer {
	return &protocol.RobotPlayer{
		Uid:      int64(rand.Intn(90-1) + 1),
		FaceUri:  db.Int64ToString(int64(rand.Intn(15-1) + 1)),
		UserName: robotNames[randomRange(0, len(robotNames)-1)],
		GameCoin: int64(randomRange(300000000, 900000000)),
	}
}

// robot player business logic

func (r *Room) robotGetAngPao(robot int) {
	rp := r.robotPlayers[robot]

	if r.gameStatus != "playing" {
		return
	}
	var last int
	last = int(r.total) - 1
	for i := 0; i < int(r.total); i++ {
		if r.angpaoList[last].Uid != 0 {

		} else {
			if r.angpaoList[i].Uid == 0 {
				r.angpaoList[i].Uid = rp.Uid
				r.angpaoList[i].FaceUri = rp.FaceUri
				r.angpaoList[i].UserName = rp.UserName
				r.group.Broadcast("angpaoPhase", &protocol.GetAngPaoResponse{
					Uid:        rp.Uid,
					Total:      r.minGameCoin,
					Left:       0,
					AngPaoList: r.angpaoList,
				})
				r.waitingPhase()
				return
			}
		}
	}
}

// end robot player business logic
