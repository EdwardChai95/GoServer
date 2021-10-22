package db

import (
	"admin/helper"
	"fmt"
	//"fmt"
	"strconv"
	//	"strings"
)

func GetStaysByPageNo(pagenumber, urlParams string, searchParams map[string]string) map[string]interface{} {
	offset := (helper.StringToInt(pagenumber) - 1) * helper.RECORDS_PER_PAGE

	conditionQuery := "WHERE 1=1 "
	if searchParams["dateStart"] != "" {
		conditionQuery += "AND s.log_day >= DATE('" + searchParams["dateStart"] + "') "
	}
	if searchParams["dateEnd"] != "" {
		conditionQuery += "AND s.log_day <= DATE('" + searchParams["dateEnd"] + "') "
	}

	COOKIENAME := "STAYLISTING" + conditionQuery
	totalRecords := 0

	if cookietotalStr := Ctx.GetCookie(COOKIENAME); cookietotalStr != "" { // total found!
		totalRecords = helper.StringToInt(cookietotalStr)
	} else {
		queryStr := "SELECT count(DISTINCT s.log_day) as total from `stay` s " + conditionQuery + " AND s.log_day != DATE(curdate()) LIMIT 1"
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

	queryStr := "SELECT s.num_reg as countReg, s.num_day1 as count_d01, s.num_day3 as count_d03, " +
		"s.num_day7 as count_d07, s1.num_day15 as count_d15, s1.num_day30 as count_d30, " +
		"s1.num_day60 as count_d60, s1.num_day90 as count_d90, s.log_day " +
		"FROM stay s " +
		"left join stay1 s1 " +
		"on s.log_day = s1.log_day " +
		conditionQuery + " AND s.log_day != DATE(curdate()) " +
		"ORDER BY s.log_day DESC " +
		" Limit " +
		strconv.Itoa(helper.RECORDS_PER_PAGE) +
		" OFFSET " +
		strconv.Itoa(offset) // limit 50 offset

	fmt.Println(queryStr)
	stays, err := db.QueryString(queryStr)
	if err != nil {
		logger.Error(err)
	}
	staysInterface := helper.MapStringToMapInterface(stays)
	for _, v := range staysInterface {
		if v["count_d01"] == "" {
			v["count_d01"] = "-"
		}
	}

	queryStrToExport := "SELECT s.num_reg as countReg, s.num_day1 as count_d01, s.num_day3 as count_d03, " +
		"s.num_day7 as count_d07, s1.num_day15 as count_d15, s1.num_day30 as count_d30, " +
		"s1.num_day60 as count_d60, s1.num_day90 as count_d90, s.log_day " +
		"FROM stay s " +
		"left join stay1 s1 " +
		"on s.log_day = s1.log_day " +
		conditionQuery + "AND s.log_day != DATE(curdate()) " +
		"ORDER BY s.log_day DESC"

	logger.Print(queryStrToExport)

	staysToExport, err := db.QueryString(queryStrToExport)
	if err != nil {
		logger.Error(err)
	}
	staysInterfaceToExport := helper.MapStringToMapInterface(staysToExport)
	for _, v := range staysInterfaceToExport {
		if v["count_d01"] == "" {
			v["count_d01"] = "-"
		}
	}

	if len(staysInterface) > 0 && totalRecords > 0 {
		numPages := totalRecords / helper.RECORDS_PER_PAGE

		if totalRecords%helper.RECORDS_PER_PAGE > 0 {
			numPages++
		}

		return map[string]interface{}{
			"Stays":         staysInterface,
			"StaysToExport": staysInterfaceToExport,
			"Total":         totalRecords,
			"NumPages":      numPages,
		}
	}
	return nil
}
