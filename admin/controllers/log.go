package controllers

import (
	"admin/db"
	"admin/helper"
	"admin/template"
	"fmt"
	"strings"

	"github.com/kataras/iris/v12"
)

func (c *controllers) Log_get(ctx iris.Context) { // listing /log
	pageNumber := ctx.URLParamDefault("pageNo", "1")
	uid := ctx.URLParamDefault("uid", "")
	reason := ctx.URLParamDefault("reason", "")
	game := ctx.URLParamDefault("game", "")
	level := ctx.URLParamDefault("level", "")
	otherTerm := ctx.URLParamDefault("otherTerm", "")
	dateStart := ctx.URLParamDefault("dateStart", "")
	dateEnd := ctx.URLParamDefault("dateEnd", "")
	params := ctx.URLParamDefault("params", "")
	gameLogType := ctx.URLParamDefault("gameLogType", "")

	searchParams := map[string]string{
		"uid":         strings.TrimSpace(uid),
		"reason":      strings.TrimSpace(reason),
		"game":        strings.TrimSpace(game),
		"level":       strings.TrimSpace(level),
		"otherTerm":   strings.TrimSpace(otherTerm),
		"dateStart":   strings.TrimSpace(dateStart),
		"dateEnd":     strings.TrimSpace(dateEnd),
		"params":      strings.TrimSpace(params), // 参数
		"gameLogType": strings.TrimSpace(gameLogType),
	}

	urlParams := ""
	for k, v := range searchParams {
		if strings.TrimSpace(v) != "" {
			urlParams += "&" + k + "=" + v
		}
	}

	data := db.GetLogsByPageNumber(pageNumber, urlParams, searchParams)

	if data == nil { // problem with parameters
		sess.SetFlash(helper.SESSION_TEMPFAILMESSAGE, "查询失败！")
		ctx.Redirect("/log", iris.StatusTemporaryRedirect)
		return
	}
	for k, v := range data {
		ctx.ViewData(k, v)
	}

	ctx.ViewData("pagination", helper.PaginationHTML("/log/",
		urlParams, helper.StringToInt(pageNumber), data["NumPages"].(int)))

	ctx.ViewData("uid", &template.Params{Name: "uid", Title: "账号ID", Value: uid})
	ctx.ViewData("reason", &template.Params{Name: "reason", Title: "操作原因", Options: db.GetAllOptions("reason", reason)})
	ctx.ViewData("game", &template.Params{Name: "game", Title: "游戏", Options: db.GetAllOptions("game", game)})
	ctx.ViewData("level", &template.Params{Name: "level", Title: "游戏场次", Options: db.GetAllOptions("level", level)})
	ctx.ViewData("otherTerm", &template.Params{Name: "otherTerm", Title: "其他参数", Value: otherTerm})
	ctx.ViewData("dateStart", &template.Params{Name: "dateStart", Title: "时间开始", Value: dateStart})
	ctx.ViewData("dateEnd", &template.Params{Name: "dateEnd", Title: "时间结束", Value: dateEnd})
	ctx.ViewData("params", &template.Params{Name: "params", Title: "参数", Value: params})

	// gameLogType
	for _, item := range db.GameLogTypes {
		if fmt.Sprintf("%v", item["val"]) == gameLogType {
			item["selected"] = true
		}
	}
	ctx.ViewData("gameLogType", &template.Params{Name: "gameLogType", Title: "游戏账变类型", Options: db.GameLogTypes})

	ctx.View("log/index.html")

}
