package db

import (
	"admin/helper"
	"fmt"
	"strconv"
)

func GetGamecoinListByPageNo(pagenumber string, urlParams string, searchParams map[string]string) map[string]interface{} {
	offset := (helper.StringToInt(pagenumber) - 1) * helper.RECORDS_PER_PAGE
	conditionQuery := "WHERE 1=1 "

	if searchParams["uid"] != "" {
		conditionQuery += "AND `uid` = '" + searchParams["uid"] + "' "
	}
	if searchParams["dateStart"] != "" && searchParams["dateEnd"] != "" {
		conditionQuery += "AND `operating_time` >= '" + searchParams["dateStart"] + "' "
		conditionQuery += "AND `operating_time` <= '" + searchParams["dateEnd"] + "' "
	}

	COOKIENAME := "GAMECOINLISTING" + urlParams
	totalRecords := 0

	if cookietotalStr := Ctx.GetCookie(COOKIENAME); cookietotalStr != "" { // total found!
		totalRecords = helper.StringToInt(cookietotalStr)
	} else {
		total, err := db.QueryString("SELECT count(distinct l.uid) as total from `log_information` l " + conditionQuery + " AND other_info LIKE '%批量添加%' LIMIT 1")
		fmt.Println("total:", total)
		if err != nil {
			logger.Error(err)
		}
		if len(total) > 0 {
			totalRecords = helper.StringToInt(total[0]["total"])
			Ctx.SetCookieKV(COOKIENAME, total[0]["total"])
		}
	}

	queryStr := "select l.uid, l.used, l.before, l.after, " +
		"l.reason, l.other_info, l.operating_time " +
		"from `log_information` l " +
		conditionQuery +
		"AND l.other_info LIKE '%批量添加%' " +
		"ORDER BY l.operating_time DESC " +
		"LIMIT " +
		strconv.Itoa(helper.RECORDS_PER_PAGE) +
		" OFFSET " +
		strconv.Itoa(offset)

	gamecoins, err := db.QueryString(queryStr)
	if err != nil {
		logger.Error(err)
	}

	if len(gamecoins) >= 0 && totalRecords >= 0 {
		numPages := totalRecords / helper.RECORDS_PER_PAGE

		if totalRecords%helper.RECORDS_PER_PAGE > 0 {
			numPages++
		}

		return map[string]interface{}{
			"Gamecoins": gamecoins,
			"Total":     totalRecords,
			"NumPages":  numPages,
		}
	}
	return nil
}
