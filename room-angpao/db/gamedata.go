package db

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

// var LogInformations = []map[string]string{}

func UpdateGameCoinByUid(uid int64, updateAmt int64) {
	// users, err := db.QueryString("select * from user where uid = '" + strconv.Itoa(int(uid)) + "' LIMIT 1")
	// game_coin, err1 := strconv.Atoi(users[0]["game_coin"])
	// if err != nil {
	// 	log.Println(err)
	// }

	game_coin := GetGameCoinByUid(uid)

	if game_coin != 0 {
		affected1, err := db.Exec("update user set game_coin = ? where uid = ?", strconv.Itoa(int(game_coin)+int(updateAmt)), uid)
		if err != nil {
			log.Println(err)
		}
		log.Println(affected1)
	}
}

//add 1007
func UpdateWinGameCoinByUid(uid int64, updateAmt int64) {
	win_game_coin := GetWinGameCoinByUid(uid)

	if win_game_coin != 0 && updateAmt != 0 {
		_, err := db.Exec("update user set win_game_coin = ? where uid = ?", strconv.Itoa(int(win_game_coin)+int(updateAmt)), uid)
		if err != nil {
			log.Println(err)
		}
		// log.Infof("UpdateWinGameCoinByUid: &v", affected1)
	}
}

func GetGameCoinByUid(uid int64) int64 {
	users, err := db.QueryString("select * from user where uid = '" +
		strconv.Itoa(int(uid)) + "' LIMIT 1")
	if err != nil {
		log.Println(err)
	}
	if len(users) > 0 {
		gamecoin, err := strconv.ParseInt(users[0]["game_coin"], 10, 64)
		if err != nil {
			log.Println(err)
			return -1
		}
		return gamecoin
	}
	return -1
}

//add 1007
func GetWinGameCoinByUid(uid int64) int64 {
        users, err := db.QueryString("select * from user where uid = '" +
                strconv.Itoa(int(uid)) + "' LIMIT 1")
        if err != nil {
                log.Println(err)
        }
        if len(users) > 0 {
                wingamecoin, err := strconv.ParseInt(users[0]["win_game_coin"], 10, 64)
                if err != nil {
                        log.Println(err)
                        return -1
                }
                return wingamecoin
        }
        return -1
}

func NewGameCoinTransaction(uid int64, value int64) {
	if value == 0 {
		return
	}
	affected3, err := db.Exec("INSERT INTO `game_coin_transaction`"+
		"(`uid`, `value`, `type`, `comment`, `datetime`)"+
		" VALUES (?, ?, ?, ?, ?)",
		strconv.Itoa(int(uid)), strconv.Itoa(int(value)),
		"游戏币", "动物乐园", getCurrentShanghaiTimeString())
	if err != nil {
		log.Println(err)
	}
	log.Println(affected3)
}

//SendLogInformation creates a new lg information in the database
func SendLogInformation(reqJSON map[string]string) int64 {
	uid := returnCorrectString(reqJSON, "uid", "0")
	reason := returnCorrectString(reqJSON, "reason", "")
	game := returnCorrectString(reqJSON, "game", "")
	level := returnCorrectString(reqJSON, "level", "")
	otherInfo := returnCorrectString(reqJSON, "otherInfo", "")
	result := returnCorrectString(reqJSON, "result", "0")
	rate := returnCorrectString(reqJSON, "rate", "0")
	betTotal := returnCorrectString(reqJSON, "betTotal", "0")
	winTotal := returnCorrectString(reqJSON, "winTotal", "0")
	bankerWinTotal := returnCorrectString(reqJSON, "bankerWinTotal", "0")
	params := returnCorrectString(reqJSON, "params", "0")
	before := returnCorrectString(reqJSON, "before", "0")
	used := returnCorrectString(reqJSON, "used", "0")
	after := returnCorrectString(reqJSON, "after", "0")
	if uid == "0" {
		affected3, err := db.Exec("INSERT INTO `log_information_system`"+
			"(`uid`,`reason`, `game`, `level`, `other_info`, `result`, `rate`, "+
			"`bet_total`, `win_total`, `banker_win_total`,`operating_time`,`params`,`before`,`used`,`after`) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", uid, reason, game, level, otherInfo, result,
			rate, betTotal, winTotal, bankerWinTotal, getCurrentShanghaiTimeString(), params,
			before, used, after)

		if err != nil {
			logger.Warn(err)
			return -1
		}

		if id, err := affected3.LastInsertId(); err == nil {
			return id
		}
	} else {
		affected3, err := db.Exec("INSERT INTO `log_information`"+
			"(`uid`,`reason`, `game`, `level`, `other_info`, `result`, `rate`, "+
			"`bet_total`, `win_total`, `banker_win_total`,`operating_time`,`params`,`before`,`used`,`after`) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", uid, reason, game, level, otherInfo, result,
			rate, betTotal, winTotal, bankerWinTotal, getCurrentShanghaiTimeString(), params,
			before, used, after)

		if err != nil {
			logger.Warn(err)
			return -1
		}

		if id, err := affected3.LastInsertId(); err == nil {
			return id
		}
	}

	return -1
}

// helper
func returnCorrectString(dict map[string]string, key string, defaultValue string) string {
	newValue, ok := dict[key]
	if !ok {
		newValue = defaultValue
	}
	//log.Println(key + ": " + newValue)
	return newValue
}

func GetCurrentShanghaiTime() time.Time {
	loc, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	return time.Now().In(loc)
}

func getCurrentShanghaiTimeString() string {
	createdFormat := "2006-01-02 15:04:05"
	return GetCurrentShanghaiTime().Format(createdFormat)
}

func StringToInt64(str string) int64 {
	n, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		fmt.Printf("%d of type %T", n, n)
		fmt.Printf("StringToInt64 err %v", err)
	}
	return n
}

func StringToFloat64(str string) float64 {
	n, err := strconv.ParseFloat(str, 64)
	if err != nil {
		fmt.Printf("%d of type %T", n, n)
		fmt.Printf("StringToFloat64 err %v", err)
	}
	return n
}

func Int64ToString(n int64) string {
	return strconv.FormatInt(n, 10)
}
