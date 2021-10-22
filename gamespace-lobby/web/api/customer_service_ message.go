package api

import (
	"errors"
	"log"
	"net/http"
	"strconv"

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

//MakeCustomerServiceMessageService creates the Webservice for customer service message
func MakeCustomerServiceMessageService() http.Handler {
	router := mux.NewRouter()
	router.Handle("/customerservicemessage/getmessagewithplayerid", nex.Handler(getMessageWithPlayerID)).Methods("POST")
	router.Handle("/customerservicemessage/sendmessagetoplayerid", nex.Handler(sendMessageToPlayerID)).Methods("POST")
	// router.Handle("/customerservicemessage/checkadmin", nex.Handler(checkAdminHandler)).Methods("POST")
	router.Handle("/customerservicemessage/getmessageslistingfrompagenumber", nex.Handler(getMessagesListingFromPageNumber)).Methods("POST")
	router.Handle("/customerservicemessage/getunreadmessagesnumber", nex.Handler(getUnreadMessagesNumber)).Methods("POST")
	router.Handle("/customerservicemessage/getmessagelistingwithplayerid", nex.Handler(getMessageListingWithPlayerID)).Methods("POST")
	/*router.Handle("/club/checkuser/", nex.Handler(checkUserHandler)).Methods("POST")
	router.Handle("/club/donate/", nex.Handler(donateToClubHandler)).Methods("POST") // 捐赠
	router.Handle("/club/transaction/", nex.Handler(getTransactionByClubAndUser)).Methods("POST")
	router.Handle("/club/gift/", nex.Handler(clubDonateToUserHandler)).Methods("POST") // 赠送*/
	return router
}

func getMessageListingWithPlayerID(r *http.Request) (map[string]interface{}, error) {
	_, _, isValid, isAdmin := helper.VerifyAdminJWT(r)
	if !isValid {
		payload := map[string]interface{}{"error": "Invalid token"}
		return payload, errors.New("Invalid token")
	}
	reqJSON := helper.ReadParameters(r)
	var payload map[string]interface{}
	//Need to change uncoded values
	payload = db.GetMessageListingWithPlayerID(reqJSON["player_id"], isAdmin)
	return payload, nil
}

func getMessageWithPlayerID(r *http.Request) (map[string]interface{}, error) {
	_, _, isValid, isAdmin := helper.VerifyAdminJWT(r)
	if !isValid {
		payload := map[string]interface{}{"error": "Invalid token"}
		return payload, errors.New("Invalid token")
	}

	reqJSON := helper.ReadParameters(r)
	var payload map[string]interface{}
	//Need to change uncoded values
	payload = db.GetMessagesWithPlayerID(reqJSON["player_id"], isAdmin)
	return payload, nil
}

func sendMessageToPlayerID(r *http.Request) (map[string]interface{}, error) {
	uid, _, isValid, isAdmin := helper.VerifyAdminJWT(r)
	if !isValid {
		payload := map[string]interface{}{"error": "Invalid token"}
		return payload, errors.New("Invalid token")
	}
	reqJSON := helper.ReadParameters(r)
	var payload map[string]interface{}
	reqJSON["message"] = helper.CheckForSensitiveWords(reqJSON["message"])
	payload = db.SendMessageToPlayerID(reqJSON["player_id"], uid, isAdmin, reqJSON["message"])
	updateCustomerServiceMessageRealTime(reqJSON["player_id"])
	return payload, nil
}

func getUnreadMessagesNumber(r *http.Request) (map[string]interface{}, error) {
	uid, _, isValid, isAdmin := helper.VerifyAdminJWT(r)
	if !isValid {
		payload := map[string]interface{}{"error": "Invalid token"}
		return payload, errors.New("Invalid token")
	}
	// reqJSON := helper.ReadParameters(r)
	payload := map[string]interface{}{
		"unreadMessagesNumber": db.GetUnreadMessagesNumber(uid, isAdmin),
	}
	//Need to change uncoded values
	return payload, nil
}

func getMessagesListingFromPageNumber(r *http.Request) (map[string]interface{}, error) {
	uid, _, isValid, isAdmin := helper.VerifyAdminJWT(r)
	if !isValid {
		payload := map[string]interface{}{"error": "Invalid token"}
		return payload, errors.New("Invalid token")
	}
	reqJSON := helper.ReadParameters(r)
	status := "-1"
	/*if reqJSON["isAdmin"] == "true" {
		isAdmin = true
	}*/
	// logger.Printf("isAdmin: %v", isAdmin)
	if reqJSON["tabWanted"] == "Read" {
		status = "1"
	} else if reqJSON["tabWanted"] == "Unread" {
		status = "0"
	}
	pageWanted, err := strconv.Atoi(reqJSON["pageWanted"])
	if err != nil {
		log.Println(err)
	}
	var payload map[string]interface{}
	//Need to change uncoded values
	payload = db.GetMessageListingFromPageNumber(uid, pageWanted, isAdmin, status) // db.CheckAdmin(uid)
	return payload, err
}

// func checkAdminHandler(r *http.Request) (map[string]interface{}, error) {
// 	_, _, isValid, isAdmin := helper.VerifyAdminJWT(r)
// 	if !isValid {
// 		payload := map[string]interface{}{"error": "Invalid token"}
// 		return payload, errors.New("Invalid token")
// 	}
// 	// reqJSON := helper.ReadParameters(r)
// 	payload := map[string]interface{}{
// 		"isAdmin": isAdmin,
// 	}
// 	return payload, nil
// }

func updateCustomerServiceMessageRealTime(playerID string) {
	data := map[string]string{
		"type": "updateCustomerServiceMessageRealTime",
	}
	uid, err := strconv.Atoi(playerID)
	if err != nil {
		log.Println(err)
	}
	LobbyUpdateCustomerServiceMessage(uid, data)
}
