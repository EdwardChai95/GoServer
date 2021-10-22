package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

//处理 用户 访问进入游戏 GET请求
func MakeFileService() http.Handler {
	router := mux.NewRouter()
	fs := http.FileServer(http.Dir("./static/"))
	router.PathPrefix("/file/").Handler(http.StripPrefix("/file/", fs))
	return router
}
