package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/lonng/nex"
	log "github.com/sirupsen/logrus"
	"gitlab.com/wolfplus/gamespace-lobby/db"
	"gitlab.com/wolfplus/gamespace-lobby/db/model"
	"gitlab.com/wolfplus/gamespace-lobby/define"
	"gitlab.com/wolfplus/gamespace-lobby/helper"
	proto "gitlab.com/wolfplus/gamespace-lobby/proto"
)

var (
	logger = log.WithFields(log.Fields{"component": "http", "service": "login"})
)

func MakeUserService() http.Handler {
	router := mux.NewRouter()

	// router.Handle("/user/updateGameCoin", nex.Handler(updateGameCoinHandler)).Methods("POST")
	router.Handle("/user/changePassword", nex.Handler(changePasswordHandler)).Methods("POST")
	router.Handle("/user/bindGuest", nex.Handler(bindGuestHandler)).Methods("POST")

	router.Handle("/user/basicInfo", nex.Handler(basicInfoHandler)).Methods("POST")

	//login
	// router.Handle("/user/createGuest", nex.Handler(createGuestHandler)).Methods("POST")
	router.Handle("/user/getGuestBasicInfo", nex.Handler(getGuestBasicInfoHandler)).Methods("POST")
	router.Handle("/user/getGuest", nex.Handler(getGuestHandler)).Methods("POST")
	router.Handle("/user/updateLoginReward", nex.Handler(updateLoginRewardHandler)).Methods("POST")
	router.Handle("/user/phoneLogin", nex.Handler(phoneLoginHandler)).Methods("POST")
	router.Handle("/user/updateExperience", nex.Handler(updateExperienceHandler)).Methods("POST")
	router.Handle("/user/checkLoginReward", nex.Handler(checkLoginRewardHandler)).Methods("POST")

	//add 0805
	//router.Handle("/user/getLevel", nex.Handler(getLevelHandler)).Methods("POST")
	//add 1007
	router.Handle("/user/getNode", nex.Handler(getNodeHandler)).Methods("POST")
	router.Handle("/user/getAssets", nex.Handler(getAssetsHandler)).Methods("POST")
	router.Handle("/user/genderChange", nex.Handler(genderChangeHandler)).Methods("POST")
	router.Handle("/user/changeHeadIcon", nex.Handler(changeHeadIconHandler)).Methods("POST")

	//add 1012
	router.Handle("/user/updateCountCompleted", nex.Handler(updateCountCompletedHandler)).Methods("POST")
	return router
}

//add 1012
func updateCountCompletedHandler(r *http.Request) (int, error) {
        uid, isValid := helper.VerifyJWT(r)
        if !isValid {
                return 0, errors.New("Invalid token")
        }
	db.UpdateCountCompleted(uid)
	return 0,nil
}

//add 0805
/*
func getLevelHandler(r *http.Request) (map[string]interface{}, error) {
        uid, isValid := helper.VerifyJWT(r)
        if !isValid {
                return nil, errors.New("Invalid token")
        }

        data := map[string]interface{}{
                "level": db.GetLevel(uid),
        }
        return data, nil
}
*/

//add 1007
func getNodeHandler(r *http.Request) (map[string]interface{}, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}

	data := map[string]interface{}{
		"node": db.GetNode(uid),
	}
	return data, nil
}

func getAssetsHandler(r *http.Request) (map[string]interface{}, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}

	data := map[string]interface{}{
		"winGameCoin": db.GetAssets(uid),
	}
	return data, nil
}

//处理 前端 更新游戏币 POST请求
// 安全问题
// func updateGameCoinHandler(r *http.Request, req *proto.UpdateGameCoinReq) (*proto.UpdateGameCoinRes, error) {
// 	uid, isValid := helper.VerifyJWT(r)
// 	if !isValid {// 		return &proto.UpdateGameCoinRes{Code: "Invalid token"}, errors.New("Invalid token"// 	}
// 	logger.Println("update game coin ", uid)

// 	user, err := db.UpdateGameCoin(model.AuthType(req.Type), helper.StringToInt64(uid),
// 		req.CoinNumber, req.Operate, r.RemoteAddr)

// 	if err != nil {// 		return &proto.UpdateGameCoinRes{Code: err.Error()}, nil// 	}

// 	if user == nil {// 		return &proto.UpdateGameCoinRes{Code: "user not exist"}, nil// 	}

// 	res := &proto.UpdateGameCoinRes{// 		GuestAcc:    user.Uid,
// 		CoinBalance: user.GameCoin,// 	}
// 	logInformation := map[string]string{// 		"uid":       uid,// 		"reason":    "登录奖励",
// 		"otherInfo": fmt.Sprintf("%v", req.CoinNumber// 		"before":    fmt.Sprintf("%v", user.GameCoin-req.CoinNumber),// 		"used":      fmt.Sprintf("%v", req.CoinNumber),
// 		"after":     fmt.Sprintf("%v", user.GameCoin),
// 	}// 	db.SendLogInformation(logInformation)// 	return res, nil
// }

func changeHeadIconHandler(r *http.Request) (map[string]interface{}, error) {
	reqJSON := helper.ReadParameters(r)
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}
	var headIcon string
	headIcon = reqJSON["faceUri"]
	db.UpdateHeadIcon(uid, headIcon)
	fmt.Println("headIcon:", headIcon)
	payload := map[string]interface{}{
		"face_uri": headIcon,
	}

	return payload, nil
}

func genderChangeHandler(r *http.Request) (map[string]interface{}, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}
	var data string
	gender := db.GetGenderByUid(uid)

	if gender == "0" {
		data = "1"
	} else {
		data = "0"
	}

	db.UpdateGender(uid, data)
	payload := map[string]interface{}{
		"gender": data,
	}

	return payload, nil
}

func checkLoginRewardHandler(r *http.Request) (map[string]interface{}, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}

	payload := map[string]interface{}{
		"login": db.CheckLoginRewardStatus(uid),
	}
	return payload, nil
}

func updateExperienceHandler(r *http.Request) (int, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return 0, errors.New("Invalid token")
	}
	reqJSON := helper.ReadParameters(r)
	levelAndExp := db.GetExperienceAndLevel(uid)
	if levelAndExp != nil {
		// log.Println(levelAndExp)
		level, err := strconv.Atoi(levelAndExp["level"])
		if err != nil {
			log.Println(err)
		}
		if level >= 20 {
			return 20, err
		}
		currExp, err := strconv.ParseFloat(levelAndExp["experience"], 64)
		if err != nil {
			log.Println("error 1st")
			log.Println(err)
		}
		expToGain, err := strconv.ParseFloat(reqJSON["experienceToUpdate"], 64)
		if err != nil {
			log.Println("error 2nd")
			log.Println(err)
		}
		newExperience := currExp + expToGain
		levelAndUpperLimitMap := map[int]float64{
			0: 20, 1: 1000, 2: 2000, 3: 4000, 4: 10000, 5: 20000, 6: 40000, 7: 80000, 8: 160000,
			9: 300000, 10: 600000, 11: 1200000, 12: 2400000, 13: 4800000, 14: 10000000, 15: 20000000,
			16: 40000000, 17: 100000000, 18: 200000000, 19: 400000000,
		}
		newLevel := level
		if newExperience >= levelAndUpperLimitMap[level] {
			newLevel = level + 1
			newExperience = 0
		}
		// log.Println(newLevel)
		// log.Println(newExperience)
		newLevelAsString := strconv.Itoa(newLevel)
		newExperienceAsString := strconv.FormatFloat(newExperience, 'f', -1, 64)
		// log.Println(newExperienceAsString)
		db.UpdateLevelAndExpereience(uid, newLevelAsString, newExperienceAsString)
		// log.Println(db.GetExperienceAndLevel(reqJSON["UID"]))
		return newLevel, err
	}
	return 0, nil
}
func updateLoginRewardHandler(r *http.Request) (map[string]interface{}, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}
	// reqJSON := helper.ReadParameters(r) // parameters only got 1 uid
	updatedAmount := db.PlayerObtainedLoginReward(uid)
	
	data := map[string]interface{}{
		"uid":       uid,
		"game_coin": updatedAmount,
	}
	LobbyCoinUpdate(data)
	
	//add 1007
	//db.UpdateWinGameCoin(uid,updatedAmount)
	
	return nil, nil
}

//处理 前端 更新密码 POST请求
func changePasswordHandler(r *http.Request, req *proto.PhoneLoginReq) (*proto.GetGuestRes, error) {
	_, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}

	logger.Println("phone login ", req.UserAcc)

	user, err := db.ChangePassword(model.AuthType(req.Type), req.UserAcc, req.Password, r.RemoteAddr)

	if err != nil {
		return &proto.GetGuestRes{Code: err.Error()}, nil
	}

	if user == nil {
		return &proto.GetGuestRes{Code: "account not exist"}, nil
	}

	return &proto.GetGuestRes{Code: "password changed"}, nil

}

//处理 前端 手机登录 POST请求
func phoneLoginHandler(r *http.Request, req *proto.PhoneLoginReq) (map[string]interface{}, error) {
	logger.Println("phone login ", req.UserAcc)
	user, err := db.PhoneLogin(model.AuthType(req.Type), req.UserAcc, req.Password, r.RemoteAddr)
	if err != nil {
		payload := map[string]interface{}{
			"code": err.Error(),
		}
		return payload, err
	}
	if user == nil {
		payload := map[string]interface{}{
			"code": "password incorrect",
		}
		return payload, err
	}
	item, err := db.GetItem(model.AuthType(req.Type), user.Uid, r.RemoteAddr)
	if err != nil {
		payload := map[string]interface{}{
			"code": err.Error(),
		}
		return payload, err
	}

	uid := strconv.Itoa(int(user.Uid))

	db.UpdateLoginRewardStatus(uid)

	payload := buildUserPayload(user, item)

	jwtString := helper.NewJWT(uid, payload)
	payload["jwt"] = jwtString

	return payload, nil
}

//处理 前端 游客手机绑定 POST请求
func bindGuestHandler(r *http.Request, req *proto.BindGuestReq) (*proto.BindGuestRes, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}
	logger.Infof("binding guest:%d ,phoneNumber:%d,password:%s ", uid, req.PhoneNumber, req.Password, r.RemoteAddr)

	user, err := db.BindGuest(model.AuthType(req.Type), helper.StringToInt64(uid), req.PhoneNumber, req.Password, r.RemoteAddr)

	if err != nil {
		return &proto.BindGuestRes{Code: err.Error()}, nil
	}

	if user == nil {
		//return &proto.BindGuestRes{Code: "已被绑定"}, nil
		return &proto.BindGuestRes{Code: "Đã bị ràng buộc"}, nil
	}

	res := &proto.BindGuestRes{UserAcc: user.UserAcc, Password: user.Password}

	return res, nil
}

//处理 前端 提取用户信息 POST请求 update basic info
func basicInfoHandler(r *http.Request, req *proto.BasicInfoReq) (*proto.BasicInfoRes, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}
	logger.Infof("basic info update, guestAcc:%d, userAcc:%s, nickName:%s,signature:%s", uid, req.UserAcc, req.NickName, req.Signature)

	user, err := db.BasicInfo(model.AuthType(req.Type), helper.StringToInt64(uid), req.UserAcc, req.NickName, req.Signature, r.RemoteAddr)

	if err != nil {
		return &proto.BasicInfoRes{Code: err.Error()}, nil
	}

	res := &proto.BasicInfoRes{
		NickName:  user.NickName,
		Signature: user.Signature,
	}

	return res, err
}

func getGuestByIMEI(imei string) string {
	uid, err := db.GetUidByIMEI(imei)
	if err != nil {
		return ""
	}
	return uid
}

func getGuestById(uid string) (*model.User, map[string]interface{}) {
	user, err := db.GetUser(helper.StringToInt64(uid))
	if err != nil || user == nil {
		log.Println(err)
		return nil, nil
	}
	item, _ := db.GetItem(model.AuthType(1), helper.StringToInt64(uid), "")

	payload := buildUserPayload(user, item)

	return user, payload
}

func getGuestBasicInfoHandler(r *http.Request) (map[string]interface{}, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}
	user, payload := getGuestById(uid) // helper.Int64ToString(req.GuestAcc)
	if user == nil {
		return nil, errors.New("Invalid user")
	}
	return payload, nil
}

//处理 前端 提取游客信息 POST请求
// TODO
func getGuestHandler(r *http.Request) (map[string]interface{}, error) {
	reqJSON := helper.ReadParameters(r)
	var uid, imei string
	if reqJSON["imei"] != "" {
		imei1 := getGuestByIMEI(reqJSON["imei"])
		imei = imei1
	} else {
		imei = ""
	}
	token := r.Header.Get("Authorization")
	if token == "" && imei != "" {
		uid = imei
	} else {
		uid1, isValid := helper.VerifyJWT(r)
		if !isValid {
			if uid1 == define.IPBLOCKEDMESSAGE {
				return map[string]interface{}{"error": define.IPBLOCKEDMESSAGE}, errors.New(define.IPBLOCKEDMESSAGE)
			}
			return createGuestHandler(r, reqJSON)
		}
		uid = uid1
	}

	// logger.Println("account login ", uid)

	db.UpdateLoginRewardStatus(uid) // only for when login
	user, payload := getGuestById(uid)

	if imei != "" && token == "" {
		jwtString := helper.NewJWT(uid, payload)
		payload["jwt"] = jwtString
	}

	// user, err := db.GetGuest(model.AuthType(req.Type), helper.StringToInt64(uid), r.RemoteAddr)
	// res2B, _ := json.Marshal(user)
	// fmt.Println(string(res2B))
	if user == nil { // cannot find user with id
		return createGuestHandler(r, reqJSON)
	}

	if val, ok := reqJSON["myOS"]; ok {
		fmt.Println("val:", val)
		user.LastLoginOS = val
		logger.Println(val)
	}

	db.UpdateLastLogin(*user, user.Uid)

	return payload, nil
}

//处理 前端 创建游客 POST请求
func createGuestHandler(r *http.Request, reqJSON map[string]string) (map[string]interface{}, error) {
	// logger.Println("account login ", req.GuestAcc)
	// intUid := req.GuestAcc

	user, err := db.CreateGuest(r.RemoteAddr, reqJSON)
	if err != nil {
		payload := map[string]interface{}{
			"error": err.Error(),
		}
		return payload, err
	}

	intUid := user.Uid

	item, err := db.CreateItem(intUid, r.RemoteAddr)
	if err != nil {
		payload := map[string]interface{}{
			"error": err.Error(),
		}
		return payload, err
	}

	uid := strconv.Itoa(int(intUid))

	payload := buildUserPayload(user, item)
	payload["isNewUser"] = true

	jwtString := helper.NewJWT(uid, payload)
	payload["jwt"] = jwtString

	logInformation := map[string]string{
		"uid":       uid,
		"reason":    "注册奖励",
		"otherInfo": fmt.Sprintf("%v", payload["gameCoin"]),
	}
	db.SendLogInformation(logInformation)
	db.UpdateLastLogin(*user, user.Uid)
	return payload, nil
}

func buildUserPayload(user *model.User, item *model.Item) map[string]interface{} {
	// use by createguest,getguestbyid and phonelogin

	club_user := db.GetValidClubUserByUserId(helper.Int64ToString(user.Uid))

	payload := map[string]interface{}{
		"uid":            helper.Int64ToString(user.Uid),
		"user_name":      user.Username,
		"face_uri":       user.FaceUri,
		"guestAcc":       user.Uid,
		"userAcc":        user.UserAcc,
		"money":          user.Money,
		"gameCoin":       user.GameCoin,
		"nickName":       user.NickName,
		"signature":      user.Signature,
		"LaBa":           item.LaBa,
		"accLogin":       user.AccLogin,
		"loginReward":    user.LoginReward,
		"userLevel":      user.UserLevel,
		"isAdmin":        helper.IsAdminBool(user),
		"level":          user.Level,
		"viplevel":       user.VipLevel,
		"club_id":        club_user["club_id"],
		"club_user_type": club_user["type"],
		"club_is_mute":   club_user["is_mute"],
		"gender":         user.Gender,
	}
	return payload
}
