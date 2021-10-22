package controllers

import (
	"admin/db"
	"admin/helper"
	"admin/template"
	"strings"

	"github.com/kataras/iris/v12"
)

func (c *controllers) Orders_get(ctx iris.Context) { // GET /orders
	pageNumber := ctx.URLParamDefault("pageNo", "1")
	order_id := ctx.URLParamDefault("order_id", "")
	uid := ctx.URLParamDefault("uid", "")
	order_status := ctx.URLParamDefault("order_status", "")

	searchParams := map[string]string{
		"order_id":     strings.TrimSpace(order_id),
		"uid":          strings.TrimSpace(uid),
		"order_status": strings.TrimSpace(order_status),
	}

	urlParams := ""
	for k, v := range searchParams {
		if strings.TrimSpace(v) != "" {
			urlParams += "&" + k + "=" + v
		}
	}

	data := db.GetOrdersByPageNo(pageNumber, urlParams, searchParams)
	if data == nil { // problem with parameters
		ctx.Redirect("/orders", iris.StatusTemporaryRedirect)
		return
	}

	for k, v := range data {
		ctx.ViewData(k, v)
	}
	ctx.ViewData("pagination", helper.PaginationHTML("/orders/", urlParams,
		helper.StringToInt(pageNumber), data["NumPages"].(int)))

	ctx.ViewData("order_id", &template.Params{Name: "order_id", Title: "订单号", Value: order_id})
	ctx.ViewData("uid", &template.Params{Name: "uid", Title: "UID", Value: uid})
	ctx.ViewData("order_status", &template.Params{Name: "order_status", Title: "支付状态",
		Options: db.GetAllOrderStatus(order_status)})

	ctx.View("orders/index.html")

}
