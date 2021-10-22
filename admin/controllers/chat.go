package controllers

import (
	"admin/helper"

	"github.com/kataras/iris/v12"
	"github.com/spf13/viper"
)

func (c *controllers) Chat_get(ctx iris.Context) { // GET /chat

	ctx.ViewData("jwt", sess.Get(helper.SESSION_USER).(string))
	ctx.ViewData("url", "http://"+viper.GetString("webserver.addr")+":12307")
	ctx.ViewData("wsurl", "ws://"+viper.GetString("webserver.addr")+":12307/coinws/net")
	ctx.View("chat/index.html")
}
