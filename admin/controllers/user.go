package controllers

import (
	"admin/db"
	"admin/helper"
	"admin/template"
	"fmt"
	"strings"

	"github.com/kataras/iris/v12"
)

func (c *controllers) UserUpdateGamecoin_post(ctx iris.Context) { // POST edit /user/updategamecoin

	adminUid, _ := helper.VerifyJWTString(sess.Get(helper.SESSION_USER).(string))

	uid := ctx.URLParamDefault("uid", "")
	user := db.GetUserByUid(uid)
	if user == nil { // problem with user means not found or some error
		ctx.Redirect("/user", iris.StatusTemporaryRedirect)
		return
	}
	formValues := ctx.FormValues()
	// logger.Print(formValues)
	err := db.UpdateGameCoinByUid(helper.StringToInt64(uid), adminUid, formValues)
	if err != nil {
		sess.SetFlash(helper.SESSION_TEMPFAILMESSAGE, err.Error())
	} else {
		sess.SetFlash(helper.SESSION_TEMPMESSAGE, "成功加了游戏币")
	}
	ctx.Redirect("/user/edit?uid="+uid, iris.StatusSeeOther)
}

func (c *controllers) UserEdit_post(ctx iris.Context) { // POST edit /user/edit
	uid := ctx.URLParamDefault("uid", "")
	user := db.GetUserByUid(uid)
	if user == nil { // problem with user means not found or some error
		ctx.Redirect("/user", iris.StatusTemporaryRedirect)
		return
	}
	formValues := ctx.FormValues()
	// logger.Print(formValues)
	success := db.UpdateUserByUid(uid, formValues)
	if success {
		sess.SetFlash(helper.SESSION_TEMPMESSAGE, "修改成功")
	} else {
		sess.SetFlash(helper.SESSION_TEMPFAILMESSAGE, "修改失败")

	}
	ctx.Redirect("/user/edit?uid="+uid, iris.StatusSeeOther)

}

func (c *controllers) UserEdit_get(ctx iris.Context) { // GET edit /user/edit
	uid := ctx.URLParamDefault("uid", "")
	user := db.GetUserByUid(uid)
	if user == nil { // problem with user means not found or some error
		ctx.Redirect("/user", iris.StatusTemporaryRedirect)
		return
	}

	userPermissionOptions := []map[string]interface{}{
		{"val": "", "text": "普通玩家"},
		{"val": "admin", "text": "管理员"},
	}

	for _, option := range userPermissionOptions {
		if option["val"] == user["user_permission"] {
			option["selected"] = true
		}
	}

	vipLevelPermissionOptions := []map[string]interface{}{
		{"val": "0", "text": "普通玩家"},
		{"val": "1", "text": "高级玩家"},
		{"val": "2", "text": "特级玩家"},
	}

	for _, option := range vipLevelPermissionOptions {
		if option["val"] == user["vip_level"] {
			option["selected"] = true
		}
	}

	ctx.ViewData("form", template.Form{
		Method: "POST",
		Fields: []*template.Params{
			{Name: "uid", Title: "用户ID", Value: user["uid"], IsReadOnly: true},
			{Name: "rounds_played", Title: "游戏局数", Value: user["rounds_played"], IsReadOnly: true},
			{Name: "total_played", Title: "总输赢", Value: user["total_played"], IsReadOnly: true},
			{Name: "total_win", Title: "赢游戏币", Value: user["total_win"], IsReadOnly: true},
			{Name: "total_lose", Title: "输游戏币", Value: user["total_lose"], IsReadOnly: true},
			{Name: "totalPurchaseAmount", Title: "充值金额", Value: user["totalPurchaseAmount"], IsReadOnly: true},
			{Name: "nick_name", Title: "昵称", Value: user["nick_name"], IsText: true},
			{Name: "user_acc", Title: "手机号", Value: user["user_acc"], IsText: true, Help: "未绑定手机号将会显示1"},
			{Name: "password", Title: "登录密码", Value: "", IsPassword: true, Help: "没修改密码则不需填写"},
			{Name: "user_permission", Title: "权限", Value: user["user_permission"], IsSelect: true,
				Options: userPermissionOptions, Help: "警告：管理员将会有各种特权"},
			{Name: "vip_level", Title: "特级等级", Value: user["vip_level"], IsSelect: true,
				Options: vipLevelPermissionOptions, Help: "玩家ETH购买返利按照这个等级"},
		},
	})
	ctx.ViewData("gamecoinform", template.Form{
		Action: "/user/updategamecoin?uid=" + uid,
		Method: "POST",
		Fields: []*template.Params{
			{Name: "game_coin", Title: "游戏币", Value: user["game_coin"], IsReadOnly: true},
			{Name: "update_amt", Title: "加减游戏币", Value: "", IsText: true, Help: "如果是减游戏币请填写负数"},
			{Name: "comment", Title: "备注", Value: "", IsText: true},
		},
	})
	ctx.View("user/edit.html")
}

func (c *controllers) User_get(ctx iris.Context) { // listing /user
	pageNumber := ctx.URLParamDefault("pageNo", "1")
	uid := ctx.URLParamDefault("uid", "")
	userType := ctx.URLParamDefault("userType", "")
	userAcc1 := ctx.URLParamDefault("userAcc1", "")
	userAcc := ctx.URLParamDefault("userAcc", "")

	levelStart := ctx.URLParamDefault("levelStart", "")
	levelEnd := ctx.URLParamDefault("levelEnd", "")

	dateStartLogin := ctx.URLParamDefault("dateStartLogin", "")
	dateEndLogin := ctx.URLParamDefault("dateEndLogin", "")
	dateStartRegister := ctx.URLParamDefault("dateStartRegister", "")
	dateEndRegister := ctx.URLParamDefault("dateEndRegister", "")

	searchParams := map[string]string{
		"uid":      strings.TrimSpace(uid),
		"userType": strings.TrimSpace(userType),
		"userAcc1": strings.TrimSpace(userAcc1),
		"userAcc":  strings.TrimSpace(userAcc),

		"levelStart": strings.TrimSpace(levelStart),
		"levelEnd":   strings.TrimSpace(levelEnd),

		"dateStartLogin":    strings.TrimSpace(dateStartLogin),
		"dateEndLogin":      strings.TrimSpace(dateEndLogin),
		"dateStartRegister": strings.TrimSpace(dateStartRegister),
		"dateEndRegister":   strings.TrimSpace(dateEndRegister),
	}

	urlParams := ""
	for k, v := range searchParams {
		if strings.TrimSpace(v) != "" {
			urlParams += "&" + k + "=" + v
		}
	}

	data := db.GetUsersByPageNo(pageNumber, urlParams, searchParams)
	if data == nil { // problem with parameters
		ctx.Redirect("/user", iris.StatusTemporaryRedirect)
		return
	}
	for k, v := range data {
		ctx.ViewData(k, v)
	}
	ctx.ViewData("searchTerm", uid)
	ctx.ViewData("pagination", helper.PaginationHTML("/user/", urlParams, helper.StringToInt(pageNumber), data["NumPages"].(int)))

	ctx.ViewData("uid", &template.Params{Name: "uid", Title: "账号ID", Value: uid})

	// userType
	for _, item := range db.UserTypes {
		if fmt.Sprintf("%v", item["val"]) == userType {
			item["selected"] = true
		} else {
			item["selected"] = false
		}
	}
	ctx.ViewData("userType", &template.Params{Name: "userType", Title: "玩家类型", Options: db.UserTypes})

	ctx.ViewData("userAcc1", &template.Params{Name: "userAcc1", Title: "手机号码", Value: userAcc1})
	// userAcc
	for _, item2 := range db.UserAccs {
		if fmt.Sprintf("%v", item2["val"]) == userAcc {
			item2["selected"] = true
		} else {
			item2["selected"] = false
		}
	}
	ctx.ViewData("userAcc", &template.Params{Name: "userAcc", Title: "绑定手机", Options: db.UserAccs})

	ctx.ViewData("levelStart", &template.Params{Name: "levelStart", Title: "等级开始", Value: levelStart})
	ctx.ViewData("levelEnd", &template.Params{Name: "levelEnd", Title: "等级结束", Value: levelEnd})

	ctx.ViewData("dateStartLogin", &template.Params{Name: "dateStartLogin", Title: "最近登陆时间开始", Value: dateStartLogin})
	ctx.ViewData("dateEndLogin", &template.Params{Name: "dateEndLogin", Title: "最近登陆时间结束", Value: dateEndLogin})
	ctx.ViewData("dateStartRegister", &template.Params{Name: "dateStartRegister", Title: "注册时间开始", Value: dateStartRegister})
	ctx.ViewData("dateEndRegister", &template.Params{Name: "dateEndRegister", Title: "注册时间结束", Value: dateEndRegister})

	ctx.View("user/index.html")
}
