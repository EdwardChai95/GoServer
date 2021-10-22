package model

import "time"

type HchTaxCollection struct {
	TaxCollectionId int64 `xorm:"not null pk autoincr"`
	TaxCollected    int64
	CreateAt        time.Time `xorm:"created index"`
}

type HchWinningItems struct {
	WinningItemsId int64 `xorm:"not null pk autoincr"`
	WinningKey     int64
	CreateAt       time.Time `xorm:"created index"`
}
