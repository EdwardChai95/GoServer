package protocol

import (
	"time"
)

type (
	RobotPlayer struct {
		UserName        string
		CurrentBettings map[int]int64
		Awarded         int64
		GameCoin        int
	}

	Winner struct {
		Username string `json:"username"`
		Awarded  int    `json:"awarded"`
		IsRobot  bool   `json:"isrobot"`
	}

	Player struct {
		FaceUri  string // 头像地址
		UserName string // 玩家名字
		GameCoin int64
	}

	ResultPhase struct {
		GameCoin           int64
		Deadline           time.Time `json:"deadline"`
		TotalBetting       int64     `json:"totalBetting"`
		Winnings           int64     `json:"winnings"`
		Winners            []Winner  `json:"winners"`
		BankerWinnings     int64     `json:"bankerWinnings"`
		BankerCoins        int       `json:"bankerCoins"`
		BankerImg          int       `json:"bankerImg"`
		Gamblers           []*Gambler
		Banker             *Gambler `json:"banker"`
		CurrentWinnings    map[int]int64
		UpdateRobotWinners bool
	}

	AnimationPhaseResponse struct {
		Deadline time.Time  `json:deadline`
		Gamblers []*Gambler `json:gamblers`
		Banker   *Gambler   `json:banker`
	}

	RoomStatus struct {
		GameStatus  string    `json:gameStatus`
		Deadline    time.Time `json:deadline`
		GameCoin    int64     `json:"gameCoin"`
		BankerCoins int       `json:"bankerCoins"`
		BankerImg   int       `json:"bankerImg"`
		HistoryList [][]*Gambler
	}

	BettingPhaseResponse struct {
		Deadline time.Time `json:deadline`
	}

	PlaceBetResponse struct {
		Key    int
		Total  int64
		MyBet  int64
		Uid    int64
		Amount int64
	}

	PlaceBetRequest struct {
		Key    int   `json:"key"`
		Amount int64 `json:"amount"`
	}

	SendMessageErrorResponse struct {
		Error string `json:"error"`
	}

	JoinRoomResponse struct {
		Name    string  `json:"name"`
		FaceUri string  `json:"faceUri"`
		Chip    []int64 `json:"chip`
	}

	JoinRoomRequest struct {
		Uid        int64  `json:"uid"`
		Name       string `json:"name"`
		RoomNumber int64  `json:"roomNumber"`
		FaceUri    string `json:"faceUri"`
	}

	RoomItem struct {
		Title       string  `json:"title"`
		Label       string  `json:"label"`
		MinGameCoin int64   `json:"minGameCoin"`
		Chip        []int64 `json:"chip"`
		RoomType    int     `json:"roomtype"`
	}

	GetRoomResponse struct {
		Rooms []RoomItem `json:"rooms"`
	}

	Card struct {
		Value     int    `json:"value"` // 1 - 10
		Suit      string `json:"suit"`  // s, h, c, d
		Face      string `json:"face"`  // (A, J, Q, K, 2 - 10)
		FaceValue int    `json:"faceValue"`
	}

	Gambler struct {
		Title             string  `json:"title"`
		Totalbetting      int64   `json:"totalbetting"` // for logging
		WinLose           int64   `json:"winlose"`      // for logging
		CardValues        string  // for logging
		Odds              int64   `json:"odds"`
		Combo             string  `json:"combo"`      //
		RoundCards        []*Card `json:"roundcards"` // cards for this round
		IsWin             bool    `json:"isWin"`
		OrderedCardValues []*Card `json:"orderedCardValues"`
		Rank              int     `json:"rank"` // for comparing odds 1 mostly 0 - 13
	}
)
