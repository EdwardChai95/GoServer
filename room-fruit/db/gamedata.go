package db

import (
	"fmt"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

// var logInformations = []map[string]string{}

func RecordWinningItem(key int) {
	_, err := db.Exec("INSERT INTO `fruit_winning_item`(`winning_key`, `create_at`) VALUES (?, ?)", strconv.Itoa(key), getCurrentShanghaiTimeString())
	if err != nil {
		log.Println(err)
	}
	// log.Println(affected)
}

// update the game coin of a uid
// NOTE this is SETTING OF THE GAMECOIN
// e.g. uid 123 gamecoin = 10000
func UpdateGameCoinByUid(uid int64, updateAmt int64) {
	game_coin := GetGameCoinByUid(uid)

	if game_coin != 0 && updateAmt != 0 {
		_, err := db.Exec("update user set game_coin = ? where uid = ?", strconv.Itoa(int(game_coin)+int(updateAmt)), uid)
		if err != nil {
			log.Println(err)
		}
		// log.Infof("UpdateGameCoinByUid: &v", affected1)
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

// get the gamecoin of a uid
func GetGameCoinByUid(uid int64) int64 {
	users, err := db.QueryString("select * from user where uid = '" + strconv.Itoa(int(uid)) + "' LIMIT 1")
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
	users, err := db.QueryString("select * from user where uid = '" + strconv.Itoa(int(uid)) + "' LIMIT 1")
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

func GetClubByUid(uid int64) map[string]string {
	users, err := db.QueryString("select * from club_user where uid = '" + strconv.Itoa(int(uid)) + "' LIMIT 1")
	if err != nil {
		log.Println(err)
	}

	if len(users) > 0 {
		return users[0]
	}
	return nil
}

// log for whenever you ADD or SUBTRACT game coin
// e.g. + 100, - 200 etc
func NewGameCoinTransaction(uid int64, value int64) {
	if value != 0 {
		_, err := db.Exec("INSERT INTO `game_coin_transaction`(`uid`, `value`, `type`, `comment`, `datetime`) VALUES (?, ?, ?, ?, ?)",
			strconv.Itoa(int(uid)),
			strconv.Itoa(int(value)),
			"游戏币",
			"水果",
			getCurrentShanghaiTimeString())
		if err != nil {
			log.Println(err)
		}
		// log.Infof("NewGameCoinTransaction: &v", affected)
	}
}

//SendLogInformation creates a new log information in the database
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
		affected5, err := db.Exec("INSERT INTO `log_information_system`"+
			"(`uid`,`reason`, `game`, `level`, `other_info`, `result`, `rate`, "+
			"`bet_total`, `win_total`, `banker_win_total`,`operating_time`,`params`,`before`,`used`,`after`) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", uid, reason, game, level, otherInfo, result,
			rate, betTotal, winTotal, bankerWinTotal, getCurrentShanghaiTimeString(), params,
			before, used, after)
		if err != nil {
			logger.Warn(err)
			return -1
		}

		if id, err := affected5.LastInsertId(); err == nil {
			return id
		}
	} else if uid == "-1" {
		affected4, err := db.Exec("INSERT INTO `log_information_robot`"+
			"(`uid`,`reason`, `game`, `level`, `other_info`, `result`, `rate`, "+
			"`bet_total`, `win_total`, `banker_win_total`,`operating_time`,`params`,`before`,`used`,`after`) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", uid, reason, game, level, otherInfo, result,
			rate, betTotal, winTotal, bankerWinTotal, getCurrentShanghaiTimeString(), params,
			before, used, after)
		if err != nil {
			logger.Warn(err)
			return -1
		}

		if id, err := affected4.LastInsertId(); err == nil {
			return id
		}
	} else {
		user, err := db.QueryString("select * from `user` where uid = '" + uid + "' LIMIT 1")
		if user[0]["normal_active"] == "0" {
			_, err3 := db.Exec("Update `user` set `normal_active` = 1 where uid = '" + uid + "' ")
			if err3 != nil {
				logger.Warn(err3)
			}
		}
		if user[0]["proxy"] != "0" {
			proxyuser, err := db.QueryString("select * from `proxy_user` where date(operating_time) >= curdate() and uid = '" + uid + "' LIMIT 1")
			if err != nil {
				log.Println(err)
			}
			if len(proxyuser) > 0 {
				if StringToInt64(used) < 0 {
					_, err := db.Exec("Update `proxy_user` set `total_broad` = `total_broad` + 1, `total_lose` = `total_lose` + '" +
						used + "', `total_win_lose` = `total_win_lose` + '" + used + "'  where uid = '" + uid + "'")
					if err != nil {
						log.Println(err)
					}
				} else {
					_, err := db.Exec("Update `proxy_user` set `total_broad` = `total_broad` + 1, `total_win` = `total_win` + '" +
						used + "', `total_win_lose` = `total_win_lose` + '" + used + "'  where uid = '" + uid + "'")
					if err != nil {
						log.Println(err)
					}
				}
			} else {
				if StringToInt64(used) < 0 {
					_, err := db.Exec("INSERT INTO `proxy_user`(`uid`, `proxy_uid`, `total_win`, `total_lose`, `total_win_lose`, `total_broad`, `send_num`, `receive_num`, `total_amount`, `count_completed`, `operating_time`)"+
						"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", uid, user[0]["proxy"], 0, used, used, 1, 0, 0, 0, 0, getCurrentShanghaiTimeString())
					if err != nil {
						log.Println(err)
					}
				} else {
					_, err := db.Exec("INSERT INTO `proxy_user`(`uid`, `proxy_uid`, `total_win`, `total_lose`, `total_win_lose`, `total_broad`, `send_num`, `receive_num`, `total_amount`, `count_completed`,  `operating_time`)"+
						"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", uid, user[0]["proxy"], used, 0, used, 1, 0, 0, 0, 0, getCurrentShanghaiTimeString())
					if err != nil {
						log.Println(err)
					}
				}

			}
		}

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

// helper functions below:

func returnCorrectString(dict map[string]string, key string, defaultValue string) string {
	newValue, ok := dict[key]
	if !ok {
		newValue = defaultValue
	} else if newValue == "undefined" {
		newValue = defaultValue
	}
	//log.Println(key + ": " + newValue)
	return newValue
}

func getCurrentShanghaiTime() time.Time {
	loc, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	return time.Now().In(loc)
}

func getCurrentShanghaiTimeString() string {
	createdFormat := "2006-01-02 15:04:05"
	return getCurrentShanghaiTime().Format(createdFormat)
}

func StringToInt64(str string) int64 {
	n, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		fmt.Printf("%d of type %T", n, n)
		fmt.Printf("StringToInt64 err %v", err)
	}
	return n
}
