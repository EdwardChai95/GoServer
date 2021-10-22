package api

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lonng/nex"
	"gitlab.com/wolfplus/gamespace-lobby/db"
	"gitlab.com/wolfplus/gamespace-lobby/helper"
)

func MakeProxyService() http.Handler {
	router := mux.NewRouter()

	router.Handle("/proxy/checkProxy", nex.Handler(checkProxyHandler)).Methods("POST")
	router.Handle("/proxy/today", nex.Handler(todayHandler)).Methods("POST")
	router.Handle("/proxy/total", nex.Handler(totalHandler)).Methods("POST")
	router.Handle("/proxy/getLists", nex.Handler(getListHandler)).Methods("GET")
	return router
}

func checkProxyHandler(r *http.Request) (map[string]interface{}, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}

	proxy := db.CheckProxyByUid(uid)
	payload := map[string]interface{}{
		"proxy": proxy,
	}
	return payload, nil
}

func todayHandler(r *http.Request) (map[string]interface{}, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}

	today := db.GetTodayByUid(uid)

	payload := map[string]interface{}{
		"promo_num":       today[0]["promo_num"],
		"active_num":      today[0]["active_num"],
		"count_completed": today[0]["count_completed"],
	}
	return payload, nil
}

func totalHandler(r *http.Request) (map[string]interface{}, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}

	total := db.GetTotalByUid(uid)

	payload := map[string]interface{}{
		"promo_num":       total[0]["promo_num"],
		"active_num":      total[0]["active_num"],
		"count_completed": total[0]["count_completed"],
	}
	return payload, nil
}

func getListHandler(r *http.Request) (map[string]string, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}

	lists := db.GetListByUid(uid)
	if len(lists) > 0 {
		return lists[0], nil
	}

	return nil, nil
}
