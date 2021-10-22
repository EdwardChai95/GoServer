package controllers

import (
	"admin/db"
	"admin/helper"
	"admin/template"

	"github.com/kataras/iris/v12"
)

var ConfigFields = []*template.Params{
	{Name: "ethParams1", Title: "ETH参数 1", IsText: true,
		Help: "为商城支付界面显示的ETH数额调整值，ETH数额=汇率*普通/高级/特级玩家购买价格+参数1"},
	{Name: "ethParams2", Title: "ETH参数 2", IsText: true,
		Help: "为系统应收金额的调整值，系统应收的金额=ETH数额+参数2；"},
}

func (c *controllers) Config_post(ctx iris.Context) { // POST edit /config

	formValues := ctx.FormValues()
	logger.Print(formValues)

	err := db.UpdateConfigs(formValues)
	if err != nil {
		sess.SetFlash(helper.SESSION_TEMPFAILMESSAGE, err.Error())
	} else {
		//sess.SetFlash(helper.SESSION_TEMPMESSAGE, "修改成功")
		sess.SetFlash(helper.SESSION_TEMPMESSAGE, "Đã sửa đổi thành công")
	}

	ctx.Redirect("/config", iris.StatusSeeOther)
}

func (c *controllers) Config_get(ctx iris.Context) { // GET edit /config
	configs := db.GetConfigs()

	ConfigFieldsWithValues := ConfigFields
	for _, field := range ConfigFieldsWithValues {
		field.Value = configs[field.Name]
	}

	ctx.ViewData("form", template.Form{
		Method: "POST",
		Fields: ConfigFieldsWithValues,
	})
	ctx.View("config/edit.html")
}
