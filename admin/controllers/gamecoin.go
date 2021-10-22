package controllers

import (
	"admin/db"
	"admin/helper"
	"admin/template"

	"github.com/kataras/iris/v12"
)

func (c *controllers) Gamecoin_post(ctx iris.Context) {

	adminUid, _ := helper.VerifyJWTString(sess.Get(helper.SESSION_USER).(string))

	formValues := ctx.FormValues()

	err := db.UpdateGameCoinByAdmin(adminUid, formValues)
	if err != nil {
		sess.SetFlash(helper.SESSION_TEMPFAILMESSAGE, err.Error())
	} else {
		sess.SetFlash(helper.SESSION_TEMPMESSAGE, "成功加了游戏币")
	}
	ctx.View("gamecoin/index.html")
}

func (c *controllers) Gamecoin_get(ctx iris.Context) {

	ctx.ViewData("gamecoinform", template.Form{
		Method: "POST",
		Fields: []*template.Params{
			{Name: "uid", Title: "用户ID", Value: "", IsText: true, Help: "例子：123,111313850,111313852"},
			{Name: "update_amt", Title: "加减游戏币", Value: "", IsText: true, Help: "如果是减游戏币请填写负数"},
		},
	})

	ctx.View("gamecoin/index.html")
}
