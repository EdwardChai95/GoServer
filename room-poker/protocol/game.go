package protocol

import (
	"time"
)

type (
	ShowdownWinners struct {
		SeatNumber int
		WinAmount  int64
	}

	ShowdownBroadcast struct {
		Winners             []ShowdownWinners `json:"winners"`
		Deadline            time.Time         `json:"deadline"`
		PublicCards         []Card            `json:"publicCards"`
		SeatedPlayersUpdate interface{}       `json:"seatedPlayersUpdate"`
		PrizePool           int64             `json:"prizePool"`
	}

	RiverBroadcast struct {
		PublicCards []Card    `json:"publicCards"`
		Deadline    time.Time `json:"deadline"`
	}

	TurnBroadcast struct {
		PublicCards []Card    `json:"publicCards"`
		Deadline    time.Time `json:"deadline"`
	}

	FlopBroadcast struct {
		PublicCards []Card    `json:"publicCards"`
		Deadline    time.Time `json:"deadline"`
	}

	PlaceBetRequest struct {
		Amount int64 `json:"amount"`
	}

	SeatUpdateBroadcast struct {
		BettingAmount       int64       `json:"bettingAmount"`
		PrizePool           int64       `json:"prizePool"`
		SeatedPlayersUpdate interface{} `json:"seatedPlayersUpdate"`
		Deadline            time.Time   `json:"deadline"`
		GameStatus          string      `json:"gameStatus"`
	}

	PreflopBroadcast struct {
		SeatedPlayersUpdate interface{} `json:"seatedPlayersUpdate"`
		MyCards             []Card      `json:"myCards"`
		Deadline            time.Time   `json:"deadline"`
	}

	WaitingBroadcast struct {
		SeatedPlayersUpdate interface{} `json:"seatedPlayersUpdate"`
	}

	TakeSeatRes struct {
		SeatNumber          int         `json:"seatNumber"`
		UseableGameCoin     int64       `json:"useableGameCoin"`
		SeatedPlayersUpdate interface{} `json:"seatedPlayersUpdate"`
	}

	TakeSeatRequest struct {
		SeatNumber      int   `json:"seatNumber"`
		UseableGameCoin int64 `json:"useableGameCoin"`
	}

	Card struct {
		Value int    `json:"value"` // 2 - 14
		Suit  string `json:"suit"`  // s, h, c, d
		Face  string `json:"face"`  // (A, J, Q, K, 2 - 10)
		// FaceValue int    `json:"faceValue"`
	}

	Winner struct {
		Username string `json:"username"`
		Image    string `json:"image"`
		Level    int    `json:"level"`
		Awarded  int    `json:"awarded"`
	}

	RoomStatusResponse struct {
		GameStatus    string      `json:"gameStatus"`
		Deadline      time.Time   `json:"deadline"`
		SeatedPlayers interface{} `json:"seatedPlayersUpdate"`
		GameCoin      int64       `json:"gameCoin"`
	}

	JoinRoomRequest struct {
		Jwt        string `json:"jwt"`
		Name       string `json:"name"`
		RoomNumber int64  `json:"roomNumber"`
		FaceUri    string `json:"faceUri"`
		Level      int    `json:"level"`
	}

	JoinRoomResponse struct {
		Name         string `json:"name"`
		FaceUri      string `json:"faceUri"`
		SelectedRoom interface{}
	}

	Player struct {
		FaceUri  string // 头像地址
		UserName string // 玩家名字
		GameCoin int64
	}

	RobotPlayer struct {
		Uid      int64
		FaceUri  string
		UserName string
		GameCoin int64
	}

	SendMessageErrorResponse struct {
		Error string `json:"error"`
	}
)
