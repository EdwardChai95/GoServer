package model

type AuthType int32
type CurrType int32
const (
	AccountPassword AuthType = 1
	PhonePassword   AuthType = 2
	PhoneCode       AuthType = 3
	WeixinOpenID    AuthType = 4
)
const (
	ETH int32 = 1
	BTC int32 = 2
)
