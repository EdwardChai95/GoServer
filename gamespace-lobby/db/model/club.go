package model

import "time"

type Club struct {
	ClubId              int64 `xorm:"not null pk autoincr"`
	ClubName            string
	GameCoin            int64
	CreateAt            time.Time `xorm:"created"` //创建时间
	LastUpdateAt        time.Time `xorm:"updated"`
	Level               int
	ToDeleteTime        time.Time
	Description         string `xorm:"text"`
	Announcement        string `xorm:"text"`
	MonthlyContribution float64
	WeeklyContribution  float64
}

type ClubUser struct {
	ClubId             int64 `xorm:"unique(uniq_clubuser)"`
	Uid                int64 `xorm:"unique(uniq_clubuser)"`
	Type               string
	WeeklyContribution float64
	AbleToCollect      float64
	IsMute             int
	ApprovedDate       time.Time
}

type CustomerServiceMessage struct {
	CustomerServiceMessageId int64 `xorm:"not null pk autoincr"`
	PlayerId                 int64
	SenderID                 int64
	Message                  string
	AdminReplied             int32
	IsRead                   int32
	IsAdminRead              int32
	TimeSent                 time.Time
}

type LogInformation struct {
	LogInformationId int64 `xorm:"not null pk autoincr"`
	Uid              int64
	Reason           string
	Game             string
	Level            string
	OtherInfo        string `xorm:"text"`
	OperatingTime    time.Time
	Result           int64
	Rate             float64
	BetTotal         int64
	WinTotal         int64
	BankerWinTotal   int64
	Params           int64  // for game log
	Before           int64  // for game log
	Used             int64  // for game log
	After            int64  // for game log
	TaskKey          string // for task only
	Tax 		 int64  // for game log
}

type GameCoinTransaction struct {
	GameCoinTransactionId int64 `xorm:"not null pk autoincr"`
	Uid                   int64
	ClubId                int64
	Value                 int64
	Type                  string
	Comment               string
	Datetime              time.Time
	Collected             int8
}
