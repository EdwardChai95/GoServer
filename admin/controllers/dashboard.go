package controllers

import (
	"admin/template"

	"github.com/kataras/iris/v12"
)

func index(ctx iris.Context) { // main dashboard page

	ctx.ViewData("test", template.Form{
		Fields: []*template.Params{
			{Name: "asdf", Title: "asdf", Value: "asdf"},
			{Name: "asdf", Title: "asdf", Value: "asdf"},
			{Name: "asdf", Title: "asdf", Value: "asdf"},
			{Name: "asdf", Title: "asdf", Value: "asdf"},
		},
	})
	ctx.View("dashboard/index.html")
}

func (c *controllers) DashboardRegisterRoute() {
	app.Get("/", index)
}
