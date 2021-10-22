package model

type Config struct {
	ConfigId int64  `xorm:"not null pk autoincr"`
	Key      string `xorm:"unique"`
	Value    string `xorm:"text"`
}
