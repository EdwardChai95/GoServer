package game

const (
	// inactive game status
	GAMESTATUS_NOGAME  = "nogame"
	GAMESTATUS_WAITING = "waiting"

	// active game status
	GAMESTATUS_SEAT_UPDATE = "seatsUpdate" // receiver

	GAMESTATUS_PREFLOP  = "preflop"  // 翻牌前
	GAMESTATUS_FLOP     = "flop"     // 翻牌圈
	GAMESTATUS_TURN     = "turn"     // 转牌圈
	GAMESTATUS_RIVER    = "river"    // 河牌圈
	GAMESTATUS_SHOWDOWN = "showdown" // 摊牌

	// constants var
	ANIMATION_PLAYER1CARD     = 1
	ANIMATION_TIME_1CARD      = 2
	ANIMATION_TIME            = 2  // seconds
	ANIMATION_TIME_SHOWDOWN   = 3  // seconds
	PLAY_TIME                 = 10 // seconds // 10 seconds is requirement
	NUMBER_OF_SEATS           = 6
	NUMBER_OF_MINIMUM_PLAYERS = 3

	kCurPlayer = "player"

	LOG_GAMETITLE = "德州扑克"
)
