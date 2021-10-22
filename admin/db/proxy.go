package db

import (
	"admin/helper"
	"strconv"
)

func GetProxysByPageNo(pagenumber string, urlParams string, searchParams map[string]string) map[string]interface{} {
	offset := (helper.StringToInt(pagenumber) - 1) * helper.RECORDS_PER_PAGE
	conditionQuery := " WHERE 1=1 "

	if searchParams["uid"] != "" {
		conditionQuery += "AND `uid` = '" + searchParams["uid"] + "' "
	}
	if searchParams["dateStart"] != "" && searchParams["dateEnd"] != "" {
		conditionQuery += "AND `operating_time` >= '" + searchParams["dateStart"] + "' "
		conditionQuery += "AND `operating_time` <= '" + searchParams["dateEnd"] + "' "
	} else {
		if searchParams["formDate"] != "" {
			if searchParams["formDate"] == "current_date()" {
				conditionQuery += "AND DATE(operating_time) <= " + searchParams["formDate"] + " "
			} else if searchParams["formDate"] == "curdate() - 1" {
				conditionQuery += "AND DATE(operating_time) >= " + searchParams["formDate"] + " "
				conditionQuery += "AND DATE(operating_time) < curdate() "
			} else {
				conditionQuery += "AND DATE(operating_time) >= " + searchParams["formDate"] + " "
			}
		} else {
			conditionQuery += "AND DATE(operating_time) >= curdate() "
		}
	}

	COOKIENAME := "PROXYLISTING" + urlParams
	totalRecords := 0

	if cookietotalStr := Ctx.GetCookie(COOKIENAME); cookietotalStr != "" { // total found!
		totalRecords = helper.StringToInt(cookietotalStr)
	} else {
		total, err := db.QueryString("SELECT count(distinct p.uid) as total from `proxy` p " + conditionQuery + " LIMIT 1")
		if err != nil {
			logger.Error(err)
		}
		if len(total) > 0 {
			totalRecords = helper.StringToInt(total[0]["total"])
			Ctx.SetCookieKV(COOKIENAME, total[0]["total"])
		}
	}

	select_str := "select uid, promo_code, " +
		"coalesce(sum(promo_num), 0) as promo_num, " +
		"coalesce(sum(active_num), 0) as active_num, " +
		"coalesce(sum(send_num), 0) as send_num, " +
		"coalesce(sum(receive_num), 0) as receive_num, " +
		"coalesce(sum(total_num), 0) as total_num, " +
		"coalesce(sum(total_amount), 0) as total_amount, " +
		"coalesce(sum(service_tax), 0) as service_tax, " +
		"coalesce(sum(count_completed), 0) as count_completed, " +
		"(select operating_time from `proxy` " + conditionQuery + " LIMIT 1) as operating_time " +
		"from `proxy` " +
		conditionQuery +
		" GROUP BY uid, promo_code " +
		"ORDER by operating_time desc " +
		"Limit " +
		strconv.Itoa(helper.RECORDS_PER_PAGE) +
		" OFFSET " +
		strconv.Itoa(offset)

	logger.Println(select_str)
	proxys, err := db.QueryString(select_str)
	if err != nil {
		logger.Error(err)
	}

	if len(proxys) > 0 && totalRecords > 0 {
		numPages := totalRecords / helper.RECORDS_PER_PAGE

		if totalRecords%helper.RECORDS_PER_PAGE > 0 {
			numPages++
		}

		return map[string]interface{}{
			"Proxys":   proxys,
			"Total":    totalRecords,
			"NumPages": numPages,
		}
	}

	return nil
}

func GetProxyUsersByPageNo(pagenumber string, urlParams string, searchParams map[string]string) map[string]interface{} {
	offset := (helper.StringToInt(pagenumber) - 1) * helper.RECORDS_PER_PAGE
	conditionQuery := " WHERE 1=1 "

	if searchParams["uid"] != "" {
		conditionQuery += "AND `proxy_uid` = '" + searchParams["uid"] + "' "
	}
	if searchParams["userId"] != "" {
		conditionQuery += "AND `uid` = '" + searchParams["userId"] + "' "
	}
	if searchParams["dateStart"] != "" && searchParams["dateEnd"] != "" {
		conditionQuery += "AND `operating_time` >= '" + searchParams["dateStart"] + "' "
		conditionQuery += "AND `operating_time` <= '" + searchParams["dateEnd"] + "' "
	} else {
		if searchParams["formDate"] != "" {
			if searchParams["formDate"] == "current_date()" {
				conditionQuery += "AND DATE(operating_time) <= " + searchParams["formDate"] + " "
			} else if searchParams["formDate"] == "curdate() - 1" {
				conditionQuery += "AND DATE(operating_time) >= " + searchParams["formDate"] + " "
				conditionQuery += "AND DATE(operating_time) < curdate() "
			} else {
				conditionQuery += "AND DATE(operating_time) >= " + searchParams["formDate"] + " "
			}
		} else {
			conditionQuery += "AND DATE(operating_time) >= curdate() "
		}
	}

	COOKIENAME := "PROXYUSERLISTING" + urlParams
	totalRecords := 0

	if cookietotalStr := Ctx.GetCookie(COOKIENAME); cookietotalStr != "" { // total found!
		totalRecords = helper.StringToInt(cookietotalStr)
	} else {
		total, err := db.QueryString("SELECT count(distinct p.uid) as total from `proxy_user` p " + conditionQuery + " LIMIT 1")
		if err != nil {
			logger.Error(err)
		}
		if len(total) > 0 {
			totalRecords = helper.StringToInt(total[0]["total"])
			Ctx.SetCookieKV(COOKIENAME, total[0]["total"])
		}
	}

	select_str := "select uid, " +
		"coalesce(sum(total_lose), 0) as total_lose, " +
		"coalesce(sum(total_win), 0) as total_win, " +
		"coalesce(sum(total_win_lose), 0) as total_win_lose," +
		"coalesce(sum(total_broad), 0) as total_broad, " +
		"coalesce(sum(send_num), 0) as send_num, " +
		"coalesce(sum(receive_num), 0) as receive_num, " +
		"coalesce(sum(total_amount), 0) as total_amount, " +
		"coalesce(sum(count_completed), 0) as count_completed, " +
		"(select operating_time from `proxy_user` " + conditionQuery + " LIMIT 1) as operating_time " +
		"from `proxy_user` " +
		conditionQuery +
		" GROUP BY uid " +
		" ORDER BY `operating_time` DESC " +
		" Limit " +
		strconv.Itoa(helper.RECORDS_PER_PAGE) +
		" OFFSET " +
		strconv.Itoa(offset)

	logger.Println(select_str)
	proxyusers, err := db.QueryString(select_str)
	if err != nil {
		logger.Error(err)
	}

	if len(proxyusers) > 0 && totalRecords > 0 {
		numPages := totalRecords / helper.RECORDS_PER_PAGE

		if totalRecords%helper.RECORDS_PER_PAGE > 0 {
			numPages++
		}

		return map[string]interface{}{
			"ProxyUsers": proxyusers,
			"Total":      totalRecords,
			"NumPages":   numPages,
		}
	}

	return nil
}
