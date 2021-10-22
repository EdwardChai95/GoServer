package controllers

import (
	"admin/db"
	"admin/template"
	"fmt"
	"strings"

	"github.com/kataras/iris/v12"
)

var proxyuid string

func (c *controllers) Proxy_get(ctx iris.Context) {
	pageNumber := ctx.URLParamDefault("pageNo", "1")
	uid := ctx.URLParamDefault("uid", "")
	dateStart := ctx.URLParamDefault("dateStart", "")
	dateEnd := ctx.URLParamDefault("dateEnd", "")
	formValues := ctx.FormValues()
	var formDate []string
	var fDate string
	if formValues["today"] != nil {
		formDate = formValues["today"]
	} else if formValues["yesterday"] != nil {
		formDate = formValues["yesterday"]
	} else if formValues["past"] != nil {
		formDate = formValues["past"]
	} else if formValues["all"] != nil {
		formDate = formValues["all"]
	}
	fDate = strings.Join(formDate, "")

	searchParams := map[string]string{
		"uid":       strings.TrimSpace(uid),
		"dateStart": strings.TrimSpace(dateStart),
		"dateEnd":   strings.TrimSpace(dateEnd),
		"formDate":  strings.TrimSpace(fDate),
	}

	urlParams := ""
	for k, v := range searchParams {
		if strings.TrimSpace(v) != "" {
			urlParams += "&" + k + "=" + v
		}
	}

	data := db.GetProxysByPageNo(pageNumber, urlParams, searchParams)
	for k, v := range data {
		ctx.ViewData(k, v)
	}

	ctx.ViewData("uid", &template.Params{Name: "uid", Title: "代理ID", Value: uid})
	ctx.ViewData("dateStart", &template.Params{Name: "dateStart", Title: "开始时间", Value: dateStart})
	ctx.ViewData("dateEnd", &template.Params{Name: "dateEnd", Title: "结束时间", Value: dateEnd})

	ctx.View("proxy/index.html")
}

func (c *controllers) ProxyUser_get(ctx iris.Context) {
	pageNumber := ctx.URLParamDefault("pageNo", "1")
	uid := ctx.URLParamDefault("uid", "")
	userId := ctx.URLParamDefault("userId", "")
	dateStart := ctx.URLParamDefault("dateStart", "")
	dateEnd := ctx.URLParamDefault("dateEnd", "")
	formValues := ctx.FormValues()

	var formDate []string
	var fDate string
	if uid != "" {
		proxyuid = uid
	} else {
		uid = proxyuid
	}
	fmt.Println("proxy:", proxyuid)
	if formValues["today"] != nil {
		formDate = formValues["today"]
	} else if formValues["yesterday"] != nil {
		formDate = formValues["yesterday"]
	} else if formValues["past"] != nil {
		formDate = formValues["past"]
	} else if formValues["all"] != nil {
		formDate = formValues["all"]
	}
	fDate = strings.Join(formDate, "")

	searchParams := map[string]string{
		"uid":       strings.TrimSpace(uid),
		"userId":    strings.TrimSpace(userId),
		"dateStart": strings.TrimSpace(dateStart),
		"dateEnd":   strings.TrimSpace(dateEnd),
		"formDate":  strings.TrimSpace(fDate),
	}

	urlParams := ""
	for k, v := range searchParams {
		if strings.TrimSpace(v) != "" {
			urlParams += "&" + k + "=" + v
		}
	}

	data := db.GetProxyUsersByPageNo(pageNumber, urlParams, searchParams)
	for k, v := range data {
		ctx.ViewData(k, v)
	}
	ctx.ViewData("uid", &template.Params{Name: "uid", Title: "代理ID", Value: uid})
	ctx.ViewData("userId", &template.Params{Name: "userId", Title: "用户ID", Value: userId})
	ctx.ViewData("dateStart", &template.Params{Name: "dateStart", Title: "开始时间", Value: dateStart})
	ctx.ViewData("dateEnd", &template.Params{Name: "dateEnd", Title: "结束时间", Value: dateEnd})

	ctx.View("proxy/user.html")
}
