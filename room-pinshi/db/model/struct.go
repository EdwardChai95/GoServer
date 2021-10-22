package model

import "time"

type PinshiTaxCollection struct {
	TaxCollectionId int64 `xorm:"not null pk autoincr"`
	TaxCollected    int64
	CreateAt        time.Time `xorm:"created index"`
}

type PinshiWinningItems struct {
	WinningItemsId int64 `xorm:"not null pk autoincr"`
	Gambler1       int64
	Gambler2       int64
	Gambler3       int64
	Gambler4       int64
	Level          string    //new level
	CreateAt       time.Time `xorm:"created index"`
}
