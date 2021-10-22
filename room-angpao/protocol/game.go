package protocol

import (
	"time"
)

type (
	RobotPlayer struct {
		Uid      int64
		FaceUri  string
		UserName string
		GameCoin int64
	}

	Player struct {
		FaceUri  string // 头像地址
		UserName string // 玩家名字
		GameCoin int64
		Level    int
	}

	Winner struct {
		Username string `json:"username"`
		Awarded  int    `json:"awarded"`
	}

	ResultPhaseResponse struct {
		GameCoin   int64
		Deadline   time.Time
		Index      int64    `json:"index"`
		Winners    int64    `json:"winners"`
		Min        int64    `json:"min"`
		AngPaoList []AngPao `json:"AngPaoList"`
	}

	AnimationPhaseResponse struct {
		Deadline    time.Time
		Index       int64         `json:"index"`
		AngPaoList  []AngPao      `json:"AngPaoList"`
		HistoryList []HistoryItem `json:"historylist"` // history
	}

	PlayingPhaseResponse struct {
		Deadline time.Time
		Auto     int64 `json:"auto"`
	}

	AutoAngPaoRequest struct {
		RoomNumber int64 `json:"roomNumber"`
	}

	AutoAngPaoResponse struct {
		Uid  int64  `json:"uid"`
		Auto int64  `json:"auto"`
		Tips string `json:"tips"`
	}

	GetAngPaoResponse struct {
		Uid        int64    `json:"uid"`
		Total      int64    `json:"total"`
		Left       int64    `json:"left"`
		AngPaoList []AngPao `json:"AngPaoList"`
	}

	GetAngPaoRequest struct {
		RoomNumber int64  `json:"roomNumber"`
		Status     string `json:"status"`
	}

	AngPao struct {
		Index    int64  `json:"index"`
		Uid      int64  `json:"uid"`
		UserName string `json:"UserName"`
		FaceUri  string `json:"faceUri"`
		Amount   int64  `json:"amount"`
	}

	RoomStatus struct {
		GameStatus  string
		Deadline    time.Time
		GameCoin    int64         `json:"gameCoin"`
		HistoryList []HistoryItem `json:"historylist"` // history
	}

	KickPlayerResponse struct {
		Uid   int64  `json:"uid"`
		Error string `json:"error"`
	}

	AutoAnimationResponse struct {
		Uid      int64 `json:"uid"`
		Deadline time.Time
	}

	HistoryItem struct {
		Id     int64 `json:"id"`
		Amount int64 `json:"amount"`
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
	}

	JoinRoomRequest struct {
		Uid        int64  `json:"uid"`
		Name       string `json:"name"`
		RoomNumber int64  `json:"roomNumber"`
		FaceUri    string `json:"faceUri"`
	}

	RoomItem struct {
		Level       string `json:"level"`
		Name        string `json:"name"`
		MinGameCoin int64  `json:"minGameCoin"`
		Icon        string `json:"icon"`
	}

	GetRoomResponse struct {
		Rooms []RoomItem `json:"rooms"`
	}
)
