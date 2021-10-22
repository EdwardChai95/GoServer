package db

import (
	// "gitlab.com/wolfplus/gamespace-lobby/db/model"
	"log"
	"strconv"

	"gitlab.com/wolfplus/gamespace-lobby/helper"
)

//SendLogInformation creates a new lg information in the database
func SendLogInformation(reqJSON map[string]string) map[string]interface{} {
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
	task_key := returnCorrectString(reqJSON, "task_key", "") // new add

	affected3, err := db.Exec("INSERT INTO `log_information`"+
		"(`uid`,`reason`, `game`, `level`, `other_info`, `result`, `rate`, "+
		"`bet_total`, `win_total`, `banker_win_total`,`operating_time`,`params`, `before`, `used`, `after`, `task_key`) "+
		"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", uid, reason, game, level, otherInfo, result,
		rate, betTotal, winTotal, bankerWinTotal, helper.GetCurrentShanghaiTimeString(), params,
		before, used, after, task_key)
	if err != nil {
		log.Println(err)
	}
	logger.Println(affected3)
	return nil
}

//SendCodeInformation creates a new code information in the database
func SendCodeInformation(reqJSON map[string]string) map[string]interface{} {
	log.Println("测试！！！！", reqJSON)
	uid := returnCorrectString(reqJSON, "uid", "0")
	reason := returnCorrectString(reqJSON, "reason", "0")
	otherInfo := returnCorrectString(reqJSON, "otherInfo", "0")

	_, err := db.Exec("INSERT INTO `code_information`"+
		"(`uid`,`reason`,`other_info`,`operating_time`) "+
		"VALUES (?, ?, ?, ?)", uid, reason, otherInfo, helper.GetCurrentShanghaiTimeString())
	if err != nil {
		log.Println(err)
	}

	return nil
}

func returnCorrectString(dict map[string]string, key string, defaultValue string) string {
	newValue, ok := dict[key]
	if !ok {
		newValue = defaultValue
	} else if newValue == "" {
		newValue = defaultValue
	}
	//log.Println(key + ": " + newValue)
	return newValue
}

//NewWelfare creates a welfare entry if player has too little coins
func NewWelfare(uid string) map[string]interface{} {
	amountToUpdate := 500
	gameCoinFromUser, err := db.QueryString("select `game_coin` from `user` where `uid` = '" + uid + "' LIMIT 1")
	if err != nil {
		log.Println(err)
	}
	if len(gameCoinFromUser) > 0 {
		coins, err := strconv.Atoi(gameCoinFromUser[0]["game_coin"])
		if err != nil {
			log.Println(err)
		}
		coins-- //idiotadd1
		if coins <= 0 {
			welfareTimes, err := db.QueryString("select * from `log_information` where `uid` = '" + uid + "' and `reason` = '低保' and DATE_FORMAT(`operating_time`,'%Y-%m-%d') = CURDATE();")
			if err != nil {
				log.Println(err)
			}
			if len(welfareTimes) < 3 {
				//added one because after doing this, I'll have done the welfare once
				numWelfareToday := len(welfareTimes) + 1
				_, err := db.Exec("INSERT INTO `game_coin_transaction`(`uid`, `value`, `type`, `comment`, `datetime`) VALUES (?, ?, ?, ?, ?)", uid, amountToUpdate, "游戏币", "低保", helper.GetCurrentShanghaiTimeString())
				if err != nil {
					log.Println(err)
				}
				newCoinAmount := amountToUpdate + coins
				newCoinAmount++ //idiotadd1
				_, err = db.Exec("update `user` set `game_coin` = '" + strconv.Itoa(newCoinAmount) + "' where `uid` = '" + uid + "'")
				if err != nil {
					log.Println(err)
				}
				reqJSON := map[string]string{
					"uid":       uid,
					"reason":    "低保",
					"otherInfo": strconv.Itoa(amountToUpdate),
				}
				SendLogInformation(reqJSON)
				payload := map[string]interface{}{
					"numWelfareToday": numWelfareToday,
					"amountToUpdate":  amountToUpdate,
					"currentGameCoin": newCoinAmount,
				}
				return payload
			}
		}
	}
	return nil
}
