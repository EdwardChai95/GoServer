package db

import (
	"admin/helper"
	"strconv"
)

func GetExchangecodesByPageNumber(pagenumber, urlParams string, searchParams map[string]string) map[string]interface{} {
	offset := (helper.StringToInt(pagenumber) - 1) * helper.RECORDS_PER_PAGE

	conditionQuery := "WHERE 1=1 "
	if searchParams["exchange_code"] != "" {
		conditionQuery += "AND exchange_code = '" + searchParams["exchange_code"] + "' "
	}
	if searchParams["uid"] != "" {
		conditionQuery += "AND l.ui` != '' AND l.uid = '" + searchParams["uid"] + "' "
	}
	if searchParams["dateStart"] != "" {
		conditionQuery += "AND l.operating_time >= '" + searchParams["dateStart"] + "' "
	}
	if searchParams["dateEnd"] != "" {
		conditionQuery += "AND l.operating_time <= '" + searchParams["dateEnd"] + "' "
	}

	COOKIENAME := "EXCHANGECODELISTING" + conditionQuery
	totalRecords := 0

	if cookietotalStr := Ctx.GetCookie(COOKIENAME); cookietotalStr != "" { // total found!
		totalRecords = helper.StringToInt(cookietotalStr)
	} else {
		//queryStr := "SELECT count(*) as total from `exchange_code` " + conditionQuery + " LIMIT 1"
		queryStr := "SELECT COUNT(e.exchange_code) AS total FROM exchange_code e " +
			"JOIN log_information l " +
			"ON l.other_info = e.exchange_code " + conditionQuery + " LIMIT 1"

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
	queryStr := "SELECT * " +
		"FROM code_information " +
		conditionQuery +
		" ORDER BY operating_time DESC " +
		"Limit " +
		strconv.Itoa(helper.RECORDS_PER_PAGE) +
		" OFFSET " +
		strconv.Itoa(offset) // limit 50 offset

	logger.Print(queryStr)

	exchangecodes, err := db.QueryString(queryStr)
	if err != nil {
		logger.Error(err)
	}

	exchangecodesInterface := helper.MapStringToMapInterface(exchangecodes)
	for _, v := range exchangecodesInterface {
		if v["uid"].(string) == "" {
			v["uid"] = "-"
		}
		if v["operating_time"].(string) == "" {
			v["operating_time"] = "-"
		}
	}

	queryStrToExport := "SELECT * " +
		"FROM code_information "

	//logger.Print(queryStrToExport)
	if searchParams["dateStart"] != "" || searchParams["dateEnd"] != "" {
		queryStrToExport += conditionQuery + " ORDER BY operating_time DESC"
	} else {
		queryStrToExport += " ORDER BY operating_time DESC LIMIT 10"
	}
	logger.Print(queryStrToExport)

	exchangecodesToExport, err := db.QueryString(queryStrToExport)
	if err != nil {
		logger.Error(err)
	}
	exchangecodesToExportInterface := helper.MapStringToMapInterface(exchangecodesToExport)
	for _, v := range exchangecodesToExportInterface {
		if v["uid"].(string) == "" {
			v["uid"] = "-"
		}
		if v["operating_time"].(string) == "" {
			v["operating_time"] = "-"
			// v["operating_time"] = helper.DisplayDate(v["operating_time"].(string))
		}
	}

	if len(exchangecodesInterface) > 0 && totalRecords > 0 {
		numPages := totalRecords / helper.RECORDS_PER_PAGE

		if totalRecords%helper.RECORDS_PER_PAGE > 0 {
			numPages++
		}

		return map[string]interface{}{
			"Exchangecodes":         exchangecodesInterface,
			"ExchangecodesToExport": exchangecodesToExportInterface,
			"Total":                 totalRecords,
			"NumPages":              numPages,
		}
	}
	return nil
}
