package protocol

import (
	"time"
)

type (
	Player struct {
		FaceUri  string // 头像地址
		UserName string // 玩家名字
		GameCoin int64
	}

	SpecialPrize struct {
		SpecialPrize int64 `json:"specialPrize"`
	}

	RoomSeatResponse struct {
		Seats [6]*Player `json:"seats"`
	}

	JoinSeatRequest struct {
		SeatIndex int `json:"seatindex"`
	}

	Winner struct {
		Username string `json:"username"`
		Awarded  int    `json:"awarded"`
		IsRobot  bool   `json:"isrobot"`
	}

	ClearBetResponse struct {
		BetZones []*BetItem `json:"betZones"`
		GameCoin int64      `json:"gameCoin"`
	}

	ResultPhase struct {
		PersonalBetResult  BetResult
		Awarded            int64
		GameCoin           int64
		Deadline           time.Time `json:deadline`
		Winners            []Winner  `json:"winners"`
		BankerWinLoseAmt   int64
		WinningBetZone     *BetItem `json:"winningBetZone"`
		BankerImg          int
		BankerAmt          int
		BankerRounds       int
		UpdateRobotWinners bool
	}

	BetResult struct {
		Key       int `json:"key"`
		Odds      int
		PlacedBet int64
		Reward    int64
	}

	AnimationPhaseResponse struct {
		Deadline       time.Time `json:deadline`
		WinningBetZone *BetItem  `json:"winningBetZone"`
		RandAngle      int       `json:"randAngle"`
	}

	RoomStatus struct {
		GameStatus   string     `json:gameStatus`
		Deadline     time.Time  `json:deadline`
		BetZones     []*BetItem `json:"betZones"`
		GameCoin     int64      `json:"gameCoin"`
		WinningItems []*BetItem `json:"winningItems"`
		BankerImg    int
		BankerAmt    int
		HistoryList  []*BetItem `json:"historylist"`
	}

	BettingPhaseResponse struct {
		Deadline time.Time `json:deadline`
	}

	PlaceBetRequest struct {
		Key    int   `json:"key"`
		Amount int64 `json:"amount`
	}

	PlaceBetResponse struct {
		Key   int   `json:"key"`
		Total int64 `json:"total"`
		MyBet int64 `json:"myBet"`
		Uid   int64 `json:"uid"`
	}

	BetItem struct {
		Odds     int
		Name     string
		Size     string
		Image    string
		NickName string
		Total    int64
	}

	SendMessageErrorResponse struct {
		Error string `json:"error"`
	}

	SendMessageResponse struct {
		Uid     int64  `json:"uid"`
		Label   string `json:"label"`
		FaceUri string `json:"faceUri"`
		Message string `json:"message"`
		Channel int64  `json:"channel"`
	}

	SendMessageRequest struct {
		Message string `json:"message"`
		Channel int64  `json:"channel"`
	}

	JoinRoomResponse struct {
		Name     string     `json:"name"`
		FaceUri  string     `json:"faceUri"`
		BetZones []*BetItem `json:"betZones"`
	}

	JoinRoomRequest struct {
		Uid        int64  `json:"uid"`
		Name       string `json:"name"`
		RoomNumber int64  `json:"roomNumber"`
		FaceUri    string `json:"faceUri"`
	}

	RoomItem struct {
		Level       string  `json:"level"`
		Name        string  `json:"name"`
		MinGameCoin int64   `json:"minGameCoin"`
		Chip        []int64 `json:"chip"`
		Icon        string  `json:"icon"`
	}

	GetRoomResponse struct {
		Rooms []RoomItem `json:"rooms"`
	}
)
