package api

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lonng/nex"
	"gitlab.com/wolfplus/gamespace-lobby/db"
	"gitlab.com/wolfplus/gamespace-lobby/db/model"
	"gitlab.com/wolfplus/gamespace-lobby/helper"
	proto "gitlab.com/wolfplus/gamespace-lobby/proto"
)

func MakeEventService() http.Handler {
	router := mux.NewRouter()

	router.Handle("/event/createEvent", nex.Handler(createEventHandler)).Methods("POST")
	router.Handle("/event/getEvent", nex.Handler(getEventHandler)).Methods("POST")
	router.Handle("/event/updateEventReward", nex.Handler(updateEventRewardHandler)).Methods("POST")

	return router
}

func updateEventRewardHandler(r *http.Request) (int, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return 0, errors.New("Invalid token")
	}
	db.PlayerObtainedNewYearReward(uid)
	return 0, nil
}

func getEventHandler(r *http.Request, req *proto.GetGuestReq) (map[string]interface{}, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return createEventHandler(r, req)
	}

	event, err := db.GetEvent(model.AuthType(req.Type), helper.StringToInt64(uid))
	if err != nil {
		return createEventHandler(r, req)
	}

	if event == nil {
		return createEventHandler(r, req)
	}

	payload := map[string]interface{}{
		"uid":  event.Uid,
		"day1": event.Day1,
		"day2": event.Day2,
		"day3": event.Day3,
		"day4": event.Day4,
		"day5": event.Day5,
		"day6": event.Day6,
	}

	return payload, nil
}

//处理 前端 创建活动 POST请求
func createEventHandler(r *http.Request, req *proto.GetGuestReq) (map[string]interface{}, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, nil
	}

	user, err := db.CreateEvent(model.AuthType(req.Type), helper.StringToInt64(uid))
	if err != nil {
		payload := map[string]interface{}{
			"code": err.Error(),
		}
		return payload, err
	}

	payload := map[string]interface{}{
		"uid":  user.Uid,
		"day1": user.Day1,
		"day2": user.Day2,
		"day3": user.Day3,
		"day4": user.Day4,
		"day5": user.Day5,
		"day6": user.Day6,
	}

	return payload, nil
}
