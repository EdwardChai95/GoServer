package controllers

import (
	"admin/db"
	"admin/template"
	"strings"

	"github.com/kataras/iris/v12"
)

func (c *controllers) Vcode_get(ctx iris.Context) { // listing /exchangecode
	phone_number := ctx.URLParamDefault("phone_number", "")

	searchParams := map[string]string{
		"phone_number": strings.TrimSpace(phone_number),
	}

	urlParams := ""
	for k, v := range searchParams {
		if strings.TrimSpace(v) != "" {
			urlParams += "&" + k + "=" + v
		}
	}

	code := db.GetVcode(searchParams)
	code_header := []string{
		"电话号码",
		"验证码",
		"创建时间",
	}

	code_data := [][]string{}
	for _, c := range code {
		data := []string{
			c["phone_number"],
			c["v_code"],
			c["create_at"],
		}
		code_data = append(code_data, data)
	}

	ctx.ViewData("phone_number", &template.Params{Name: "phone_number", Title: "电话号码", Value: phone_number})

	ctx.ViewData("code_header", code_header)
	ctx.ViewData("code_data", code_data)

	ctx.View("vcode/index.html")

}
