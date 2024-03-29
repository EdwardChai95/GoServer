package template

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"strings"

	"github.com/kataras/iris/v12/view"
	log "github.com/sirupsen/logrus"
)

var (
	logger    *log.Entry
	blocksDir = "./views/blocks"
	funcMap   = template.FuncMap{}
)

type Form struct {
	Action string
	Method string
	Fields []*Params
}

type Params struct {
	Type    string
	Name    string
	Title   string
	Value   interface{}
	Options interface{}
	Help    string
	// special options
	OnChangeSubmit bool
	// classification
	IsText     bool
	IsReadOnly bool
	IsPassword bool
	IsSelect   bool
}

func customExecution(funcName string) interface{} {
	return func(params interface{}) template.HTML {
		var tpl bytes.Buffer
		parsedTemplate, err := template.New(funcName + ".html").Funcs(funcMap).ParseFiles(blocksDir + "/" + funcName + ".html")
		if err != nil {
			logger.Warn(err)
		}
		err = parsedTemplate.Execute(&tpl, params)
		if err != nil {
			logger.Warn(err)
			return "Error executing template"
		} else {
			return template.HTML(tpl.String())
		}
	}
}

func AppSetupTemplate(tmpl *view.HTMLEngine) {
	logger = log.WithField("source", "template")

	blocks, err := ioutil.ReadDir(blocksDir)
	if err != nil {
		logger.Fatal(err)
	}

	for _, f := range blocks {
		funcName := f.Name()
		funcName = strings.Replace(funcName, ".html", "", -1)
		// logger.Println(funcName)
		funcMap[funcName] = customExecution(funcName)

		tmpl.AddFunc(funcName, customExecution(funcName))
	}

}
