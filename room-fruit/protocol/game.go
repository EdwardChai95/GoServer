package protocol

import (
	"time"
)

type (
	PlaceBetRequest struct {
		BetZoneKey     int   `json:"betZoneKey"`
		PlaceBetAmount int64 `json:"placeBetAmount"`
	}

	PlaceBetResponse struct {
		BetZoneKey int   `json:"betZoneKey"`
		TotalBet   int64 `json:"totalBet"`
		MyBet      int64 `json:"myBet"`
		Uid        int64 `json:"uid"`
	}

	ClearBetResponse struct {
		BetZones []*BetZone `json:"betZones"` // 押注块
		GameCoin int64      `json:"gameCoin"`
	}

	BetResult struct {
		BetZoneKey int   `json:"betZoneKey"`
		Odds       int   `json:"odds"` // 赔率
		PlacedBet  int64 `json:"placedBet"`
		Reward     int64 `json:"reward"`
	}

	Winner struct {
		Username string `json:"username"`
		Image    string `json:"image"`
		Level    int    `json:"level"`
		Awarded  int    `json:"awarded"`
		IsRobot  bool   `json:"isrobot"`
	}

	RoomStatusResponse struct {
		GameStatus         string         `json:"gameStatus"`
		Deadline           time.Time      `json:"deadline"`
		BetZones           []*BetZone     `json:"betZones"`     // 押注块
		Chips              []int          `json:"chips"`        // 筹码
		HistoryList        []*WinningItem `json:"historyList"`  // 开奖历史
		WinningItem        []*WinningItem `json:"winningItems"` // 赢的物品
		LevelToChat        int            `json:"levelToChat"`  //聊天等级
		GameCoin           int64          `json:"gameCoin"`
		LastWinningItemKey int            `json:"lastWinningItemKey"`
	}

	WinningItem struct {
		WinningBetZoneIndex int    `json:"winningBetZone"` // 对应押注区
		Probability         int    `json:"probability"`    // chance of winning upon 1000
		Odds                int    `json:"odds"`           // 赔率
		Image               string `json:"image"`
		IsBigFruit          bool   `json:"isBigFruit"` // determine is fruit or not
		Music               string `json:"music"`      // music for frontend
	}

	BetZone struct {
		Total   int64  `json:"total"`   // 全部投注金额
		IsFruit bool   `json:"isFruit"` // determine is fruit or not
		Name    string `json:"name"`
		// Odds    int   `json:"odds"`    // 赔率
	}

	JoinRoomRequest struct {
		Uid        int64  `json:"uid"`
		Name       string `json:"name"`
		RoomNumber int64  `json:"roomNumber"`
		FaceUri    string `json:"faceUri"`
		Level      int    `json:"level"`
	}

	JoinRoomResponse struct {
		Name     string     `json:"name"`
		FaceUri  string     `json:"faceUri"`
		BetZones []*BetZone `json:"betZones"` // 押注块
	}

	Player struct {
		FaceUri  string // 头像地址
		UserName string // 玩家名字
		GameCoin int64
	}

	ResultPhaseResponse struct {
		Deadline             time.Time      `json:"deadline"`
		PersonalBetResult    BetResult      `json:"personalBetResult"`
		Awarded              int64          `json:"awarded"`
		GameCoin             int64          `json:"gameCoin"`
		Winners              []Winner       `json:"winners"`
		SelectedWinningItem  *WinningItem   `json:"selectedWinningItem"`
		SelectedWinningItems []*WinningItem `json:"selectedWinningItems"`
		UpdateRobotWinners   bool           `json:"updateRobotWinners"`
		WinningKey           int            `json:"winningKey"`
		SpecialPrize         string         `json:"specialPrize"`
	}

	AnimationPhaseResponse struct {
		Deadline              time.Time      `json:"deadline"`
		SelectedWinningItem   *WinningItem   `json:"selectedWinningItem"`
		SelectedWinningItems  []*WinningItem `json:"selectedWinningItems"`
		WinningKey            int            `json:"winningKey"`
		SpecialPrize          string         `json:"specialPrize"`
		WinningBetZoneIndexes []int          `json:"winningBetZoneIndexes"`
	}

	BettingPhaseResponse struct {
		Deadline time.Time `json:"deadline"`
	}

	SendMessageErrorResponse struct {
		Error string `json:"error"`
	}
)
