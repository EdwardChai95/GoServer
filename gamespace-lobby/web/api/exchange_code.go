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

//MakeExchangeCodeService creates the Webservice for exchange codes
func MakeExchangeCodeService() http.Handler {
	router := mux.NewRouter()
	// log.Println("CustomerServiceMessage WebService created")
	router.Handle("/exchangecode/claimcodeifvalid", nex.Handler(claimCodeIfValid)).Methods("POST")
	return router
}

func claimCodeIfValid(r *http.Request) (map[string]interface{}, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		payload := map[string]interface{}{"error": "Invalid token"}
		return payload, errors.New("Invalid token")
	}
	reqJSON := helper.ReadParameters(r)
	var payload map[string]interface{}
	payload = db.ClaimCodeIfValid(uid, reqJSON["code"])
	return payload, nil
}
