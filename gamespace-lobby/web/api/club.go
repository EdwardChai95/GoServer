package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	// "time"

	// "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/lonng/nex"
	"gitlab.com/wolfplus/gamespace-lobby/db"
	"gitlab.com/wolfplus/gamespace-lobby/db/model"
	"gitlab.com/wolfplus/gamespace-lobby/helper"

	// "gitlab.com/wolfplus/gamespace-lobby/db/model"
	// "gitlab.com/wolfplus/gamespace-lobby/define"
	// "gitlab.com/wolfplus/gamespace-lobby/errutil"
	proto "gitlab.com/wolfplus/gamespace-lobby/proto"
)

func MakeClubService() http.Handler {
	router := mux.NewRouter()

	router.Handle("/club/get/", nex.Handler(getClubHandler)).Methods("POST")
	router.Handle("/club/getallclubusers/", nex.Handler(getAllClubUsersHandler)).Methods("POST")
	router.Handle("/club/donate/", nex.Handler(donateToClubHandler)).Methods("POST")   // 捐赠
	router.Handle("/club/gift/", nex.Handler(clubDonateToUserHandler)).Methods("POST") // 赠送
	router.Handle("/club/searchuser/", nex.Handler(searchUserHandler)).Methods("POST")
	router.Handle("/club/searchclubuser/", nex.Handler(searchClubUserHandler)).Methods("POST")
	router.Handle("/club/transaction/", nex.Handler(getClubTransactionHandler)).Methods("POST")
	router.Handle("/club/getclubinfo/", nex.Handler(getClubInfoHandler)).Methods("POST")
	router.Handle("/club/create/", nex.Handler(createClubHandler)).Methods("POST")
	router.Handle("/club/join/", nex.Handler(newJoinRequestHandler)).Methods("POST")
	router.Handle("/club/updateclub/", nex.Handler(updateClubHandler)).Methods("POST")
	router.Handle("/club/retrieveclubgamecoin/", nex.Handler(retrieveClubGamecoinHandler)).Methods("POST")
	router.Handle("/club/getviceadmin/", nex.Handler(getViceAdminHandler)).Methods("POST")
	router.Handle("/club/updateclubuser/", nex.Handler(updateClubUserHandler)).Methods("POST")

	// TEMP FOR CLUB
	router.Handle("/club/temp/giftuser/", nex.Handler(giftGameCoinToUserHandler)).Methods("POST")
	return router
}

// search all user, even not in same club
func searchUserHandler(r *http.Request) (map[string]interface{}, error) {
	_, isValid := helper.VerifyJWTGetClubUser(r)
	if !isValid {
		payload := map[string]interface{}{"error": "Invalid token"}
		return payload, errors.New("Invalid token")
	}
	reqJson := helper.ReadParameters(r)
	// this func search user within club only
	// user := db.GetUserInSameClub(reqJson["Uid"], clubUser["club_id"])
	user, _ := db.GetUser(helper.StringToInt64(reqJson["Uid"]))

	if user == nil {
		//payload := map[string]interface{}{"error": "未查询到该玩家，请确认ID"}
		payload := map[string]interface{}{"error": "Vui lòng xác nhận ID"}
		return payload, nil
	} else {
		payload := map[string]interface{}{"user": user}
		return payload, nil
	}
}

// search all users within same club
func searchClubUserHandler(r *http.Request) (map[string]interface{}, error) {
	clubUser, isValid := helper.VerifyJWTGetClubUser(r)
	if !isValid {
		payload := map[string]interface{}{"error": "Invalid token"}
		return payload, errors.New("Invalid token")
	}
	reqJson := helper.ReadParameters(r)
	// this func search user within club only
	user := db.GetUserInSameClub(reqJson["Uid"], clubUser["club_id"])

	if user == nil {
		//payload := map[string]interface{}{"error": "未查询到该玩家，请确认ID"}
		payload := map[string]interface{}{"error": "Vui lòng xác nhận ID"}
		return payload, nil
	} else {
		payload := map[string]interface{}{"user": user}
		return payload, nil
	}
}

func clubDonateToUserHandler(r *http.Request) (map[string]interface{}, error) {
	clubUser, isValid := helper.VerifyJWTGetClubUser(r)
	if !isValid {
		payload := map[string]interface{}{"error": "Invalid token"}
		return payload, errors.New("Invalid token")
	}
	reqJson := helper.ReadParameters(r)
	//payload := map[string]interface{}{"error": "赠送失败"}
	payload := map[string]interface{}{"error": "sự thất bại"}

	user, _ := db.GetUser(helper.StringToInt64(reqJson["Uid"]))
	if user == nil {
		//payload := map[string]interface{}{"error": "未查询到该玩家，请确认ID"}
		payload := map[string]interface{}{"error": "Vui lòng xác nhận ID"}
		return payload, nil
	}
	if v1, v2, v3 := clubUser["club_id"], reqJson["Uid"], reqJson["amount"]; v1 != "" && v2 != "" && v3 != "" {
		success, reason := db.ClubDonateToUser(v1, v2, v3)
		if success {
			giftedUser, _ := db.GetUser(helper.StringToInt64(reqJson["Uid"]))
			data := map[string]interface{}{
				"uid":       user.Uid,
				"game_coin": giftedUser.GameCoin, //need giftuser.gamecoin to update
			}
			LobbyCoinUpdate(data)
			logInformation := map[string]string{
				"uid":       reqJson["Uid"],
				"reason":    "公会游戏",
				"otherInfo": "赠送数量" + reqJson["amount"],
				"before":    fmt.Sprintf("%v", user.GameCoin),
				"used":      reqJson["amount"],
				"after":     fmt.Sprintf("%v", giftedUser.GameCoin),
			}
			db.SendLogInformation(logInformation)
			//payload = map[string]interface{}{"success": "赠送成功", "user": user}
			payload = map[string]interface{}{"success": "sự thành công", "user": user}
		} else {
			if reason != "" {
				payload = map[string]interface{}{"error": reason}
			}
		}
	}

	return payload, nil
}

func getClubTransactionHandler(r *http.Request) (map[string]interface{}, error) {
	clubUser, isValid := helper.VerifyJWTGetClubUser(r)
	if !isValid {
		payload := map[string]interface{}{"error": "Invalid token"}
		return payload, errors.New("Invalid token")
	}
	reqJson := helper.ReadParameters(r)
	//payload := map[string]interface{}{"error": "失败"}
	payload := map[string]interface{}{"error": "sự thất bại"}
	club_id := fmt.Sprintf("%v", clubUser["club_id"])
	transactions := db.GetClubTransactions(club_id, reqJson, db.TransactionType(reqJson["TransactionType"]))

	var user *model.User
	if reqJson["Uid"] != "" {
		user, _ = db.GetUser(helper.StringToInt64(reqJson["Uid"]))
	}

	if len(transactions) > 0 {
		payload = map[string]interface{}{"success": transactions}
		if user != nil {
			payload = map[string]interface{}{"success": transactions, "user": user}
		}
	} else {
		//payload = map[string]interface{}{"error": "未查询到该玩家，请确认ID"}
		payload = map[string]interface{}{"error": "Vui lòng xác nhận ID"}
		if user != nil {
			//payload = map[string]interface{}{"error": "没有赠送记录", "user": user}
			payload = map[string]interface{}{"error": "Không có lịch sử", "user": user}
		}
	}
	return payload, nil
}

func donateToClubHandler(r *http.Request) (map[string]interface{}, error) {
	clubUser, isValid := helper.VerifyJWTGetClubUser(r)
	if !isValid {
		payload := map[string]interface{}{"error": "Invalid token"}
		return payload, errors.New("Invalid token")
	}
	reqJson := helper.ReadParameters(r)
	// log.Println ( reqJson )

	//payload := map[string]interface{}{"error": "捐赠失败"}
	payload := map[string]interface{}{"error": "sự thất bại"}

	if v1, v2, v3 := clubUser["club_id"], clubUser["uid"], reqJson["amount"]; v1 != "" && v2 != "" && v3 != "" {
		success, reason := db.DonateToClub(v1, v2, v3)
		if success {
			user, _ := db.GetUser(helper.StringToInt64(clubUser["uid"]))
			admin := db.GetAdminByClubId(clubUser["club_id"])
			donateAmount, _ := strconv.ParseInt(reqJson["amount"], 10, 64)
			logInformation := map[string]string{
				"uid":       clubUser["uid"],
				"reason":    "公会游戏",
				"otherInfo": "捐赠数量" + reqJson["amount"],
				"before":    fmt.Sprintf("%v", user.GameCoin+donateAmount),
				"used":      fmt.Sprintf("%v", -donateAmount),
				"after":     fmt.Sprintf("%v", user.GameCoin),
			}
			db.SendLogInformation(logInformation)
			//payload = map[string]interface{}{"success": "捐赠成功", "user": user, "admin": admin}
			payload = map[string]interface{}{"success": "sự thành công", "user": user, "admin": admin}
		} else {
			if reason != "" {
				payload = map[string]interface{}{"error": reason}
			}
		}
	}
	return payload, nil
}

func getClubHandler(r *http.Request, req *proto.GetClubReq) (map[string]interface{}, error) {
	_, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}
	clubs, err := db.GetClubList(req.ClubName)
	if err != nil {
		log.Println(err)
	}
	if len(clubs) > 0 {
		totalMembers, err := db.GetNumMembers(clubs[0]["club_id"])
		admin := db.GetAdminByClubId(clubs[0]["club_id"])
		payload := map[string]interface{}{
			"club_id":          clubs[0]["club_id"],
			"level":            clubs[0]["level"],
			"club_name":        clubs[0]["club_name"],
			"club_description": clubs[0]["description"],
			"club_admin":       admin["nick_name"],
			"club_admin_level": admin["level"],
			"total_member":     totalMembers,
		}
		return payload, err
	}
	return nil, nil
}

func getAllClubUsersHandler(r *http.Request) ([]map[string]string, error) { // retarded func! DONT USE
	clubUser, isValid := helper.VerifyJWTGetClubUser(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}
	return db.GetAllClubUsers(clubUser["club_id"]), nil
}

func getClubInfoHandler(r *http.Request) (map[string]interface{}, error) {
	clubUser, isValid := helper.VerifyJWTGetClubUser(r)
	logger.Printf("clubUser: %v", clubUser)
	if !isValid {
		payload := map[string]interface{}{"error": "Invalid token"}
		return payload, errors.New("Invalid token")
	}
	// reqJSON := helper.ReadParameters(r)
	//payload := map[string]interface{}{"error": "读取失败"}
	payload := map[string]interface{}{"error": "sự thất bại"}
	club_id := clubUser["club_id"]
	clubIdInString := fmt.Sprintf("%v", clubUser["club_id"])
	joinedClub, err1 := db.GetClubInfoByClubId(club_id)
	numMembersInClub, err2 := db.GetNumMembers(club_id)
	if err1 != nil && err2 != nil {
		log.Println(err1)
		log.Println(err2)
	}
	if joinedClub != nil && numMembersInClub != "0" {
		//payload = map[string]interface{}{"success": "读取成功",
		payload = map[string]interface{}{"success": "sự thành công",
			"club":  joinedClub,
			"total": numMembersInClub,
		}
		if clubUser["club_user_type"] == "admin" || clubUser["club_user_type"] == "user2" {
			payload["club_members"] = db.GetAllClubUsers(club_id)
			payload["club_members_donation"] = db.GetClubTransactions(clubIdInString,
				map[string]string{
					"ClubId": club_id,
					"Uid":    "",
				}, db.CLUB_TRANS_TYPE_DONATE)
		} else {
			payload["user_donation"] = db.GetClubTransactions(clubIdInString,
				map[string]string{
					"ClubId": club_id,
					"Uid":    clubUser["uid"],
				}, db.CLUB_TRANS_TYPE_DONATE)
		}
	}
	return payload, nil
}

func createClubHandler(r *http.Request) (map[string]interface{}, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		payload := map[string]interface{}{"error": "Invalid token"}
		return payload, errors.New("Invalid token")
	}

	//payload := map[string]interface{}{"error": "创建工会失败"}
	payload := map[string]interface{}{"error": "sự thất bại"}

	user, _ := getGuestById(uid)
	if user.UserAcc == 1 && user.Level < 3 {
		return payload, errors.New(payload["error"].(string))
	}

	reqJSON := helper.ReadParameters(r)
	clubNameToLowerCase := strings.ToLower(reqJSON["ClubName"])
	//payload["error"] = "创建工会失败，名称重复"
	payload["error"] = "Tên trùng lặp"
	if db.CanNewClub(clubNameToLowerCase, uid) {
		if db.CreateNewClub(clubNameToLowerCase, uid) {
			//payload = map[string]interface{}{"success": "创建成功"}
			payload = map[string]interface{}{"success": "sự thành công"}
		}
	}
	return payload, nil
}

func newJoinRequestHandler(r *http.Request) (map[string]interface{}, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		payload := map[string]interface{}{"error": "Invalid token"}
		return payload, errors.New("Invalid token")
	}

	reqJSON := helper.ReadParameters(r)
	//payload := map[string]interface{}{"error": "加入工会失败"}
	payload := map[string]interface{}{"error": "sự thất bại"}

	user, _ := getGuestById(uid)
	if user.UserAcc == 1 && user.Level < 1 {
		return payload, errors.New(payload["error"].(string))
	}
	if db.CanJoinClub(uid, reqJSON["ClubId"]) {
		if db.PlayerCanJoin(uid, reqJSON["ClubId"]) {
			//payload = map[string]interface{}{"success": "要求加入公会成功"}
			payload = map[string]interface{}{"success": "Gửi yêu cầu thành công"}
		}
	}
	return payload, nil
}

func updateClubHandler(r *http.Request) (map[string]interface{}, error) {
	_, isValid := helper.VerifyJWT(r)
	if !isValid {
		payload := map[string]interface{}{"error": "Invalid token"}
		return payload, errors.New("Invalid token")
	}
	reqJSON := helper.ReadParameters(r)
	//payload := map[string]interface{}{"error": "编辑失败"}
	payload := map[string]interface{}{"error": "sự thất bại"}

	if db.UpdateClubContent(reqJSON) {
		//payload = map[string]interface{}{"success": "编辑成功"}
		payload = map[string]interface{}{"success": "sự thành công"}
	}
	return payload, nil
}

func retrieveClubGamecoinHandler(r *http.Request) (map[string]interface{}, error) {
	clubUser, isValid := helper.VerifyJWTGetClubUser(r)
	if !isValid {
		payload := map[string]interface{}{"error": "Invalid token"}
		return payload, errors.New("Invalid token")
	}
	reqJSON := helper.ReadParameters(r)
	//payload := map[string]interface{}{"error": "提取失败"}
	payload := map[string]interface{}{"error": "sự thất bại"}

	if clubUser["club_user_type"] == "admin" {
		if v1, v2, v3 := clubUser["club_id"], clubUser["uid"], reqJSON["amount"]; v1 != "" && v2 != "" && v3 != "" {
			success, reason := db.RetrieveClubGameCoinFromClub(v1, v2, v3)
			if success {
				admin := db.GetAdminByClubId(clubUser["club_id"])
				club, err := db.GetClubInfoByClubId(clubUser["club_id"])
				if err != nil {
					log.Println(err)
				}
				retrieveAmount, _ := strconv.ParseInt(reqJSON["amount"], 10, 64)
				gamecoin, _ := strconv.ParseInt(admin["game_coin"], 10, 64)
				logInformation := map[string]string{
					"uid":       clubUser["uid"],
					"reason":    "公会游戏",
					"otherInfo": "提取数量" + reqJSON["amount"],
					"before":    fmt.Sprintf("%v", gamecoin-retrieveAmount),
					"used":      fmt.Sprintf("%v", retrieveAmount),
					"after":     fmt.Sprintf("%v", gamecoin),
				}
				db.SendLogInformation(logInformation)
				//payload = map[string]interface{}{"success": "提取成功", "admin": admin, "club": club}
				payload = map[string]interface{}{"success": "sự thành công", "admin": admin, "club": club}
			} else {
				if reason != "" {
					payload = map[string]interface{}{"error": reason}
				}
			}
		}
	}
	return payload, nil
}

func getViceAdminHandler(r *http.Request) (map[string]interface{}, error) {
	clubUser, isValid := helper.VerifyJWTGetClubUser(r)
	if !isValid {
		payload := map[string]interface{}{"error": "Invalid token"}
		return payload, errors.New("Invalid token")
	}
	//payload := map[string]interface{}{"error": "读取失败"}
	payload := map[string]interface{}{"error": "sự thất bại"}
	

	result, err := db.GetClubViceAdminByClubId(clubUser["club_id"])
	if err != nil {
		log.Println(err)
	}
	if result != nil {
		//payload = map[string]interface{}{"success": "提取成功", "result": result}
		payload = map[string]interface{}{"success": "sự thành công", "result": result}
	}
	return payload, nil
}

func updateClubUserHandler(r *http.Request) (map[string]interface{}, error) {
	clubUser, isValid := helper.VerifyJWTGetClubUser(r)
	if !isValid {
		payload := map[string]interface{}{"error": "Invalid token"}
		return payload, errors.New("Invalid token")
	}
	reqJSON := helper.ReadParameters(r)
	//payload := map[string]interface{}{"error": "升级管理员失败"}
	payload := map[string]interface{}{"error": "nâng cấp không thành công"}
	var maxNumsOfViceAdminInClub int

	result, err1 := db.GetClubInfoByClubId(clubUser["club_id"])
	result2, err2 := db.GetNumsOfViceAdminExistInClub(clubUser["club_id"])
	clubLevel, _ := strconv.Atoi(result["level"])
	numsOfViceAdminExist, _ := strconv.Atoi(result2)
	if err1 != nil && err2 != nil {
		log.Println(err1)
		log.Println(err2)
	}

	switch clubLevel {
	case 1, 2:
		maxNumsOfViceAdminInClub = 1
	case 3, 4:
		maxNumsOfViceAdminInClub = 2
	case 5, 6, 7, 8:
		maxNumsOfViceAdminInClub = 3
	default:
		maxNumsOfViceAdminInClub = 0
	}
	if clubLevel == 0 {
		//payload = map[string]interface{}{"error": "0级公会不可设置管理员"}
		payload = map[string]interface{}{"error": "nâng cấp không thành công"}
		return payload, nil
	}
	if reqJSON["UpgradeType"] == string(db.CLUB_UPGRADE_VICE_ADMIN) {
		if numsOfViceAdminExist < maxNumsOfViceAdminInClub {
			if db.CheckClubUserTypeByUid(clubUser["club_id"], reqJSON["Uid"]) {
				if db.UpdateClubUser(clubUser["club_id"], reqJSON["Uid"], reqJSON["UpgradeType"]) {
					//payload = map[string]interface{}{"success": "升级管理员成功"}
					payload = map[string]interface{}{"success": "sự thành công"}
				}
			} else {
				payload = map[string]interface{}{"error": "Đã là quản trị viên"}
				//payload = map[string]interface{}{"success": "升级管理员成功"}
			}
		} else {
			payload = map[string]interface{}{"error": "Số lượng đạt đến giới hạn trên"}
			//payload = map[string]interface{}{"success": "升级管理员成功"}
		}
	}

	if reqJSON["UpgradeType"] == string(db.CLUB_DOWNGRADE_VICE_ADMIN) {
		if db.UpdateClubUser(clubUser["club_id"], reqJSON["Uid"], reqJSON["UpgradeType"]) {
			//payload = map[string]interface{}{"success": "解除管理员成功"}
			payload = map[string]interface{}{"success": "sự thành công"}
		} else {
			//payload = map[string]interface{}{"error": "解除管理员失败"}
			payload = map[string]interface{}{"error": "sự thất bại"}
		}
	}
	return payload, nil
}

func giftGameCoinToUserHandler(r *http.Request) (map[string]interface{}, error) {
	user, isValid := helper.VerifyJWTGetClubUser(r)
	if !isValid {
		payload := map[string]interface{}{"error": "Invalid token"}
		return payload, errors.New("Invalid token")
	}
	// logger.Printf("user %v", user)

	reqJSON := helper.ReadParameters(r)
	receiver_uid := reqJSON["Uid"]
	// logger.Printf("receiver_uid %v", receiver_uid)
	amountToTransfer := helper.StringToInt64(reqJSON["amount"])
	if success, message, updatedSenderAmount, updatedReceiverAmount :=
		db.TransferGameCoin(user, receiver_uid, amountToTransfer); success {

		data := map[string]interface{}{
			"uid":       user["uid"],
			"game_coin": updatedSenderAmount,
		}
		LobbyCoinUpdate(data) // minus my money live update

		data2 := map[string]interface{}{
			"uid":       receiver_uid,
			"game_coin": updatedReceiverAmount,
		}
		LobbyCoinUpdate(data2) // update receiver money live update

		return map[string]interface{}{
			"success": message,
		}, nil
	} else {
		return map[string]interface{}{
			"error": message,
		}, nil
	}

	// return payload, nil

	// reqJSON := helper.ReadParameters(r)
	// payload := map[string]interface{}{"error": "赠送失败"}
	// giftUser, _ := db.GetUser(helper.StringToInt64(reqJSON["Uid"]))
	// giftAmount, _ := strconv.Atoi(reqJSON["amount"])
	// if giftUser == nil {
	// 	payload = map[string]interface{}{"error": "未查询到该玩家，请确认ID"}
	// 	return payload, nil
	// }
	// // check urself isadmin, isAdmin can gift to all users.
	// // !isAdmin gift to admin only
	// if user["isAdmin"] == "true" {
	// 	if user["uid"] == fmt.Sprintf("%v", giftUser.Uid) {
	// 		payload = map[string]interface{}{"error": "不能赠送给自己"}
	// 		return payload, nil
	// 	} else {
	// 		if v1, v2, v3 := user["uid"], reqJSON["Uid"], reqJSON["amount"]; v1 != "" && v2 != "" && v3 != "" {
	// 			success, reason := db.UserPermissionAdminGiftToUser(v1, v2, v3)
	// 			if success {
	// 				logInformation := map[string]string{
	// 					"uid":       reqJSON["Uid"],
	// 					"reason":    "被赠",
	// 					"otherInfo": "赠送数量" + reqJSON["amount"],
	// 					"before":    fmt.Sprintf("%v", giftUser.GameCoin),
	// 					"used":      reqJSON["amount"],
	// 					"after":     fmt.Sprintf("%v", giftUser.GameCoin+int64(giftAmount)),
	// 				}
	// 				db.SendLogInformation(logInformation)
	// 				payload = map[string]interface{}{"success": "赠送成功", "user": giftUser}
	// 			} else {
	// 				if reason != "" {
	// 					payload = map[string]interface{}{"error": reason}
	// 				}
	// 			}
	// 		}
	// 	}
	// } else {
	// 	// only can gift to user_permission = admin
	// 	if giftUser.UserPermission == "admin" {
	// 		if v1, v2, v3 := user["uid"], reqJSON["Uid"], reqJSON["amount"]; v1 != "" && v2 != "" && v3 != "" {
	// 			success, reason := db.UserDonateToUserPermissionAdmin(v1, v2, v3)
	// 			if success {
	// 				donateUser, _ := db.GetUser(helper.StringToInt64(user["uid"]))
	// 				logInformation := map[string]string{
	// 					"uid":       user["uid"],
	// 					"reason":    "赠出",
	// 					"otherInfo": "捐赠数量" + reqJSON["amount"],
	// 					"before":    fmt.Sprintf("%v", donateUser.GameCoin+int64(giftAmount)),
	// 					"used":      fmt.Sprintf("%v", -giftAmount),
	// 					"after":     fmt.Sprintf("%v", donateUser.GameCoin),
	// 				}
	// 				db.SendLogInformation(logInformation)
	// 				payload = map[string]interface{}{"success": "捐赠成功", "user": user}
	// 			} else {
	// 				if reason != "" {
	// 					payload = map[string]interface{}{"error": reason}
	// 				}
	// 			}
	// 		}
	// 	} else {
	// 		payload = map[string]interface{}{"error": "对方无权接受赠送"}
	// 		return payload, nil
	// 	}
	// }
	// data := map[string]interface{}{
	// 	"uid":       giftUser.Uid,
	// 	"game_coin": giftUser.GameCoin + int64(giftAmount),
	// }
	// LobbyCoinUpdate(data)
	// return payload, nil
}
