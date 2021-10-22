package model

import "time"

type GameCollection struct {
	GameCollectionId   int64  `xorm:"not null pk autoincr"`
	GameName           string `xorm:"unique(game_identifier)"`
	GameRoom           string `xorm:"unique(game_identifier)"`
	GameCoinCollection int64  `xorm:"not null DEFAULT 0"`
	GameCoinBetting    int64  `xorm:"not null DEFAULT 0"`
	GameCoinPercentage float64
}

type DailyCollection struct {
	GameCollectionId   int64  `xorm:"not null pk autoincr"`
	GameName           string `xorm:"unique(game_identifier)"`
	GameRoom           string `xorm:"unique(game_identifier)"`
	GameCoinCollection int64  `xorm:"not null DEFAULT 0"`
	GameCoinBetting    int64  `xorm:"not null DEFAULT 0"`
	GameCoinPercentage float64
}

type Account struct {
	Aid         int64    `xorm:"not null pk autoincr"`
	Uid         string   `xorm:"index"`
	Type        AuthType `xorm:"index"`
	Account     string   `xorm:"index"`
	Password    string   `xorm:"text"`
	CreateAt    time.Time
	CreateIp    string
	LastLoginAt time.Time
	LastLoginIp string
}
type UserProfile struct {
	Uid     string `xorm:"pk"`
	Phone   string
	Address string
}
type Currency struct {
	Uid     string `xorm:"index"`
	Type    int32  `xorm:"index"`
	Balance float64
}
type User struct {
	Uid int64 `xorm:"not null pk autoincr"`

	//template fields (原本模板的字段，将会取消或更替)
	Username string //用户名
	FaceUri  int64  `xorm:"not null DEFAULT 1"` //头像链接地址
	Money    int64  //钱币
	Diamond  int64  //钻石
	//

	//haige added fields (适应客户业务逻辑的字段)
	PhoneNumber int64
	UserAcc     int64
	Password    string `xorm:"text"`

	GameCoin  int64  `xorm:"not null DEFAULT 0"`
	NickName  string //昵称
	Signature string //签名

	UserLevel string

	UserPermission string //for distinguishing admin

	CreateAt    time.Time //创建时间
	LastLoginAt time.Time //上一次登陆时间
	LastLoginIp string    //上一次登陆ip

	CreateOS    string
	LastLoginOS string

	AccLogin             int64 //accumulated login
	LoginReward          string
	LoginRewardClaimTime time.Time //上一次登陆时间

	Experience float64 `xorm:"DECIMAL(16,4) not null default 0.0000"`
	Level      int     `xorm:"not null default 0"`

	//
	VipLevel     int8 // 普通玩家，高级玩家，特级玩家
	FirstPay     int  `xorm:"not null DEFAULT 0"`
	FirstDeposit int  `xorm:" DEFAULT 0"`

	NormalActive int64  `xorm:" DEFAULT 0"`
	Imei         string // 设备id
	Gender       int64  `xorm:" DEFAULT 0"`

	CountCompleted int   `xorm:" DEFAULT 0"`
	CountTaken     int   `xorm:" DEFAULT 0"`
	WinGameCoin    int64 `xorm:" DEFAULT 0"`

	Proxy int64 `xrom: "DEFAULT 0"`
}

type Item struct {
	Uid int64 `xorm:"pk"`

	LaBa int64
}

type UserGameData struct {
	Uid          string `xorm:"index"`
	GameId       int32  `xorm:"index"`
	Score        int64
	Win          int32
	Loss         int32
	Tie          int32
	CreateAt     time.Time `xorm:"created"`
	LastUpdateAt time.Time `xorm:"updated"`
}

//for wallet
type GameCoinLocker struct {
	Uid int64 `xorm:"pk"`

	Balance  int64
	Password string `xorm:"text"`
}

type GameCoinLockerHistory struct {
	Id int64 `xorm:"not null pk autoincr"`

	Uid     int64
	Operate string
	Amount  int64
	Date    time.Time
	Balance int64
}

type HighAndSpecialCustomer struct {
	Id int64 `xorm:"not null pk autoincr"`

	SpecialUid int64
	HighUid    int64
}

type TempAnnouncement struct {
	Id      int64 `xorm:"not null pk autoincr"`
	Message string
}

type ExchangeCode struct {
	ExchangeCodeId int64  `xorm:"not null pk autoincr"`
	ExchangeCode   string `xorm:"unique(uniq_code)"`
	GameCoin       int64  `xorm:"unique(uniq_code)"`
	LaBa           int64  `xorm:"unique(uniq_code)"`
	IsUsed         int32  `xorm:"not null default 0"`
	MaxUsage       int32
	CodeType       int32
	ValidUntilTime time.Time `xorm:"unique(uniq_code)"`
	ProxyUid       int64     `xorm:"not null default 0"`
}

type NewYearEvent struct {
	Uid  int64 `xorm:"not null pk"`
	Day1 int   `xorm:"not null default 0"`
	Day2 int   `xorm:"not null default 0"`
	Day3 int   `xorm:"not null default 0"`
	Day4 int   `xorm:"not null default 0"`
	Day5 int   `xorm:"not null default 0"`
	Day6 int   `xorm:"not null default 0"`
}

type Order struct {
	OrderId         int64 `xorm:"not null pk autoincr"`
	Uid             int64
	CreatedDatetime time.Time
	OrderStatus     string
	PaymentAmount   int64  // 越南币的数量
	GameCoinAmount  int64  // how much game coin purchased
	Rebate          int64  // 返利
	Type            string // offline or vc
	UpdatedDatetime time.Time
	Comment         string
	CallbackData    string `xorm:"text"`
}

type TotalGameCoin struct {
	TotalId  int64 `xorm:"not null pk autoincr"`
	GameCoin int64
	CreateAT time.Time
}

type VCode struct {
	VCodeId     int64 `xorm:"not null pk autoincr"`
	PhoneNumber int64
	VCode       string
	CreateAt    time.Time
}

type Proxy struct {
	ProxyId        int64 `xorm:"not null pk autoincr"`
	Uid            int64
	PromoCode      string
	PromoNum       int64 `xorm:"not null default 0"`
	ActiveNum      int64 `xorm:"not null default 0"`
	SendNum        int64 `xorm:"not null default 0"`
	ReceiveNum     int64 `xorm:"not null default 0"`
	TotalNum       int64 `xorm:"not null default 0"`
	TotalAmount    int64 `xorm:"not null default 0"`
	ServiceTax     int64 `xorm:"not null default 0"`
	CountCompleted int64 `xorm:"not null default 0"`
	OperatingTime  time.Time
}

type ProxyUser struct {
	ProxyUserId    int64 `xorm:"not null pk autoincr"`
	Uid            int64
	ProxyUid       int64 `xorm:"not null default 0"`
	TotalWin       int64 `xorm:"not null default 0"`
	TotalLose      int64 `xorm:"not null default 0"`
	TotalWinLose   int64 `xorm:"not null default 0"`
	TotalBroad     int64 `xorm:"not null default 0"`
	SendNum        int64 `xorm:"not null default 0"`
	ReceiveNum     int64 `xorm:"not null default 0"`
	TotalAmount    int64 `xorm:"not null default 0"`
	CountCompleted int64 `xorm:"not null default 0"`
	OperatingTime  time.Time
}

type CodeInformation struct {
	CodeInformationId int64  `xorm:"not null pk autoincr"`
	ExchangeCode      string `xorm:"unique(uniq_code)"`
	Uid               int64
	GameCoin          int64 `xorm:"unique(uniq_code)"`
	LaBa              int64 `xorm:"unique(uniq_code)"`
	Reason            string
	OtherInfo         string `xorm:"text"`
	OperatingTime     time.Time
	IsUsed            int
	MaxUsage          int
	ValidUntilTime    time.Time
}
