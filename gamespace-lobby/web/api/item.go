package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lonng/nex"
	"gitlab.com/wolfplus/gamespace-lobby/db"
	"gitlab.com/wolfplus/gamespace-lobby/db/model"
	"gitlab.com/wolfplus/gamespace-lobby/helper"
	proto "gitlab.com/wolfplus/gamespace-lobby/proto"
)

func MakeItemService() http.Handler {
	// log.Printf("called!!~~!!`")

	router := mux.NewRouter()
	router.Handle("/item/updateItem", nex.Handler(updateItem)).Methods("POST")

	return router
}

//处理 前端 更新游戏道具 POST请求
func updateItem(r *http.Request, req *proto.UpdateItemReq) (*proto.UpdateItemReq, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}
	logger.Println("update item ")

	user, err := db.UpdateItem(model.AuthType(req.Type), helper.StringToInt64(uid), req.ItemName, req.Operate, req.Number, r.RemoteAddr)
	if err != nil {
		return &proto.UpdateItemReq{Code: err.Error()}, nil
	}

	if user == nil {
		return &proto.UpdateItemReq{Code: "account not exist"}, nil
	}
	if req.ItemName == "LaBa" && req.Operate == "minus" {
		logInformation := map[string]string{
			"uid":       uid,
			"reason":    "新手卡",
			"otherInfo": "使用道具" + fmt.Sprintf("%v", req.Number),
		}
		db.SendLogInformation(logInformation)
	}

	return &proto.UpdateItemReq{Code: "item updated"}, nil
}
