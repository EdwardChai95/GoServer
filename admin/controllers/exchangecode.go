package controllers

import (
	"admin/db"
	"admin/helper"
	"admin/template"
	"strings"

	"github.com/kataras/iris/v12"
)

func (c *controllers) Exchangecode_get(ctx iris.Context) { // listing /log
	pageNumber := ctx.URLParamDefault("pageNo", "1")
	exchange_code := ctx.URLParamDefault("exchange_code","")
	uid := ctx.URLParamDefault("uid", "")
	dateStart := ctx.URLParamDefault("dateStart", "")
	dateEnd := ctx.URLParamDefault("dateEnd", "")

	searchParams := map[string]string{
		"exchange_code" : strings.TrimSpace(exchange_code),
		"uid" : strings.TrimSpace(uid),
		"dateStart":   strings.TrimSpace(dateStart),
		"dateEnd":     strings.TrimSpace(dateEnd),
	}

	urlParams := ""
	for k, v := range searchParams {
		if strings.TrimSpace(v) != "" {
			urlParams += "&" + k + "=" + v
		}
	}

	data := db.GetExchangecodesByPageNumber(pageNumber, urlParams, searchParams)

	if data == nil { // problem with parameters
		sess.SetFlash(helper.SESSION_TEMPFAILMESSAGE, "查询失败！")
		ctx.Redirect("/exchangecode", iris.StatusTemporaryRedirect)
		return
	}
	for k, v := range data {
		ctx.ViewData(k, v)
	}

	ctx.ViewData("pagination", helper.PaginationHTML("/exchangecode/",
		urlParams, helper.StringToInt(pageNumber), data["NumPages"].(int)))

	ctx.ViewData("exchange_code", &template.Params{Name: "exchange_code", Title: "礼包码号码", Value: exchange_code})
	ctx.ViewData("uid", &template.Params{Name: "uid", Title: "账号ID", Value: uid})
	ctx.ViewData("dateStart", &template.Params{Name: "dateStart", Title: "时间开始", Value: dateStart})
	ctx.ViewData("dateEnd", &template.Params{Name: "dateEnd", Title: "时间结束", Value: dateEnd})

	ctx.View("exchangecode/index.html")

}
