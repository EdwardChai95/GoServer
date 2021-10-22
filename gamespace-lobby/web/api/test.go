package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lonng/nex"
	proto "gitlab.com/wolfplus/gamespace-lobby/proto"
)

func MakeTestService() http.Handler {
	// log.Printf("called!!~~!!`")

	router := mux.NewRouter()
	router.Handle("/test", nex.Handler(Test)).Methods("GET")
	// router.Handle("/pay/deposit", nex.Handler(depositReq)).Methods("POST")

	return router
}

//处理 测试服务器是否响应 POST请求
func Test(r *http.Request, data *proto.TestRequest) (*proto.TestMessage, error) {

	log.Printf("called!!~~!!`")

	testResponse := &proto.TestMessage{
		Code:    200,
		Message: "ok",
	}

	// testResponse.Code = 200
	// testResponse.Message = "test ok!"

	// Code: 200, Message: "test ok!"

	return testResponse, nil

}
