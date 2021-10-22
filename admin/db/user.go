package db

import (
	"admin/helper"
	"errors"
	"fmt"
	"strconv"
)

func UpdateGameCoinByUid(uid int64, adminUid string, data map[string][]string) error { // requires data["update_amt"][0], data["comment"][0]
	// check data["update_amt"][0]
	if len(data["update_amt"]) == 0 {
		return errors.New("修改失败")
	}

	if update_amt, err := strconv.Atoi(data["update_amt"][0]); err == nil {
		// logger.Printf("%q looks like a number.\n", data["update_amt"][0])

		comment := "管理员减"
		if update_amt > 0 {
			comment = "管理员加"
		}

		command := adminUid + " 备注了: " + data["comment"][0]

		user := GetUserByUid(fmt.Sprintf("%v", uid))

		affected1, err := db.Exec("update user set game_coin = game_coin + ? where uid = ?",
			update_amt, uid)
		if err != nil {
			logger.Println(err)
			return err
		}
		// logger.Println(affected1)
		numRowsAffected, _ := affected1.RowsAffected()
		if len(data["comment"]) == 0 || data["comment"][0] == "" {
			data["comment"][0] = comment
		}
		if numRowsAffected > 0 {
			err = NewGameCoinTransaction(uid, data["update_amt"][0], data["comment"][0])
			if err != nil {
				logger.Println(err)
				return err
			}
			newLogInformation(map[string]string{
				"uid":        fmt.Sprintf("%v", uid),
				"reason":     comment,
				"game":       "",
				"level":      "",
				"before":     user["game_coin"],
				"used":       data["update_amt"][0],
				"after":      fmt.Sprintf("%v", helper.StringToInt(user["game_coin"])+update_amt),
				"other_info": command,
			})

		}
	} else {
		return errors.New("加减值格式编号不正确")
	}
	return nil
}

func newLogInformation(data map[string]string) {
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

func NewGameCoinTransaction(uid int64, value, comment string) error {
	affected3, err := db.Exec("INSERT INTO `game_coin_transaction` (`uid`, `value`, `type`, `comment`, `datetime`) VALUES (?, ?, ?, ?, ?)",
		strconv.Itoa(int(uid)), value, "游戏币", comment, helper.GetCurrentShanghaiTimeString())
	if err != nil {
		logger.Warn(err)
		return err
	}
	logger.Println(affected3)
	return nil
}

const (
	sum_win_total             string = "COALESCE(SUM(`result`),0)"
	log_where_query           string = "WHERE `uid`=u.`uid` AND `game` !=''"
	nested_query              string = "(SELECT %v FROM `log_information` " + log_where_query + " LIMIT 1) as '%v'"
	nested_withwhere_query    string = "(SELECT %v FROM `log_information` " + log_where_query + "%v LIMIT 1) as '%v'"
	totalPurchaseAmount_query string = "(SELECT COALESCE(SUM(`payment_amount`),0) as sum_results FROM `order` " +
		"WHERE `order_status` = 'đã thanh toán' AND `uid`=u.`uid` LIMIT 1) as 'totalPurchaseAmount' "
)

func GetUserByUid(uid string) map[string]string {
	//users, err := db.QueryString("select * from `user` where `uid` = '" + uid + "' Limit 1")
	rounds_played_query := fmt.Sprintf(nested_query, "COUNT(*)", "rounds_played")
	total_played_query := fmt.Sprintf(nested_query, sum_win_total, "total_played")
	total_win_query := fmt.Sprintf(nested_withwhere_query, sum_win_total, " AND `result` > '0'", "total_win")
	total_lose_query := fmt.Sprintf(nested_withwhere_query, sum_win_total, " AND `result` < '0'", "total_lose")
	select_query := fmt.Sprintf("select u.*, %v, %v, %v, %v, %v from `user` u where u.`uid` = '"+uid+"' Limit 1",
		rounds_played_query, total_played_query, total_win_query, total_lose_query, totalPurchaseAmount_query)

	users, err := db.QueryString(select_query)

	if err != nil {
		logger.Error(err)
	}
	// logger.Print("sql", "select * from `user` where `uid` = '"+uid+"' Limit 1")
	// logger.Print("Users: %v", users)

	if len(users) > 0 {
		return users[0]
	}
	return nil
}

func UpdateUserByUid(uid string, data map[string][]string) bool {
	data["password"] = append(data["password"], "isPassword")
	data["uid"] = append(data["uid"], "toIgnore")
	// add special element to tell the sql builder how to update this var

	sqlStr := "UPDATE user SET "
	sqlStr += helper.SQLUpdateDataStr(data)
	sqlStr += "WHERE uid='" + uid + "'"
	logger.Print(sqlStr)
	_, err := db.Exec(sqlStr)
	if err != nil {
		logger.Error(err)
		return false
	}
	return true
}

func GetUsersByPageNo(pagenumber string, urlParams string, searchParams map[string]string) map[string]interface{} {
	offset := (helper.StringToInt(pagenumber) - 1) * helper.RECORDS_PER_PAGE
	//	conditionQuery := " LEFT JOIN `game_coin_locker` ON u.`uid` = `game_coin_locker`.`uid` "
	//	conditionQuery += " LEFT JOIN `log_information` li ON u.`uid` = li.`uid` AND li.`game`<>'' "
	//	conditionQuery += " LEFT JOIN `order` o ON o.`uid` = u.`uid` AND o.`order_status` = '已付款' "
	conditionQuery := " WHERE 1=1 "

	/*for col, val := range searchParams {
		if strings.TrimSpace(val) != "" {
			conditionQuery += fmt.Sprintf(" AND `user`.`%v`='%v' ", col, val)
		}
	}*/

	if searchParams["uid"] != "" {
		conditionQuery += "AND u.`uid` = '" + searchParams["uid"] + "' "
	}
	/*	if searchParams["userType"] != "" {
			if searchParams["userType"] == fmt.Sprintf("%v", UserTypeAdmin) {
				conditionQuery += "AND u.`user_permission` = '" + searchParams["userType"] + "' "
			} else {
				conditionQuery += "AND u.`user_permission` != '" + fmt.Sprintf("%v", UserTypeAdmin) + "' "
			}

		}
	*/
	if searchParams["userType"] != "" {
		if searchParams["userType"] == fmt.Sprintf("%v", UserTypeAdmin) {
			conditionQuery += "AND u.`user_permission` = '" + searchParams["userType"] + "' "
		} else if searchParams["userType"] == fmt.Sprintf("%v", UserTypeProxy) {
			conditionQuery += "AND u.`user_permission` = '" + searchParams["userType"] + "' "
		} else {
			conditionQuery += "AND u.`user_permission` = '' "
		}

	}

	if searchParams["userAcc1"] != "" {
		conditionQuery += "AND u.`user_acc` = '" + searchParams["userAcc1"] + "' "
	}
	if searchParams["userAcc"] != "" {
		if searchParams["userAcc"] == fmt.Sprintf("%v", UserAccsFalse) {
			conditionQuery += "AND u.`user_acc` = '1' "
		} else {
			conditionQuery += "AND u.`user_acc` != '1' "
		}
	}
	if searchParams["levelStart"] != "" {
		conditionQuery += "AND u.`level` >= '" + searchParams["levelStart"] + "' "
	}
	if searchParams["levelEnd"] != "" {
		conditionQuery += "AND u.`level` <= '" + searchParams["levelEnd"] + "' "
	}

	if searchParams["dateStartLogin"] != "" {
		conditionQuery += "AND `last_login_at` >= '" + searchParams["dateStartLogin"] + "' "
	}
	if searchParams["dateEndLogin"] != "" {
		conditionQuery += "AND `last_login_at` <= '" + searchParams["dateEndLogin"] + "' "
	}
	if searchParams["dateStartRegister"] != "" {
		conditionQuery += "AND `create_at` >= '" + searchParams["dateStartRegister"] + "' "
	}
	if searchParams["dateEndRegister"] != "" {
		conditionQuery += "AND `create_at` <= '" + searchParams["dateEndRegister"] + "' "
	}
	//conditionQuery += " GROUP BY u.`uid`"

	COOKIENAME := "USERLISTING" + urlParams
	totalRecords := 0

	if cookietotalStr := Ctx.GetCookie(COOKIENAME); cookietotalStr != "" { // total found!
		totalRecords = helper.StringToInt(cookietotalStr)
	} else {
		// logger.Printf("query: %v", "SELECT count(*) as total from `user` "+conditionQuery+" LIMIT 1")
		total, err := db.QueryString("SELECT count(distinct u.uid) as total from `user` u " + conditionQuery + " LIMIT 1")
		if err != nil {
			logger.Error(err)
		}
		if len(total) > 0 {
			totalRecords = helper.StringToInt(total[0]["total"])
			Ctx.SetCookieKV(COOKIENAME, total[0]["total"])
		}
	}

	//users, err := db.QueryString("select user.*, game_coin_locker.balance as game_coin_locker from `user`" +
	select_str := "select u.uid, u.nick_name, u.level, " +
		"u.game_coin, u.user_acc, u.user_permission, " +
		"u.create_at, u.last_login_at, u.create_os, u.last_login_os,u.normal_active " +
		//"COUNT(distinct li.log_information_id) as 'rounds_played', " +
		//" COALESCE(SUM(o.`payment_amount`),0) as totalPurchaseAmount," +
		//" COALESCE(MAX(li.`bet_total`), 0) as max_bet_total," +
		//" game_coin_locker.balance as game_coin_locker from `user` u " +
		"from `user` u " +
		conditionQuery +
		" AND `user_permission` != 'super_admin' " +
		"GROUP BY u.`uid` " +
		" ORDER BY `create_at` DESC " +
		" Limit " +
		strconv.Itoa(helper.RECORDS_PER_PAGE) +
		" OFFSET " +
		strconv.Itoa(offset)

	logger.Println(select_str)
	users, err := db.QueryString(select_str)
	if err != nil {
		logger.Error(err)
	}

	/*	select_str1 := "select u.*, COUNT(distinct li.log_information_id) as 'rounds_played', " +
		" COALESCE(SUM(o.`payment_amount`),0) as totalPurchaseAmount," +
		" COALESCE(MAX(li.`bet_total`), 0) as max_bet_total," +
		" game_coin_locker.balance as game_coin_locker from `user` u " +
		conditionQuery +
		" AND `user_permission` != 'super_admin' " +
		"GROUP BY u.`uid` " +
		" ORDER BY `create_at` DESC "
	*/
	select_str1 := "select u.uid, u.nick_name, u.level, " +
		"u.game_coin, u.user_acc, u.user_permission, " +
		"u.create_at, u.last_login_at, u.create_os, u.last_login_os,u.normal_active " +
		"from `user` u " +
		conditionQuery +
		" AND `user_permission` != 'super_admin' " +
		"GROUP BY u.`uid` " +
		" ORDER BY `create_at` DESC "

	usersToExport, err := db.QueryString(select_str1)
	if err != nil {
		logger.Error(err)
	}

	if len(users) > 0 && totalRecords > 0 {
		numPages := totalRecords / helper.RECORDS_PER_PAGE

		if totalRecords%helper.RECORDS_PER_PAGE > 0 {
			numPages++
		}

		for _, v := range users {
			/*
				v["is_activated"] = "0"
				if helper.StringToInt(v["rounds_played"]) > 0 { //&& // 2：用户玩了一局以上游戏；
					//helper.StringToInt(v["totalPurchaseAmount"]) > 0 && // 3：用户有充值；
					//helper.StringToInt(v["max_bet_total"]) >= 50000 && // 4：用户单次压注金额超过5万；
					//v["user_acc"] != "1" { // 1：用户绑定手机；
					v["is_activated"] = "1"
				}
			*/
			if v["user_permission"] == "admin" {
				v["user_permission"] = "管理员"
			} else if v["user_permission"] == "proxy_admin" {
				v["user_permission"] = "代理"
			} else {
				v["user_permission"] = "玩家"
			}
			if v["user_acc"] == "1" {
				v["user_acc"] = "没绑定"
			}
			if v["game_coin_locker"] == "" {
				v["game_coin_locker"] = "没绑定"
			}
			v["create_at"] = helper.DisplayDate(v["create_at"])
			v["last_login_at"] = helper.DisplayDate(v["last_login_at"])
		}

		for _, v := range usersToExport {
			if v["user_permission"] == "admin" {
				v["user_permission"] = "管理员"
			} else {
				v["user_permission"] = "玩家"
			}
			if v["user_acc"] == "1" {
				v["user_acc"] = "没绑定"
			}
			if v["game_coin_locker"] == "" {
				v["game_coin_locker"] = "没绑定"
			}
			v["create_at"] = helper.DisplayDate(v["create_at"])
			v["last_login_at"] = helper.DisplayDate(v["last_login_at"])
		}

		return map[string]interface{}{
			"Users":         users,
			"UsersToExport": usersToExport,
			"Total":         totalRecords,
			"NumPages":      numPages,
		}
	}

	return nil
}

const (
	UserTypeAdmin  string = "admin"
	UserTypeProxy  string = "proxy_admin"
	UserTypePlayer string = "player"
)

var UserTypes = []map[string]interface{}{
	{"val": "", "text": "全部"},
	{"val": UserTypeAdmin, "text": "管理员"},
	{"val": UserTypeProxy, "text": "代理"},
	{"val": UserTypePlayer, "text": "玩家"},
}

const (
	UserAccsFalse string = "1"
	UserAccsTrue  string = "2"
)

var UserAccs = []map[string]interface{}{
	{"val": "", "text": "全部"},
	{"val": UserAccsFalse, "text": "未绑定"},
	{"val": UserAccsTrue, "text": "已绑定"},
}
