package protocol

import (
	"sync"
	"time"
)

type (
	OddsProbReferenceRow struct {
		Odds int
		Prob float64
	}

	RobotPlayer struct {
		UserName        string
		CurrentBettings map[CurrentBetting]int64
		Awarded         int64
		GameCoin        int64
		LogInformation  map[string]string
		sync.RWMutex
	}

	CurrentBetting struct {
		Zone string
		Key  int
	}

	Player struct {
		FaceUri  string // 头像地址
		UserName string // 玩家名字
		GameCoin int64
		Level    int
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
	}

	ClearBetResponse struct {
		BetZones BetZones `json:"betZones"`
		GameCoin int64    `json:"gameCoin"`
	}

	ResultPhase struct {
		AnimalBet          BetResult
		TextBet            BetResult
		Awarded            int64
		Bonus              int64
		GameCoin           int64
		Deadline           time.Time
		Winners            []Winner `json:"winners"`
		Special            string   `json:"special"`
		UpdateRobotWinners bool
	}

	BetResult struct {
		Zone      string `json:"zone"`
		Key       int    `json:"key"`
		Odds      int
		PlacedBet int64
		Reward    int64
	}

	AnimationPhaseResponse struct {
		Deadline           time.Time
		WinningBetZone     BetItem       `json:"winningBetZone"`
		WinningTextBetZone BetItem       `json:"winningTextBetZone"`
		RandAngle          int           `json:"randAngle"`
		RandAnimals        [24]string    `json:"randAnimals"`
		RandLights         [24]string    `json:"randLights"`
		HistoryList        []HistoryItem `json:"historylist"` // history
	}

	RoomStatus struct {
		GameStatus  string
		Deadline    time.Time
		BetZones    BetZones   `json:"betZones"`
		GameCoin    int64      `json:"gameCoin"`
		RandAnimals [24]string `json:"randAnimals"`
		RandLights  [24]string `json:"randLights"`
		Bg          string
		HistoryList []HistoryItem `json:"historylist"` // history
	}

	BettingPhaseResponse struct {
		Deadline    time.Time
		BetZones    BetZones   `json:"betZones"`
		RandAnimals [24]string `json:"randAnimals"`
		RandLights  [24]string `json:"randLights"`
	}

	PlaceBetResponse struct {
		Zone  string `json:"zone"`
		Key   int    `json:"key"`
		Total int64  `json:"total"`
		MyBet int64  `json:"myBet"`
		Uid   int64  `json:"uid"`
	}

	PlaceBetRequest struct {
		Zone   string `json:"zone"`
		Key    int    `json:"key"`
		Amount int64
	}

	HistoryItem struct {
		Color   string `json:"color"`
		Animal  string `json:"animal"`
		Text    string `json:"text"`
		Special string `json:"special"`
	}

	BetItem struct {
		Bg     string
		IconBg string
		Icon   string
		Odds   int
		Total  int64
		Name   string
	}

	BetZones struct {
		Red    []BetItem
		Green  []BetItem
		Yellow []BetItem
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
		Name      string `json:"name"`
		RoomLevel string `json:"roomlevel"`
		FaceUri   string `json:"faceUri"`
		Chip      []int64
		BetZones  BetZones `json:"betZones"`
	}

	JoinRoomRequest struct {
		Uid        int64  `json:"uid"`
		Name       string `json:"name"`
		RoomNumber int64  `json:"roomNumber"`
		FaceUri    string `json:"faceUri"`
		Level      int    `json:"level"`
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
