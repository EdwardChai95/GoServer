package controllers

import (
	"admin/db"
	"admin/helper"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"
	log "github.com/sirupsen/logrus"
)

var (
	app    *iris.Application
	logger *log.Entry
	sess   *sessions.Session
	dbmod  *db.DBModule

	CONTROLLERS_DIR = "./controllers"
)

type controllers struct {
}

func withCookieOptions(ctx iris.Context) {
	ctx.AddCookieOptions(iris.CookieHTTPOnly(true), iris.CookieExpires(5*time.Minute))
	ctx.Next()
}

func initControllerHandler() iris.Handler {
	return func(ctx iris.Context) {
		sess = sessions.Get(ctx)
		db.Ctx = ctx
		ctx.Next()
	}
}

func initHandler() iris.Handler {
	return func(ctx iris.Context) {
		if strings.Contains(ctx.GetCurrentRoute().Path(), "/assets") {
			ctx.Next()
			return
		}
		// s := sessions.Get(ctx)
		tempMessage := sess.GetFlash(helper.SESSION_TEMPMESSAGE)
		if tempMessage != "" {
			ctx.ViewData(helper.SESSION_TEMPMESSAGE, tempMessage)
		}
		userVal := sess.Get(helper.SESSION_USER)
		str := fmt.Sprintf("%v", sess.Get(helper.SESSION_USER))

		if str != "<nil>" {
			adminUid, _ := helper.VerifyJWTString(str)
			user := db.GetUserByUid(adminUid)
			if userVal != nil {
				// TODO checking user
				if user["user_permission"] == "super_admin" {
					ctx.ViewData(helper.SESSION_ISLOGGEDIN, true)
				}
				if user["user_permission"] == "admin" {
					ctx.ViewData(helper.SESSION_ISADMINLOGGEDIN, true)
				}
			} else if !strings.Contains(ctx.GetCurrentRoute().Path(), "/auth") {
				// Not logged in
				sess.SetFlash(helper.SESSION_TEMPMESSAGE, "请先登录")
				//sess.SetFlash(helper.SESSION_TEMPMESSAGE, "vui lòng đăng nhập trước")
				//ctx.Redirect("/auth", iris.StatusTemporaryRedirect)
				return
			}
		}
		ctx.ViewLayout("layout/main.html")
		ctx.Next()
	}
}

func AppSetupControllers(mainApp *iris.Application) {
	logger = log.WithField("source", "controllers")

	c := &controllers{}
	cVal := reflect.ValueOf(c)

	app = mainApp
	app.Use(initControllerHandler())
	app.Use(initHandler())
	app.Use(withCookieOptions)

	files, err := ioutil.ReadDir(CONTROLLERS_DIR)
	if err != nil {
		logger.Fatal(err)
	}

	for _, f := range files {
		// logger.Println(f.Name())
		fileName := f.Name()

		if !strings.Contains(fileName, "controllers") { // ignore this file
			funcPrefix := strings.Replace(fileName, ".go", "", -1)
			registerRouteFuncName := strings.Title(funcPrefix + "RegisterRoute")
			m := cVal.MethodByName(registerRouteFuncName)
			if !m.IsValid() {
				// logger.Print("Register route func not found: ", registerRouteFuncName)
				// continue
			} else {
				m.Call(nil)
			}

			// read functions from controllers
			fname := CONTROLLERS_DIR + "/" + fileName
			file, err := os.Open(fname)
			if err != nil {
				logger.Warn(err)
				continue
			}
			defer file.Close()

			// read the whole file in
			srcbuf, err := ioutil.ReadAll(file)
			if err != nil {
				logger.Warn(err)
				continue
			}
			src := string(srcbuf)

			// file set
			fset := token.NewFileSet()
			astFile, err := parser.ParseFile(fset, fileName, src, 0)
			if err != nil {
				logger.Warn(err)
				continue
			}

			// main inspection
			ast.Inspect(astFile, func(n ast.Node) bool {

				switch fn := n.(type) {
				// catching all function declarations
				// other intersting things to catch FuncLit and FuncType
				case *ast.FuncDecl:
					var funcName = fn.Name.Name
					if strings.Contains(funcName, "_get") || strings.Contains(funcName, "_post") {
						handler := cVal.MethodByName(funcName)

						if handler.IsValid() {
							funcRoute := strings.ToLower(strings.Replace(funcName, "_get", "", -1))
							funcRoute = strings.ToLower(strings.Replace(funcRoute, "_post", "", -1))
							funcRoute = strings.Replace(funcRoute, funcPrefix, "", -1)
							relativePath := "/" + funcPrefix + "/" + funcRoute

							if strings.Contains(funcName, "_get") {
								// logger.Print("init GET " + relativePath)
								app.Get(relativePath, handler.Interface().(func(iris.Context)))
							} else if strings.Contains(funcName, "_post") {
								// logger.Print("init POST " + relativePath)
								app.Post(relativePath, handler.Interface().(func(iris.Context)))
							}
						}
					}

				}

				return true
			})
		}
	}

}
