package db

import (
	"admin/helper"
	"fmt"
	"strconv"
//	"strings"
)

func GetLockerlogsByPageNumber(pagenumber, urlParams string, searchParams map[string]string) map[string]interface{} {
	offset := (helper.StringToInt(pagenumber) - 1) * helper.RECORDS_PER_PAGE

	conditionQuery := "WHERE 1=1  "
	if searchParams["uid"] != "" {
		conditionQuery += "AND `uid` != '0' AND `uid`= '" + searchParams["uid"] + "' "
	}
	if searchParams["operate"] != "" {
		if searchParams["operate"] == fmt.Sprintf("%v", deposit) {
                        conditionQuery += "AND `operate` = '" + fmt.Sprintf("%v", deposit) +  "' "
                } else if searchParams["operate"] == fmt.Sprintf("%v", withdrawl){
                        conditionQuery += "AND `operate` = '" + fmt.Sprintf("%v", withdrawl) + "' "
                } else{
			conditionQuery += "AND `operate` = '" + fmt.Sprintf("%v", deposit) +  "' or  `operate` = '" + fmt.Sprintf("%v", withdrawl) + "' "
		}

	}

//	if searchParams["operate"] != "" {
//		conditionQuery += "AND `operate`= '" + searchParams["operate"] + "' "
//	}
	if searchParams["amount"] != "" {
		conditionQuery += "AND `amount`= '" + searchParams["amount"] + "' "
	}
	if searchParams["dateStart"] != "" {
		conditionQuery += "AND `date` >= '" + searchParams["dateStart"] + "' "
	}
	if searchParams["dateEnd"] != "" {
		conditionQuery += "AND `date` <= '" + searchParams["dateEnd"] + "' "
	}
	if searchParams["balance"] != "" {
		conditionQuery += "AND (`balance` = '" + searchParams["balance"] + "' "
	}

	COOKIENAME := "LOCKERLOGLISTING" + conditionQuery
	totalRecords := 0

	if cookietotalStr := Ctx.GetCookie(COOKIENAME); cookietotalStr != "" { // total found!
		totalRecords = helper.StringToInt(cookietotalStr)
	} else {
		queryStr := "SELECT count(*) as total from `game_coin_locker_history` " + conditionQuery + " LIMIT 1"
		// logger.Print(queryStr)
		total, err := db.QueryString(queryStr)
		if err != nil {
			logger.Error(err)
		}
		if len(total) > 0 {
			totalRecords = helper.StringToInt(total[0]["total"])
			Ctx.SetCookieKV(COOKIENAME, total[0]["total"])
		}
	}

	queryStr := "select * from game_coin_locker_history " + conditionQuery +
		" Limit " +
		strconv.Itoa(helper.RECORDS_PER_PAGE) +
		" OFFSET " +
		strconv.Itoa(offset) // limit 50 offset

	logger.Print(queryStr)

	lockerlogs, err := db.QueryString(queryStr)
	if err != nil {
		logger.Error(err)
	}
	lockerlogsInterface := helper.MapStringToMapInterface(lockerlogs)
	for _, v := range lockerlogsInterface {
		if v["operate"].(string) == "Tiền gửi" {
			v["operate"] = "存入"
		}
		if v["operate"].(string) == "đưa ra" {
			v["operate"] = "取出"
		}
		v["date"] = helper.DisplayDate(v["date"].(string))
	}

	//logger.Print(lockerlogsInterface)

	queryStrToExport := "select * from game_coin_locker_history "

	if searchParams["dateStart"] != "" || searchParams["dateEnd"] != "" || searchParams["uid"] != "" {
		queryStrToExport += conditionQuery + " ORDER BY date DESC "
	} else {
		queryStrToExport += " ORDER BY date DESC LIMIT 10"
	}

	lockerlogsToExport, err := db.QueryString(queryStrToExport)
	if err != nil {
		logger.Error(err)
	}
	lockerlogsInterfaceToExport := helper.MapStringToMapInterface(lockerlogsToExport)
	for _, v := range lockerlogsInterfaceToExport {
		if v["operate"].(string) == "Tiền gửi" {
                        v["operate"] = "存入"
                }
                if v["operate"].(string) == "đưa ra" {
                        v["operate"] = "取出"
                }
                v["date"] = helper.DisplayDate(v["date"].(string))
	}

	if len(lockerlogsInterface) > 0 && totalRecords > 0 {
		numPages := totalRecords / helper.RECORDS_PER_PAGE

		if totalRecords%helper.RECORDS_PER_PAGE > 0 {
			numPages++
		}

		return map[string]interface{}{
			"Lockerlogs":         lockerlogsInterface,
			"LockerlogsToExport": lockerlogsInterfaceToExport,
			"Total":        totalRecords,
			"NumPages":     numPages,
		}
	}
	return nil
}

const (
        deposit  string = "Tiền gửi"
        withdrawl string = "đưa ra"
)

var LockerOperate = []map[string]interface{}{
        {"val": "", "text": "全部"},
        {"val": deposit, "text": "存入"},
        {"val": withdrawl, "text": "取出"},
}

