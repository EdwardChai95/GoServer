package controllers

import (
	"admin/db"
	"admin/helper"
	"admin/template"
	"fmt"
	"strings"

	"github.com/kataras/iris/v12"
)

func (c *controllers) Coin_get(ctx iris.Context) {
	pageNumber := ctx.URLParamDefault("pageNo", "1")
	uid := ctx.URLParamDefault("uid", "")
	dateStart := ctx.URLParamDefault("dateStart", "")
	dateEnd := ctx.URLParamDefault("dateEnd", "")

	searchParams := map[string]string{
		"uid":       strings.TrimSpace(uid),
		"dateStart": strings.TrimSpace(dateStart),
		"dateEnd":   strings.TrimSpace(dateEnd),
	}

	urlParams := ""
	for k, v := range searchParams {
		if strings.TrimSpace(v) != "" {
			urlParams += "&" + k + "=" + v
		}
	}
	fmt.Println("ok gamecoinlist")

	data := db.GetGamecoinListByPageNo(pageNumber, urlParams, searchParams)
	if data == nil { // problem with parameters
		ctx.Redirect("/coin", iris.StatusTemporaryRedirect)
		return
	}
	for k, v := range data {
		ctx.ViewData(k, v)
	}
	ctx.ViewData("pagination", helper.PaginationHTML("/coin/", urlParams, helper.StringToInt(pageNumber), data["NumPages"].(int)))
	ctx.ViewData("uid", &template.Params{Name: "uid", Title: "账号ID", Value: uid})
	ctx.ViewData("dateStart", &template.Params{Name: "dateStart", Title: "开始时间", Value: dateStart})
	ctx.ViewData("dateEnd", &template.Params{Name: "dateEnd", Title: "结束时间", Value: dateEnd})
	ctx.View("coin/index.html")
}
