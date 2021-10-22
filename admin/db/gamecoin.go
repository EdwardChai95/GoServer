package db

import (
	"admin/helper"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func UpdateGameCoinByAdmin(adminUid string, data map[string][]string) error { // requires data["update_amt"][0], data["comment"][0]
	// check data["update_amt"][0]
	if len(data["update_amt"]) == 0 {
		return errors.New("修改失败")
	}

	s := strings.Split(data["uid"][0], ",")
	if update_amt, err := strconv.Atoi(data["update_amt"][0]); err == nil {
		comment := "管理员加"
		command := adminUid + "批量添加了: " + data["update_amt"][0]

		for _, v := range s {
			affected1, err := db.Exec("update user set game_coin = game_coin + ? where uid = ?",
				update_amt, v)
			if err != nil {
				logger.Println(err)
				return err
			}
			message, err1 := db.QueryString("select * from `user` where uid = ? ", v)
			if err1 != nil {
				logger.Println(err1)
				return err1
			}
			fmt.Println("gamecoin:", message[0]["game_coin"])
			numRowsAffected, _ := affected1.RowsAffected()
			if numRowsAffected > 0 {
				err = NewGameCoinTransaction1(helper.StringToInt64(v), data["update_amt"][0])
				if err != nil {
					logger.Println(err)
					return err
				}
				newLogInformation1(map[string]string{
					"uid":        fmt.Sprintf("%v", v),
					"reason":     comment,
					"game":       "",
					"level":      "",
					"before":     message[0]["game_coin"],
					"used":       data["update_amt"][0],
					"after":      fmt.Sprintf("%v", helper.StringToInt(message[0]["game_coin"])+update_amt),
					"other_info": command,
				})

			}
		}
	} else {
		return errors.New("加减值格式编号不正确")
	}
	return nil
}

func newLogInformation1(data map[string]string) {
	cols := ""
	vals := ""

	for col, val := range data {
		cols += fmt.Sprintf("`%v`, ", col)
		vals += fmt.Sprintf("'%v', ", val)
	}

	cols += "`operating_time`"
	vals += fmt.Sprintf("'%v'", helper.GetCurrentShanghaiTimeString())

	sql := fmt.Sprintf("INSERT INTO `log_information` (%v) VALUES (%v)", cols, vals)
	// logger.Printf("sql: %v", sql)
	_, err := db.Exec(sql)

	if err != nil {
		logger.Warn(err)
	}
}

func NewGameCoinTransaction1(uid int64, value string) error {
	affected3, err := db.Exec("INSERT INTO `game_coin_transaction` (`uid`, `value`, `type`, `comment`, `datetime`) VALUES (?, ?, ?, ?, ?)",
		strconv.Itoa(int(uid)), value, "游戏币", "", helper.GetCurrentShanghaiTimeString())
	if err != nil {
		logger.Warn(err)
		return err
	}
	logger.Println(affected3)
	return nil
}
