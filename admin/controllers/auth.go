package controllers

import (
	"admin/db"
	"admin/helper"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
)

type AuthController struct {
}

func (c *AuthController) Post(ctx iris.Context) {
	// username: 1610684668247
	// pw: 666666

	formValues := ctx.FormValues()
	user := db.GetUserByUid(formValues["uid"][0])
	// logger.Print(user)

	// check user is it admin
	if user["user_permission"] == "super_admin" {
		// user["password"]
		passwordSlice := strings.Split(user["password"], helper.PASSWORDSEPERATOR)
		if helper.VerifyPassword(formValues["password"][0], passwordSlice[1], passwordSlice[0]) {
			sess.Set(helper.SESSION_USER, helper.NewJWT(user["uid"]))
			ctx.Redirect("/", iris.StatusSeeOther) // log in succeed
			return
		}
	}

	if user["user_permission"] == "admin" {
		// user["password"]
		passwordSlice := strings.Split(user["password"], helper.PASSWORDSEPERATOR)
		if helper.VerifyPassword(formValues["password"][0], passwordSlice[1], passwordSlice[0]) {
			sess.Set(helper.SESSION_USER, helper.NewJWT(user["uid"]))
			userVal := sess.Get(helper.SESSION_USER)
			if userVal=="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbnVpZCI6IjExMTMxMzg1MCIsInVpZCI6IjExMTMxMzg1MCJ9.uAM0w5DkGbu-uSGSri2FP5xL1hAq3Xjtj4Ri28SP6rk" || userVal == "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbnVpZCI6IjExMTMxMzg1OSIsInVpZCI6IjExMTMxMzg1OSJ9.kp2FS-jSbiBi9780d617WCcMpOx-ndc-bvJVNJ-BZno" || userVal == "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbnVpZCI6IjExMTMxNTY4MCIsInVpZCI6IjExMTMxNTY4MCJ9.2rE__45GZ3lZOac2Zneyd8Wo99DCI5BQcpYLGZ0U8As" || userVal == "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbnVpZCI6IjExMTMxMzk3NyIsInVpZCI6IjExMTMxMzk3NyJ9.JFJ75Bc3MvjYfm7MN8xTFT1jbe3nSuNf7K4a_GINK5w" {
				ctx.Redirect("/chat", iris.StatusSeeOther) // log in succeed
				return
			} else {
				ctx.Redirect("/auth", iris.StatusSeeOther)
				return
			}
		}
	}

	//sess.SetFlash(helper.SESSION_TEMPMESSAGE, "错误：登陆失败")
	sess.SetFlash(helper.SESSION_TEMPMESSAGE, "Lỗi: đăng nhập không thành công")
	ctx.Redirect("/auth", iris.StatusSeeOther) // if log in failed
	// ctx.View("auth/login.html")
}

func (c *AuthController) Get(ctx iris.Context) {
	sess.Clear()
	ctx.View("auth/login.html")
}

func (c *controllers) AuthRegisterRoute() {
	authAPI := app.Party("/auth")
	a := mvc.New(authAPI)
	a.Handle(new(AuthController))
}
