package db

import (
	"admin/helper"
	"strconv"
	"strings"
)

func GetOrdersByPageNo(pagenumber, urlParams string, searchParams map[string]string) map[string]interface{} { // UID，支付状态筛选，订单号
	offset := (helper.StringToInt(pagenumber) - 1) * helper.RECORDS_PER_PAGE
	conditionQuery := "WHERE 1=1  "
	if searchParams["order_id"] != "" {
		conditionQuery += "AND `order_id`= '" + searchParams["order_id"] + "' "
	}
	if searchParams["uid"] != "" {
		conditionQuery += "AND `uid` != '0' AND `uid`= '" + searchParams["uid"] + "' "
	}
	if searchParams["order_status"] != "" {
		conditionQuery += "AND `order_status`= '" + searchParams["order_status"] + "' "
	}

	COOKIENAME := "ORDERLISTING" + urlParams
	totalRecords := 0

	if cookietotalStr := Ctx.GetCookie(COOKIENAME); cookietotalStr != "" { // total found!
		totalRecords = helper.StringToInt(cookietotalStr)
	} else {
		total, err := db.QueryString("SELECT count(*) as total from `order` " + conditionQuery + " LIMIT 1")
		if err != nil {
			logger.Error(err)
		}
		if len(total) > 0 {
			totalRecords = helper.StringToInt(total[0]["total"])
			Ctx.SetCookieKV(COOKIENAME, total[0]["total"])
		}
	}

	orders, err := db.QueryString("SELECT `order`.*, " +
		"CASE `order`.`order_status` " +
		"WHEN 'tật nguyền' then '订单失效' " +
        	"WHEN 'Chế biến' then '待付款' " +
        	"WHEN 'đã thanh toán' then '已付款' " +
        	"END as order_status " +
		" from `order` " +
		conditionQuery +
		" ORDER BY `order_id` DESC " +
		" Limit " +
		strconv.Itoa(helper.RECORDS_PER_PAGE) +
		" OFFSET " +
		strconv.Itoa(offset))
	if err != nil {
		logger.Error(err)
	}

	if len(orders) > 0 && totalRecords > 0 {
		numPages := totalRecords / helper.RECORDS_PER_PAGE

		if totalRecords%helper.RECORDS_PER_PAGE > 0 {
			numPages++
		}

		return map[string]interface{}{
			"Orders":   orders,
			"Total":    totalRecords,
			"NumPages": numPages,
		}
	}

	return nil
}

// 读取所有支付状态
func GetAllOrderStatus(selectedVal string) []map[string]interface{} {
	COOKIENAME := "ORDERSTATUS" // cookie key can be any string

	optionsArr := []string{""}

	// logger.Printf("optionsArr: %v", optionsArr)
	if optionsStr := Ctx.GetCookie(COOKIENAME); optionsStr != "" { // reasons found!
		//  convert reasonsStr to slice
		optionsArr = strings.Split(optionsStr, ",")
	} else {
		results, err := db.QueryString("SELECT distinct `order_status` FROM `order`;")
		if err != nil {
			logger.Error(err)
		}
		if len(results) > 0 {
			for _, result := range results {
				optionsArr = append(optionsArr, result["order_status"])
			}
			Ctx.SetCookieKV(COOKIENAME, strings.Join(optionsArr[:], ",")) // save to cookie so that it is easier to retrieve
		}
	}

	dropdownOptions := []map[string]interface{}{}
	for _, option := range optionsArr {
		dropdownOption := map[string]interface{}{"val": option, "text": option}
		if strings.TrimSpace(option) == "" {
			dropdownOption["text"] = "全部"
		}else if strings.TrimSpace(option) == "Chế biến" {
			dropdownOption["text"] = "待付款"
		}else if strings.TrimSpace(option) == "tật nguyền" {
			dropdownOption["text"] = "订单失效"
		}else if strings.TrimSpace(option) == "đã thanh toán" {
			dropdownOption["text"] = "已付款"
		}
		if option == selectedVal {
			dropdownOption["selected"] = true
		}
		dropdownOptions = append(dropdownOptions, dropdownOption)
	}

	return dropdownOptions
}
