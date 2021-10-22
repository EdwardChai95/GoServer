package db

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"
)

var logInformations = []map[string]string{}
var logLock sync.Mutex

func RecordWinningItem(key int) {
	// currentTime := time.Now()
	affected, err := db.Exec("INSERT INTO `hch_winning_items`(`winning_key`, `create_at`) VALUES (?, ?)", strconv.Itoa(key), getCurrentShanghaiTimeString())
	if err != nil {
		log.Println(err)
	}
	log.Println(affected)
}

func CollectTax(totalPlayerTaxes int64) {
	currentTime := getCurrentShanghaiTime()
	// - retrieve latest record by largest id
	latestTax, err := db.QueryString("select * from `hch_tax_collection` ORDER BY `hch_tax_collection`.`tax_collection_id` DESC LIMIT 1")
	if err != nil {
		log.Println(err)
	}
	if len(latestTax) > 0 {
		// i, err := strconv.ParseInt(latestTax[0]["create_at"], 10, 64)
		// if err != nil {
		// 	log.Println(err)
		// 	return
		// }
		// taxCreated := time.Unix(i, 0) // convert from unix to time

		layout := "2018-09-19 18:26:32.000000"
		taxCreated, _ := time.Parse(layout, latestTax[0]["create_at"])
		// - compare unix of latest record and now
		if currentTime.Year() == taxCreated.Year() && currentTime.YearDay() == taxCreated.YearDay() {
			// - if same date do update (add into db)
			tax_collected, err := strconv.Atoi(latestTax[0]["tax_collected"])
			if err != nil {
				log.Println(err)
				return
			}
			affected1, err := db.Exec("update `hch_tax_collection` set `tax_collected` = ? where tax_collection_id = ?", strconv.Itoa(tax_collected+int(totalPlayerTaxes)), latestTax[0]["tax_collection_id"])
			if err != nil {
				log.Println(err)
			}
			log.Println(affected1)
			return
		}
	}
	// - if not insert tax
	affected, err := db.Exec("INSERT INTO `hch_tax_collection`(`tax_collected`, `create_at`) VALUES (?, ?)", strconv.Itoa(int(totalPlayerTaxes)), getCurrentShanghaiTimeString())
	if err != nil {
		log.Println(err)
	}
	log.Println(affected)
}

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

func NewGameCoinTransaction(uid int64, value int64) {
	affected3, err := db.Exec("INSERT INTO `game_coin_transaction`(`uid`, `value`, `type`, `comment`, `datetime`) VALUES (?, ?, ?, ?, ?)", strconv.Itoa(int(uid)), strconv.Itoa(int(value)), "游戏币", "豪车汇", getCurrentShanghaiTimeString())
	if err != nil {
		log.Println(err)
	}
	log.Println(affected3)
}

func NewLogInformation(logInformation map[string]string) {
	logLock.Lock()
	logInformations = append(logInformations, logInformation)
	logLock.Unlock()
}

func InsertAllLogInformations(gameLogInformation map[string]string, prizes string) {
	var wg sync.WaitGroup
	paramsInt := SendLogInformation(gameLogInformation)
	params := strconv.FormatInt(paramsInt, 10)

	if paramsInt == -1 {
		return
	}
	for _, logInformation := range logInformations {
		logInformation["otherInfo"] = "[开奖：" + prizes + "]" + logInformation["otherInfo"] + " [参数：" + params + "]"
		logInformation["params"] = params
		wg.Add(1)
		go func(logInformation map[string]string) {

			SendLogInformation(logInformation)
			wg.Done()
		}(logInformation)
	}

	wg.Wait()
	logInformations = []map[string]string{}
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
	tax := returnCorrectString(reqJSON, "tax", "0")

	if uid == "0" {
		affected5, err := db.Exec("INSERT INTO `log_information_system`"+
			"(`uid`,`reason`, `game`, `level`, `other_info`, `result`, `rate`, "+
			"`bet_total`, `win_total`, `banker_win_total`,`operating_time`,`params`,`before`,`used`,`after`, `tax`) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", uid, reason, game, level, otherInfo, result,
			rate, betTotal, winTotal, bankerWinTotal, getCurrentShanghaiTimeString(), params,
			before, used, after, tax)
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
			"`bet_total`, `win_total`, `banker_win_total`,`operating_time`,`params`,`before`,`used`,`after`, `tax`) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", uid, reason, game, level, otherInfo, result,
			rate, betTotal, winTotal, bankerWinTotal, getCurrentShanghaiTimeString(), params,
			before, used, after, tax)
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
