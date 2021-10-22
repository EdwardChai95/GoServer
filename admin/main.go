package main

import (
	"admin/controllers"
	"admin/db"
	"admin/template"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/recover"
	"github.com/kataras/iris/v12/sessions"
	"github.com/kataras/iris/v12/view"
	"github.com/spf13/viper"
)

var (
	tmpl  *view.HTMLEngine
	dbmod *db.DBModule
)

func main() {
	viper.SetConfigType("toml")
	viper.SetConfigFile("./config/config.toml")
	viper.ReadInConfig()
	
	app := iris.New()
	sess := sessions.New(sessions.Config{Cookie: "_session_id", AllowReclaim: true})

	app.Use(sess.Handler())

	tmpl = iris.HTML("./views", ".html")
	tmpl.Delims("{{", "}}") // Set custom delimeters.
	tmpl.Reload(true)       // Enable re-build on local template files changes.
	
	controllers.AppSetupControllers(app)
	template.AppSetupTemplate(tmpl)

	app.RegisterView(tmpl)
	app.HandleDir("/assets", iris.Dir("./assets"))
	app.UseRouter(recover.New()) // Recovery middleware recovers from any panics and writes a 500 if there was one.
	
	dbmod = db.NewDBModule()
	
	app.Listen(viper.GetString("webserver.port"))
}
