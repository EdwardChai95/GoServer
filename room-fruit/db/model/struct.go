package model

import "time"

type FruitWinningItem struct {
	WinningItemsId int64 `xorm:"not null pk autoincr"`
	WinningKey     int64
	CreateAt       time.Time `xorm:"created index"`
}
