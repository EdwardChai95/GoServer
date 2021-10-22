package api

import (
	"errors"
	"net/http"

	// "time"

	// "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/lonng/nex"

	"gitlab.com/wolfplus/gamespace-lobby/db"
	"gitlab.com/wolfplus/gamespace-lobby/helper"
	// "gitlab.com/wolfplus/gamespace-lobby/db/model"
	// "gitlab.com/wolfplus/gamespace-lobby/define"
	// "gitlab.com/wolfplus/gamespace-lobby/errutil"
)

//MakeLogInformationService creates the Webservice for Log Information
func MakeLogInformationService() http.Handler {
	router := mux.NewRouter()
	router.Handle("/loginformation/sendloginformation", nex.Handler(sendLogInformation)).Methods("POST")
	router.Handle("/loginformation/newwelfare", nex.Handler(newWelfare)).Methods("POST")
	return router
}

func sendLogInformation(r *http.Request) (map[string]interface{}, error) {
	_, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}
	reqJSON := helper.ReadParameters(r)
	var payload map[string]interface{}
	payload = db.SendLogInformation(reqJSON)
	return payload, nil
}

func newWelfare(r *http.Request) (map[string]interface{}, error) {

	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}
	// reqJSON := helper.ReadParameters(r)
	var payload map[string]interface{}
	payload = db.NewWelfare(uid)
	return payload, nil
}
