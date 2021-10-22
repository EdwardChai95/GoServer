package model

import "time"

type RouletteWinningItems struct {
	WinningItemsId int64 `xorm:"not null pk autoincr"`
	WinningAnimal  string
	WinningZone    string
	WinningText    string
	Level          string    // new level
	CreateAt       time.Time `xorm:"created index"`
}
