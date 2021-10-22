package controllers

import (
	"admin/db"
	"admin/helper"
	"admin/template"
	"fmt"
	"strings"
	"log"
	"github.com/kataras/iris/v12"
)

func (c *controllers) LockerLog_get(ctx iris.Context) { // listing /lockerLog
	pageNumber := ctx.URLParamDefault("pageNo", "1")
	uid := ctx.URLParamDefault("uid", "")
	operate := ctx.URLParamDefault("operate", "")
//	amount := ctx.URLParamDefault("amount", "")
//	date := ctx.URLParamDefault("date", "")
	dateStart := ctx.URLParamDefault("dateStart", "")
	dateEnd := ctx.URLParamDefault("dateEnd", "")
//	balance := ctx.URLParamDefault("balance", "")

	searchParams := map[string]string{
		"uid":         strings.TrimSpace(uid),
		"operate":      strings.TrimSpace(operate),
//		"amount":        strings.TrimSpace(amount),
		//"date":   strings.TrimSpace(date),
		"dateStart":   strings.TrimSpace(dateStart),
		"dateEnd":     strings.TrimSpace(dateEnd),
//		"balance":      strings.TrimSpace(balance),
	}

	urlParams := ""
	for k, v := range searchParams {
		if strings.TrimSpace(v) != "" {
			urlParams += "&" + k + "=" + v
		}
	}
	log.Println("here is connnnnnnnnttttttttttttttttttttttttttttttt")
	data := db.GetLockerlogsByPageNumber(pageNumber, urlParams, searchParams)

	if data == nil { // problem with parameters
		sess.SetFlash(helper.SESSION_TEMPFAILMESSAGE, "查询失败！")
		ctx.Redirect("/lockerLog", iris.StatusTemporaryRedirect)
		return
	}
	for k, v := range data {
		ctx.ViewData(k, v)
	}

	ctx.ViewData("pagination", helper.PaginationHTML("/lockerlog/",
		urlParams, helper.StringToInt(pageNumber), data["NumPages"].(int)))

	ctx.ViewData("uid", &template.Params{Name: "uid", Title: "账号ID", Value: uid})

	// operate
        for _, item := range db.LockerOperate {
                if fmt.Sprintf("%v", item["val"]) == operate {
                        item["selected"] = true
                } 
        }
	ctx.ViewData("operate", &template.Params{Name: "operate", Title: "操作原因", Options: db.LockerOperate})

//	ctx.ViewData("operate", &template.Params{Name: "operate", Title: "操作原因", Value: operate})
//	ctx.ViewData("amount", &template.Params{Name: "amount", Title: "金额", Value: amount})
//	ctx.ViewData("date", &template.Params{Name: "date", Title: "时间", Value: date})
	ctx.ViewData("dateStart", &template.Params{Name: "dateStart", Title: "时间开始", Value: dateStart})
	ctx.ViewData("dateEnd", &template.Params{Name: "dateEnd", Title: "时间结束", Value: dateEnd})
//	ctx.ViewData("balance", &template.Params{Name: "balance", Title: "保险箱余额", Value: balance})

	ctx.View("lockerlog/index.html")
//	ctx.View("log/index.html")

}
