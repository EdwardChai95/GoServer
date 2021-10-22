package db

import (
	"admin/helper"
	"fmt"
	"strconv"
	"strings"
)

func GetLogsByPageNumber(pagenumber, urlParams string, searchParams map[string]string) map[string]interface{} {
	offset := (helper.StringToInt(pagenumber) - 1) * helper.RECORDS_PER_PAGE

	conditionQuery := "WHERE 1=1  "
	if searchParams["uid"] != "" {
		conditionQuery += "AND `uid` != '0' AND `uid`= '" + searchParams["uid"] + "' "
	}
	if searchParams["reason"] != "" {
		conditionQuery += "AND `reason`= '" + searchParams["reason"] + "' "
	}
	if searchParams["game"] != "" {
		conditionQuery += "AND `game`= '" + searchParams["game"] + "' "
	}
	if searchParams["level"] != "" {
		conditionQuery += "AND `level`= '" + searchParams["level"] + "' "
	}
	if searchParams["otherTerm"] != "" {
		conditionQuery += "AND `other_info`LIKE '%" + searchParams["otherTerm"] + "%' "
	}
	if searchParams["dateStart"] != "" {
		conditionQuery += "AND `operating_time` >= '" + searchParams["dateStart"] + "' "
	}
	if searchParams["dateEnd"] != "" {
		conditionQuery += "AND `operating_time` <= '" + searchParams["dateEnd"] + "' "
	}
	if searchParams["params"] != "" {
		conditionQuery += "AND (`params` = '" + searchParams["params"] + "' OR `log_information_id` = '" + searchParams["params"] + "') "
	}
	if searchParams["gameLogType"] != "" {
		if searchParams["gameLogType"] == fmt.Sprintf("%v", GameLogTypeRobot) ||
			searchParams["gameLogType"] == fmt.Sprintf("%v", GameLogTypeSystem) {
			//conditionQuery += fmt.Sprintf("AND (`uid` = '%v')", searchParams["gameLogType"])
			conditionQuery += fmt.Sprintf("AND (`game` != '' AND `uid` = '%v')", searchParams["gameLogType"])

		} else {
			conditionQuery += fmt.Sprintf("AND (`uid` != '%v' AND `uid` != '%v')", GameLogTypeRobot, GameLogTypeSystem)
			//conditionQuery += fmt.Sprintf("AND (`game` != '' AND `uid` != '%v' AND `uid` != '%v')", GameLogTypeRobot, GameLogTypeSystem)
		}
	}

	COOKIENAME := "LOGLISTING" + conditionQuery
	totalRecords := 0

	if cookietotalStr := Ctx.GetCookie(COOKIENAME); cookietotalStr != "" { // total found!
		totalRecords = helper.StringToInt(cookietotalStr)
	} else {
		queryStr := "SELECT count(*) as total from `log_information` " + conditionQuery + " LIMIT 1"
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

	queryStr := "select log_information_id, uid, game, " +
		"other_info, operating_time, l.before, used, l.after, " +
		"CASE " +
		"WHEN REGEXP_INSTR(other_info, '总输赢') != 0 AND uid = 0 " +
		"THEN CASE  " +
		"WHEN game != '水果' " +
		"THEN SUBSTRING(other_info,REGEXP_INSTR(other_info, '总押注')+4,(REGEXP_INSTR(other_info, '输赢') - REGEXP_INSTR(other_info, '总押注'))-7) " +
		"ELSE SUBSTRING(other_info,REGEXP_INSTR(other_info, '总押注')+4,(REGEXP_INSTR(other_info, '输赢') - REGEXP_INSTR(other_info, '总押注'))-6) " +
		"END " +

		"ELSE SUBSTRING(other_info,REGEXP_INSTR(other_info, '总押注')+4,(REGEXP_INSTR(other_info, '输赢') - REGEXP_INSTR(other_info, '总押注'))-5) " +
		"END as total_bet, " +

		"CASE l.`level` " +
		"WHEN 'Trận sơ cấp' THEN '初级场' " +
		"WHEN 'Trận trung cấp' THEN '中级场' " +
		"WHEN 'Trận cao cấp' THEN '高级场' " +
		"WHEN 'Trận đặc biệt' THEN '超级场' " +
		"WHEN 'Khu VIP' THEN 'VIP场' " +
		"WHEN 'Vòng lớn Bàn một' THEN '百人一号场' " +
		"WHEN 'Vòng lớn Bảng hai' THEN '百人二号场' " +
		"WHEN 'Vòng lớn Bảng ba' THEN '百人三号场' " +
		"WHEN 'Vòng lớn Bảng bốn' THEN '百人四号场' " +
		"WHEN 'Tomby tròn Bàn một' THEN '通比一号场' " +
		"WHEN 'Tomby tròn Bảng hai' THEN '通比二号场' " +
		"WHEN 'Vòng lớn Bảng năm' THEN '百人五号场' " +
		"WHEN 'Vòng lớn Bảng sáu' THEN '百人六号场' " +
		"WHEN 'Vòng lớn Bảng bảy' THEN '百人七号场' " +
		"END as level, " +
		"CASE l.`reason` " +
		"WHEN 'Phát phần thưởng' THEN '赠送' " +
		"WHEN 'Quyên góp' THEN '捐赠' " +
		"WHEN '注册奖励' THEN '注册奖励' " +
		"WHEN '登陆奖励' THEN '登陆奖励' " +
		"WHEN '登陆' THEN '登陆' " +
		"WHEN '低保' THEN '低保' " +
		"WHEN '新手卡' THEN '新手卡' " +
		"WHEN '使用兑换码' THEN '使用兑换码' " +
		"WHEN '充值' THEN '充值' " +
		"WHEN '保险箱帐变' THEN '保险箱帐变' " +
		"WHEN '管理员加' THEN '管理员加' " +
		"WHEN '管理员减' THEN '管理员减' " +
		"END as reason " +
		"from `log_information` l " +
		conditionQuery + " ORDER BY operating_time DESC " +
		" Limit " +
		strconv.Itoa(helper.RECORDS_PER_PAGE) +
		" OFFSET " +
		strconv.Itoa(offset) // limit 50 offset

	//logger.Print(queryStr)

	logs, err := db.QueryString(queryStr)
	if err != nil {
		logger.Error(err)
	}
	logsInterface := helper.MapStringToMapInterface(logs)
	for _, v := range logsInterface {
		if v["game"] == "" {
			v["game"] = "-"
		}
		if v["level"] == "" {
			v["level"] = "-"
		}
		// if v["other_info"] != "" && v["reason"] == "保险箱帐变" {
		// 	out := helper.JsonStrToMap(v["other_info"].(string))

		// 	v["after"] = "操作前：" + helper.Int64ToString(helper.JsonObjToInt(out, "after"))
		// 	v["before"] = "操作后：" + helper.Int64ToString(helper.JsonObjToInt(out, "before"))
		// 	v["used"] = "游戏币：" + (helper.JsonObjToStr(out, "deposit"))

		// 	v["other_info"] = ""
		// }
		if v["game"] != "" && v["reason"] == "" && v["uid"] == "0" {
			v["reason"] = "-" // 系统
			v = helper.CustomServerLog(v)
		} else if v["game"] != "" && v["reason"] == "" {
			v["reason"] = "押注" // 玩家
			v = helper.CustomPlayerLog(v)
		}
		if v["uid"].(string) == "0" {
			v["uid"] = "系统"
		}
		if v["uid"].(string) == "-1" {
			v["uid"] = "机器人"
		}
		v["operating_time"] = helper.DisplayDate(v["operating_time"].(string))
		if v["reason"].(string) == "保险箱帐变" {
			v["game"] = "保险箱"
			v["other_info"] = ""
			if string(v["used"].(string)[0]) != "-" {
				v["reason"] = "取出"
			} else if string(v["used"].(string)[0]) == "-" {
				v["reason"] = "存入"
			} else if v["after"] == v["before"] {
				v["reason"] = "保险箱查看"
			}
		}
	}

	//logger.Print(logsInterface)

	queryStrToExport := "select log_information_id, uid, game, " +
		"other_info, operating_time, l.before, used, l.after, " +
		"CASE " +
		"WHEN REGEXP_INSTR(other_info, '总输赢') != 0 AND uid = 0 " +
		"THEN CASE  " +
		"WHEN game != '水果' " +
		"THEN SUBSTRING(other_info,REGEXP_INSTR(other_info, '总押注')+4,(REGEXP_INSTR(other_info, '输赢') - REGEXP_INSTR(other_info, '总押注'))-7) " +
		"ELSE SUBSTRING(other_info,REGEXP_INSTR(other_info, '总押注')+4,(REGEXP_INSTR(other_info, '输赢') - REGEXP_INSTR(other_info, '总押注'))-6) " +
		"END " +

		"ELSE SUBSTRING(other_info,REGEXP_INSTR(other_info, '总押注')+4,(REGEXP_INSTR(other_info, '输赢') - REGEXP_INSTR(other_info, '总押注'))-5) " +
		"END as total_bet, " +

		"CASE l.`level` " +
		"WHEN 'Trận sơ cấp' THEN '初级场' " +
		"WHEN 'Trận trung cấp' THEN '中级场' " +
		"WHEN 'Trận cao cấp' THEN '高级场' " +
		"WHEN 'Trận đặc biệt' THEN '超级场' " +
		"WHEN 'Khu VIP' THEN 'VIP场' " +
		"WHEN 'Vòng lớn Bàn một' THEN '百人一号场' " +
		"WHEN 'Vòng lớn Bảng hai' THEN '百人二号场' " +
		"WHEN 'Vòng lớn Bảng ba' THEN '百人三号场' " +
		"WHEN 'Vòng lớn Bảng bốn' THEN '百人四号场' " +
		"WHEN 'Tomby tròn Bàn một' THEN '通比一号场' " +
		"WHEN 'Tomby tròn Bảng hai' THEN '通比二号场' " +
		"WHEN 'Vòng lớn Bảng năm' THEN '百人五号场' " +
		"WHEN 'Vòng lớn Bảng sáu' THEN '百人六号场' " +
		"WHEN 'Vòng lớn Bảng bảy' THEN '百人七号场' " +
		"END as level, " +
		"CASE l.`reason` " +
		"WHEN 'Phát phần thưởng' THEN '赠送' " +
		"WHEN 'Quyên góp' THEN '捐赠' " +
		"WHEN '注册奖励' THEN '注册奖励' " +
		"WHEN '登陆奖励' THEN '登陆奖励' " +
		"WHEN '登陆' THEN '登陆' " +
		"WHEN '低保' THEN '低保' " +
		"WHEN '新手卡' THEN '新手卡' " +
		"WHEN '使用兑换码' THEN '使用兑换码' " +
		"WHEN '充值' THEN '充值' " +
		"WHEN '保险箱帐变' THEN '保险箱帐变' " +
		"WHEN '管理员加' THEN '管理员加' " +
		"WHEN '管理员减' THEN '管理员减' " +
		"END as reason " +
		"from `log_information` l "

	if searchParams["dateStart"] != "" || searchParams["dateEnd"] != "" || searchParams["uid"] != "" {
		queryStrToExport += conditionQuery + " ORDER BY operating_time DESC "
	} else {
		queryStrToExport += " ORDER BY operating_time DESC LIMIT 10"
	}

	logsToExport, err := db.QueryString(queryStrToExport)
	if err != nil {
		logger.Error(err)
	}
	logsInterfaceToExport := helper.MapStringToMapInterface(logsToExport)
	for _, v := range logsInterfaceToExport {
		if v["game"] == "" {
			v["game"] = "-"
		}
		if v["level"] == "" {
			v["level"] = "-"
		}
		if v["game"] != "" && v["reason"] == "" && v["uid"] == "0" {
			v["reason"] = "-" // 系统
			v = helper.CustomServerLog(v)
		} else if v["game"] != "" && v["reason"] == "" {
			v["reason"] = "押注" // 玩家
			v = helper.CustomPlayerLog(v)
		}
		if v["uid"].(string) == "0" {
			v["uid"] = "系统"
		}
		v["operating_time"] = helper.DisplayDate(v["operating_time"].(string))
		if v["reason"].(string) == "保险箱帐变" {
			v["game"] = "保险箱"
			v["other_info"] = ""
			if string(v["used"].(string)[0]) != "-" {
				v["reason"] = "取出"
			} else if string(v["used"].(string)[0]) == "-" {
				v["reason"] = "存入"
			} else if v["after"] == v["before"] {
				v["reason"] = "保险箱查看"
			}
		}
	}

	if len(logsInterface) > 0 && totalRecords > 0 {
		numPages := totalRecords / helper.RECORDS_PER_PAGE

		if totalRecords%helper.RECORDS_PER_PAGE > 0 {
			numPages++
		}

		return map[string]interface{}{
			"Logs":         logsInterface,
			"LogsToExport": logsInterfaceToExport,
			"Total":        totalRecords,
			"NumPages":     numPages,
		}
	}
	return nil
}

func GetLogInformationOptionList(col string) []string {
	COOKIENAME := "allOptions_" + col

	optionsArr := []string{}

	if optionsStr := Ctx.GetCookie(COOKIENAME); optionsStr != "" { // reasons found!
		//  convert reasonsStr to slice
		optionsArr = strings.Split(optionsStr, ",")
	} else {
		results, err := db.QueryString("SELECT distinct `" + col + "` FROM `log_information`;")
		if err != nil {
			logger.Error(err)
		}
		if len(results) > 0 {
			for _, result := range results {
				optionsArr = append(optionsArr, result[col])
			}
			Ctx.SetCookieKV(COOKIENAME, strings.Join(optionsArr[:], ","))
		}
	}

	return optionsArr
}

// 读取所有操作原因
func GetAllOptions(col, selectedVal string) []map[string]interface{} {
	optionsArr := GetLogInformationOptionList(col)

	dropdownOptions := []map[string]interface{}{}
	for _, option := range optionsArr {
		dropdownOption := map[string]interface{}{"val": option, "text": option}
		if strings.TrimSpace(option) == "" {
			dropdownOption["text"] = "全部"
		} else if strings.TrimSpace(option) == "Trận sơ cấp" {
			dropdownOption["text"] = "初级场"
		} else if strings.TrimSpace(option) == "Trận trung cấp" {
			dropdownOption["text"] = "中级场"
		} else if strings.TrimSpace(option) == "Trận cao cấp" {
			dropdownOption["text"] = "高级场"
		} else if strings.TrimSpace(option) == "Trận đặc biệt" {
			dropdownOption["text"] = "超级场"
		} else if strings.TrimSpace(option) == "Khu VIP" {
			dropdownOption["text"] = "VIP场"
		} else if strings.TrimSpace(option) == "Vòng lớn Bàn một" {
			dropdownOption["text"] = "百人一号场"
		} else if strings.TrimSpace(option) == "Vòng lớn Bảng hai" {
			dropdownOption["text"] = "百人二号场"
		} else if strings.TrimSpace(option) == "Vòng lớn Bảng ba" {
			dropdownOption["text"] = "百人三号场"
		} else if strings.TrimSpace(option) == "Vòng lớn Bảng bốn" {
			dropdownOption["text"] = "百人四号场"
		} else if strings.TrimSpace(option) == "Tomby tròn Bàn một" {
			dropdownOption["text"] = "通比一号场"
		} else if strings.TrimSpace(option) == "Tomby tròn Bảng hai" {
			dropdownOption["text"] = "通比二号场"
		} else if strings.TrimSpace(option) == "Vòng lớn Bảng năm" {
			dropdownOption["text"] = "百人五号场"
		} else if strings.TrimSpace(option) == "Vòng lớn Bảng sáu" {
			dropdownOption["text"] = "百人六号场"
		} else if strings.TrimSpace(option) == "Vòng lớn Bảng bảy" {
			dropdownOption["text"] = "百人七号场"
		} else if strings.TrimSpace(option) == "Phát phần thưởng" {
			dropdownOption["text"] = "赠送"
		} else if strings.TrimSpace(option) == "Quyên góp" {
			dropdownOption["text"] = "捐赠"
		}
		if option == selectedVal {
			dropdownOption["selected"] = true
		}
		dropdownOptions = append(dropdownOptions, dropdownOption)
	}

	return dropdownOptions
}

const (
	GameLogTypePlayer int = 1
	GameLogTypeRobot  int = -1
	GameLogTypeSystem int = 0
)

var GameLogTypes = []map[string]interface{}{
	{"val": GameLogTypePlayer, "text": "玩家"},
	{"val": GameLogTypeRobot, "text": "机器人"},
	{"val": GameLogTypeSystem, "text": "系统"},
	{"val": "", "text": "全部"},
}
