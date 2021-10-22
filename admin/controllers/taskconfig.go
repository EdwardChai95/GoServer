package controllers

import (
	"admin/db"
	"admin/helper"
	"admin/template"

	"github.com/kataras/iris/v12"
)

const (
	TASK_TIMEPLAYED        string = "task_timePlayed"
	TASK_TIMEPLAYED_REWARD string = TASK_TIMEPLAYED + "_reward"
	TASK_TIMEPLAYED_GUAGE  string = TASK_TIMEPLAYED + "_guage"

	TASK_ROUNDSPLAYED        string = "task_roundsPlayed"
	TASK_ROUNDSPLAYED_REWARD string = TASK_ROUNDSPLAYED + "_reward"
	TASK_ROUNDSPLAYED_GUAGE  string = TASK_ROUNDSPLAYED + "_guage"

	TASK_WONAMOUNT1        string = "task_wonAmount1"
	TASK_WONAMOUNT1_REWARD string = TASK_WONAMOUNT1 + "_reward"
	TASK_WONAMOUNT1_GUAGE  string = TASK_WONAMOUNT1 + "_guage"

	TASK_WONAMOUNT2        string = "task_wonAmount2"
	TASK_WONAMOUNT2_REWARD string = TASK_WONAMOUNT2 + "_reward"
	TASK_WONAMOUNT2_GUAGE  string = TASK_WONAMOUNT2 + "_guage"

	TASK_WONAMOUNT3        string = "task_wonAmount3"
	TASK_WONAMOUNT3_REWARD string = TASK_WONAMOUNT3 + "_reward"
	TASK_WONAMOUNT3_GUAGE  string = TASK_WONAMOUNT3 + "_guage"
)

var TaskConfigFields = []*template.Params{
	{Name: TASK_TIMEPLAYED, Title: "可调任务名称", IsText: true},
	{Name: TASK_TIMEPLAYED_REWARD, Title: "可调奖励", IsText: true},
	{Name: TASK_TIMEPLAYED_GUAGE, Title: "可调测量", IsText: true},

	{Name: TASK_ROUNDSPLAYED, Title: "可调任务名称", IsText: true},
	{Name: TASK_ROUNDSPLAYED_REWARD, Title: "可调奖励", IsText: true},
	{Name: TASK_ROUNDSPLAYED_GUAGE, Title: "可调测量", IsText: true},

	{Name: TASK_WONAMOUNT1, Title: "可调任务名称", IsText: true},
	{Name: TASK_WONAMOUNT1_REWARD, Title: "可调奖励", IsText: true},
	{Name: TASK_WONAMOUNT1_GUAGE, Title: "可调测量", IsText: true},

	{Name: TASK_WONAMOUNT2, Title: "可调任务名称", IsText: true},
	{Name: TASK_WONAMOUNT2_REWARD, Title: "可调奖励", IsText: true},
	{Name: TASK_WONAMOUNT2_GUAGE, Title: "可调测量", IsText: true},

	{Name: TASK_WONAMOUNT3, Title: "可调任务名称", IsText: true},
	{Name: TASK_WONAMOUNT3_REWARD, Title: "可调奖励", IsText: true},
	{Name: TASK_WONAMOUNT3_GUAGE, Title: "可调测量", IsText: true},
}

func (c *controllers) TaskConfig_post(ctx iris.Context) { // POST edit /taskconfig

	formValues := ctx.FormValues()
	logger.Print(formValues)

	err := db.UpdateConfigs(formValues)
	if err != nil {
		sess.SetFlash(helper.SESSION_TEMPFAILMESSAGE, err.Error())
	} else {
		sess.SetFlash(helper.SESSION_TEMPMESSAGE, "修改成功")
	}

	ctx.Redirect("/taskconfig", iris.StatusSeeOther)
}

func (c *controllers) TaskConfig_get(ctx iris.Context) { // GET edit /taskconfig
	configs := db.GetConfigs()

	ConfigFieldsWithValues := TaskConfigFields
	for _, field := range ConfigFieldsWithValues {
		field.Value = configs[field.Name]
	}

	ctx.ViewData("form", template.Form{
		Method: "POST",
		Fields: TaskConfigFields,
	})
	ctx.View("config/edit.html")
}
