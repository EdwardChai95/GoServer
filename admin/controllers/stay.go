package controllers

import (
	"admin/db"
	"admin/helper"
	"admin/template"

	//"fmt"
	"strings"

	"github.com/kataras/iris/v12"
)

func (c *controllers) Stay_get(ctx iris.Context) { // listing /stay
	pageNumber := ctx.URLParamDefault("pageNo", "1")
	dateStart := ctx.URLParamDefault("dateStart", "")
	dateEnd := ctx.URLParamDefault("dateEnd", "")

	searchParams := map[string]string{
		"dateStart": strings.TrimSpace(dateStart),
		"dateEnd":   strings.TrimSpace(dateEnd),
	}

	urlParams := ""
	for k, v := range searchParams {
		if strings.TrimSpace(v) != "" {
			urlParams += "&" + k + "=" + v
		}
	}

	data := db.GetStaysByPageNo(pageNumber, urlParams, searchParams)
	if data == nil { // problem with parameters
		ctx.Redirect("/stay", iris.StatusTemporaryRedirect)
		return
	}
	for k, v := range data {
		ctx.ViewData(k, v)
	}
	//	ctx.ViewData("searchTerm", uid)
	ctx.ViewData("pagination", helper.PaginationHTML("/stay/", urlParams, helper.StringToInt(pageNumber), data["NumPages"].(int)))

	ctx.ViewData("dateStart", &template.Params{Name: "dateStart", Title: "时间开始", Value: dateStart})
	ctx.ViewData("dateEnd", &template.Params{Name: "dateEnd", Title: "时间结束", Value: dateEnd})

	ctx.View("stay/index.html")
}
