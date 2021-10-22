package db

import (
	"admin/helper"
	"fmt"
)

func GetTopupReport(searchParams map[string]string) []map[string]string {

	//offset := (helper.StringToInt(pagenumber) -1) * helper.RECORDS_PER_PAGE

	isFirstQuery := "SELECT COUNT(*) FROM `order` o " +
		"WHERE o.`uid`=`user`.`uid` AND o.`updated_datetime` < `order`.`updated_datetime` " +
		"AND o.`order_status` = 'đã thanh toán' LIMIT 1"
	queryStr := "SELECT "

	queryStr += "`user`.`uid`, "               // uid
	queryStr += "`user`.`create_at`, "         // 注册时间
	queryStr += "`order`.`updated_datetime`, " // 充值时间
	queryStr += "`order`.`payment_amount`, "   // 充值金额
	queryStr += "`order`.`game_coin_amount`, " // 充值数量
	queryStr += "CASE `order`.`order_status` "
	queryStr += "WHEN 'tật nguyền' then '订单失效' "
	queryStr += "WHEN 'Chế biến' then '待付款' "
	queryStr += "WHEN 'đã thanh toán' then '已付款' "
	queryStr += "END as order_status, "
	queryStr += "IF("
	queryStr += "(" // 是否首充
	queryStr += isFirstQuery
	queryStr += ")"
	queryStr += " = 0 AND `order`.`order_status` = 'đã thanh toán', \"是\", \"否\") as 'isFirst' "

	queryStr += "FROM `order` JOIN `user` ON `user`.uid = `order`.uid "
	queryStr += "WHERE "

	if searchParams["dateStart"] != "" && searchParams["dateEnd"] != "" {
		queryStr += "`user`.`create_at` BETWEEN '" + searchParams["dateStart"] + "' AND '" + searchParams["dateEnd"] + "' "
		queryStr += "AND `order`.`updated_datetime` BETWEEN '" + searchParams["dateStart"] + "' AND '" + searchParams["dateEnd"] + "' "
	} else {
		queryStr += "DATE(`user`.`create_at`) = DATE('" + helper.GetCurrentShanghaiTimeString() + "') "
	}
	if searchParams["uid"] != "" {
		queryStr += "AND `order`.`uid` = '" + searchParams["uid"] + "' "
	}

	if searchParams["topupReportType"] != "" {
		if searchParams["topupReportType"] == TopUpReportIsFirst {
			queryStr += "AND (" + isFirstQuery + ") = 0 AND `order`.`order_status` = 'đã thanh toán' "
		} else if searchParams["topupReportType"] == TopUpReportIsNotFirst {
			queryStr += "AND (" + isFirstQuery + ") > 0 AND `order`.`order_status` = 'đã thanh toán' "
		}
	}
	queryStr += "ORDER BY `user`.`create_at` DESC "

	reports, err := db.QueryString(queryStr)
	if err != nil {
		logger.Printf(queryStr)
		logger.Error(err)
	}

	// logger.Printf("reports: %v", reports)

	if len(reports) > 0 {
		return reports
	}

	return nil
}

func GetStatReport(searchParams map[string]string) []map[string]string {

	queryStr := "SELECT r.*, (haochehui_tax + pinshi_tax) as 'total_tax', "

	queryStr += fmt.Sprintf("(SELECT COALESCE(SUM(`game_coin`), '未统计') as sum_results FROM `total_game_coin` " +
		"WHERE DATE(`create_at`) = r.create_at) as 'totalPlayerGamecoin' ")
            
	queryStr += "FROM `report` r "

	queryStr += "WHERE "

	if searchParams["dateStart"] != "" && searchParams["dateEnd"] != "" {
		queryStr += "r.create_at >= DATE('" + searchParams["dateStart"] + "') "
		queryStr += "AND r.create_at <= DATE('" + searchParams["dateEnd"] + "') "
	} else {
		queryStr += "r.create_at = DATE('" + helper.GetCurrentShanghaiTimeString() + "') "
	}
	queryStr += "ORDER BY r.create_at DESC"

	reports, err := db.QueryString(queryStr)
	logger.Println("queryStr of report")
	logger.Println(queryStr)

	if err != nil {
		logger.Printf(queryStr)
		logger.Error(err)
	}

	if len(reports) > 0 {
		//for _, v := range reports {
		//	if v["game_coin"] == "" {
		//		v["totalPlayerGamecoin"] = "未统计"
		//	}
		//}
		//return reports, games
		return reports
	}
	//return nil, nil
	return nil
	// logger.Printf("queryStr: %v", queryStr)
}

func GetCurrentReport() map[string]interface{} {
/*	queryStr := "SELECT " +
		    "(SELECT COUNT(*) as num_results FROM log_information WHERE reason = '注册奖励' AND DATE(operating_time) = dates.date1 LIMIT 1) as register_num, " +
                    "(SELECT COUNT(*) as num_results FROM log_information WHERE reason = '登陆' AND DATE(operating_time) = dates.date1 LIMIT 1) as login_num, " +
        "(SELECT COALESCE(SUM(payment_amount), 0) as sum_results FROM gamespace.order WHERE order_status = 'đã thanh toán' AND DATE(updated_datetime) = dates.date1 LIMIT 1) as total_payment_amount, " +
                    "(SELECT COALESCE(SUM(bet_total), 0) as sum_results FROM log_information WHERE uid > 0 AND game != '' AND DATE(operating_time) = dates.date1 LIMIT 1) as total_bet, " +
		    "(SELECT COUNT(*) as sum_results FROM log_information WHERE uid > 0 AND game != '' AND DATE(operating_time) = dates.date1 LIMIT 1) as bet_num, " +
		    "(SELECT COALESCE(SUM(used), 0) as sum_results FROM log_information WHERE uid > 0 AND game != '' AND DATE(operating_time) = dates.date1 LIMIT 1) as total_win_lose " +
		    "FROM(  " +
		    "SELECT DISTINCT DATE(updated_datetime) as date1 " +
		    "FROM gamespace.order " +
		    "UNION " +
		    "SELECT DISTINCT DATE(operating_time) as date1 " +
		    "FROM gamespace.log_information " +
		    ") dates  " +
		    "WHERE dates.date1 = DATE(curdate() + INTERVAL 0 DAY) "

	condition := "CASE WHEN " +
		"(TIME(curtime()) >= '01:00:00' AND TIME(curtime()) <= '23:59:59') " +
        	"THEN (operating_time >= curdate() - INTERVAL 0 DAY + INTERVAL 1 HOUR " +
		"AND operating_time <= curdate() + INTERVAL 1 DAY + INTERVAL 59 MINUTE + INTERVAL 59 SECOND) " +
		"ELSE (operating_time >= curdate() - INTERVAL 1 DAY + INTERVAL 1 HOUR " +
		"AND operating_time < curdate() + INTERVAL 0 DAY + INTERVAL 59 MINUTE + INTERVAL 59 SECOND) " +
		"END "
*/
	condition := "DATE(operating_time) = DATE(now()) "

	queryStr1 := "SELECT COUNT(*) as register_num FROM log_information WHERE reason = '注册奖励' AND " + condition
	queryStr2 := "SELECT COUNT(*) as login_num FROM log_information WHERE reason = '登陆' AND " + condition	
	/*
	condition2 := "CASE WHEN " +
                "(TIME(curtime()) >= '01:00:00' AND TIME(curtime()) <= '23:59:59') " +
                "THEN (updated_datetime >= curdate() - INTERVAL 0 DAY + INTERVAL 1 HOUR " +
                "AND updated_datetime <= curdate() + INTERVAL 1 DAY + INTERVAL 59 MINUTE + INTERVAL 59 SECOND) " +
                "ELSE (updated_datetime >= curdate() - INTERVAL 1 DAY + INTERVAL 1 HOUR " +
                "AND updated_datetime < curdate() + INTERVAL 0 DAY + INTERVAL 59 MINUTE + INTERVAL 59 SECOND) " +
                "END "
       	*/
	condition2 := " DATE(updated_datetime) = DATE(now()) "

	queryStr3 := "SELECT COALESCE(SUM(payment_amount), 0) as total_payment_amount FROM gamespace.order WHERE order_status = 'đã thanh toán' AND " + condition2
	queryStr4 := "SELECT COALESCE(SUM(bet_total), 0) as total_bet FROM log_information WHERE uid > 0 AND game != '' AND " + condition
	queryStr5 := "SELECT COUNT(*) as bet_num FROM log_information WHERE uid > 0 AND game != '' AND " + condition
	queryStr6 := "SELECT COALESCE(SUM(used), 0) as total_win_lose FROM log_information WHERE uid > 0 AND game != '' AND " + condition

	current_r1, err1 := db.QueryString(queryStr1)
        logger.Println("queryStr1 is: " + queryStr1)
        if err1 != nil {
                logger.Error(err1)
        }

	current_r2, err2 := db.QueryString(queryStr2)
	logger.Println("queryStr2 is: " + queryStr2)
	if err2 != nil {
                logger.Error(err2)
        }

	current_r3, err3 := db.QueryString(queryStr3)
        logger.Println("queryStr3 is: " + queryStr3)
        if err3 != nil {
                logger.Error(err3)
        }

        current_r4, err4 := db.QueryString(queryStr4)
        logger.Println("queryStr4 is: " + queryStr4)
        if err4 != nil {
                logger.Error(err4)
        }

	current_r5, err5 := db.QueryString(queryStr5)
        logger.Println("queryStr5 is: " + queryStr5)
        if err5 != nil {
                logger.Error(err5)
        }

        current_r6, err6 := db.QueryString(queryStr6)
        logger.Println("queryStr6 is: " + queryStr6)
        if err6 != nil {
                logger.Error(err6)
        }
	

	current_reports1 := helper.MapStringToMapInterface(current_r1)
	for _, v := range current_reports1{
		if v["register_num"] == "" {
			v["register_num"] = "0"
		}
	}
	
	current_reports2 := helper.MapStringToMapInterface(current_r2)
	for _, v := range current_reports2{
		if v["login_num"] == "" {
                        v["login_num"] = "0"
                }
	}

	current_reports3 := helper.MapStringToMapInterface(current_r3)
        for _, v := range current_reports3{
		if v["total_payment_amount"] == "" {
                        v["total_payment_amount"] = "0"
                }
        }

	current_reports4 := helper.MapStringToMapInterface(current_r4)
        for _, v := range current_reports4{
		if v["total_bet"] == "" {
                        v["total_bet"] = "0"
                }
        }

	current_reports5 := helper.MapStringToMapInterface(current_r5)
        for _, v := range current_reports5{
		if v["bet_num"] == "" {
                        v["bet_num"] = "0"
                }
        }

        current_reports6 := helper.MapStringToMapInterface(current_r6)
        for _, v := range current_reports6{
        	if v["total_win_lose"] == "" {
                        v["total_win_lose"] = "0"
                }
	}

	if len(current_r1) > 0 && len(current_r2) > 0 && len(current_r3) > 0 && len(current_r4) > 0 && len(current_r5) > 0 && len(current_r6) > 0{
		logger.Println("in return now")
                return map[string]interface{}{
			"current_reports1": current_reports1,
			"current_reports2": current_reports2,
			"current_reports3": current_reports3,
                        "current_reports4": current_reports4,
                        "current_reports5": current_reports5,
                        "current_reports6": current_reports6,
		}
        }
        return nil
}

func GetCurrentActive() []map[string]string {
/*	condition := "CASE WHEN " +
                "(TIME(curtime()) >= '01:00:00' AND TIME(curtime()) <= '23:59:59') " +
                "THEN (create_at >= curdate() - INTERVAL 0 DAY + INTERVAL 1 HOUR " +
                "AND create_at <= curdate() + INTERVAL 1 DAY + INTERVAL 59 MINUTE + INTERVAL 59 SECOND) " +
                "ELSE (create_at >= curdate() - INTERVAL 1 DAY + INTERVAL 1 HOUR " +
                "AND create_at < curdate() + INTERVAL 0 DAY + INTERVAL 59 MINUTE + INTERVAL 59 SECOND) " +
                "END " 
*/
	queryStr := "SELECT CASE WHEN COUNT(distinct li.uid) = '' THEN '0' " +
		"ELSE COUNT(distinct li.uid) END as 'active_num' " +
		"FROM `user` u " +
		"LEFT JOIN `game_coin_locker` ON u.`uid` = `game_coin_locker`.`uid` " +
		"LEFT JOIN `log_information` li ON u.`uid` = li.`uid` AND li.`game`<>'' " +
		"LEFT JOIN `order` o ON o.`uid` = u.`uid` AND o.`order_status` = '已付款'  " +
		"WHERE 1=1 AND " +
		"u.`user_permission` != 'super_admin' AND " +
		"DATE(create_at) = DATE(curdate())  " 

//"ORDER BY u.`create_at` DESC " 

	current_active, err := db.QueryString(queryStr)
        //logger.Printf(queryStr)
        if err != nil {
                logger.Error(err)
        }

        if len(current_active) > 0 {
                return current_active
        }
        return nil
}

func GetCurrentGameOnline() map[string]interface{} {
	selectStr := "SELECT COUNT(DISTINCT uid) as "
	fromStr	  := "FROM gamespace.log_information WHERE uid > 0 AND game = "
	coDateStr := "AND operating_time >= now() - INTERVAL 1 MINUTE "

	queryStr1 := selectStr + "pinshi_online " + fromStr + "'拼十' " + coDateStr
	queryStr2 := selectStr + "roulette_online " + fromStr + "'动物乐园' " + coDateStr
        queryStr3 := selectStr + "haochehui_online " + fromStr + "'豪车汇' " + coDateStr
        queryStr4 := selectStr + "fruit_online " + fromStr + "'水果' " + coDateStr
	queryStr5 := "SELECT now() as time_online "
	c_time, err5 := db.QueryString(queryStr5)
	logger.Printf(queryStr5)
	if err5 != nil {
                logger.Error(err5)
	}
	current_time := helper.MapStringToMapInterface(c_time)
	for _, v := range current_time{
		if v["time_online"] == "" {
                        v["time_online"] = "--00:00:00--"
                }
        }        

	c_pinshi, err1 := db.QueryString(queryStr1)
        //logger.Printf(queryStr1)
        if err1 != nil {
                logger.Error(err1)
        }
	current_pinshi := helper.MapStringToMapInterface(c_pinshi)
	for _, v := range current_pinshi{
                if v["pinshi_online"] == "" {
                        v["pinshi_online"] = "00.pinshi"
       		}
	}

        c_roulette, err2 := db.QueryString(queryStr2)
        //logger.Printf(queryStr2)
        if err2 != nil {
                logger.Error(err2)
        }
        current_roulette := helper.MapStringToMapInterface(c_roulette)
        for _, v := range current_roulette{
                if v["roulette_online"] == "" {
                        v["roulette_online"] = "00.roulette"
                }
        }

        c_haochehui, err3 := db.QueryString(queryStr3)
        //logger.Printf(queryStr3)
        if err3 != nil {
                logger.Error(err3)
        }
        current_haochehui := helper.MapStringToMapInterface(c_haochehui)
        for _, v := range current_haochehui{
                if v["haochehui_online"] == "" {
                        v["haochehui_online"] = "00.haochehui"
                }
        }

	c_fruit, err4 := db.QueryString(queryStr4)
	//logger.Printf(queryStr4)
	if err4 != nil {
		logger.Error(err4)
	}
	current_fruit := helper.MapStringToMapInterface(c_fruit)
	for _, v := range current_fruit{
		if v["fruit_online"] == ""{
			v["fruit"] = "00.fruit"
		}
	}


	if len(c_pinshi) > 0 && len(c_roulette) > 0 && len(c_haochehui) > 0 && len(c_fruit) > 0 {
		return map[string]interface{}{
                        "current_pinshi": current_pinshi,
			"current_roulette": current_roulette,
                        "current_haochehui": current_haochehui,
                        "current_fruit": current_fruit,
			"current_time": current_time,
		}
        }
        return nil
}

const (
	TopUpReportIsFirst    string = "是"
	TopUpReportIsNotFirst string = "否"
)

var TopupReportTypes = []map[string]interface{}{
	//{"val": "", "text": "全部"},
	{"val": TopUpReportIsFirst, "text": "首充"},
	{"val": TopUpReportIsNotFirst, "text": "非首充"},
	{"val": "", "text": "全部"},
}
