package db

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"gitlab.com/wolfplus/gamespace-lobby/algoutil"
	"gitlab.com/wolfplus/gamespace-lobby/db/model"
	"gitlab.com/wolfplus/gamespace-lobby/helper"
)

//add 1012
func UpdateCountCompleted(uid string) {
	user, err1 := GetUser(helper.StringToInt64(uid))
	if err1 != nil {
		log.Println(err1)
	}
	u, err := db.QueryString("select * from `user` where uid = '" + uid + "' and proxy != 0")
	proxyadmin, err := db.QueryString("select * from `proxy` where uid = '" + u[0]["proxy"] + "' and date(operating_time) >= curdate() LIMIT 1")
	proxyuser, err := db.QueryString("select * from `proxy_user` where uid = '" + uid + "' and date(operating_time) >= curdate() LIMIT 1")
	code, err := db.QueryString("select * from `exchange_code` where proxy_uid = '" + u[0]["proxy"] + "' LIMIT 1")
	if err != nil {
		log.Println(err)
	}
	if len(proxyadmin) > 0 {
		_, err := db.Exec("`Update `proxy` set `count_completed` = `count_completed` + 1 where uid = '" + u[0]["proxy"] + "'")
		if err != nil {
			log.Println(err)
		}
	} else {
		_, err := db.Exec("INSERT INTO `proxy`(`uid`, `promo_code`, `promo_num`, `active_num`, `send_num`, `receive_num`, `total_num`, `total_amount`, `service_tax`, `count_completed`, `operating_time`)"+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", uid, code[0]["exchange_code"], 0, 0, 0, 0, 0, 0, 0, 1, helper.GetCurrentShanghaiDateOnlyString())
		if err != nil {
			log.Println(err)
		}
	}

	if len(proxyuser) > 0 {
		_, err := db.Exec("`Update `proxy_user` set `count_completed` = `count_completed` + 1 where uid = '" + uid + "'")
		if err != nil {
			log.Println(err)
		}
	} else {
		_, err := db.Exec("INSERT INTO `proxy_user`(`uid`, `proxy_uid`, `total_win`, `total_lose`, `total_win_lose`, `total_broad`, `send_num`, `receive_num`, `total_amount`, `count_completed`, `operating_time`)"+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", uid, u[0]["proxy"], 0, 0, 0, 0, 0, 0, 0, 1, helper.GetCurrentShanghaiDateOnlyString())
		if err != nil {
			log.Println(err)
		}
	}

	_, err2 := db.Exec("Update `user` set `count_completed = ` = 2 WHERE `uid` = '" + uid + "'")
	if err2 != nil {
		log.Println(err2)
	}

	ucc := UpdateGameCoin(user, 10000, "积累游戏币领奖励", "积累游戏币领奖励", fmt.Sprintf("%v", 10000), "")
	logger.Println("ucc: ", ucc)

}

func UpdateGameCoin(user *model.User, amount int64,
	transaction_comment, log_reason, log_otherInfo, task_key string) int64 {
	uid := user.Uid
	// user, _ := GetUser(helper.StringToInt64(uid)) // before

	_, err := db.Exec("update `user` set `game_coin` = `game_coin` + ? where `uid` = ?",
		helper.Int64ToString(amount), uid)
	if err != nil {
		logger.Println(err)
	}

	_, err2 := db.Exec("INSERT INTO `game_coin_transaction`(`uid`,"+
		" `value`, `type`, `comment`, `datetime`) VALUES (?, ?, ?, ?, ?)", uid,
		helper.Int64ToString(amount), "游戏币", transaction_comment, helper.GetCurrentShanghaiTimeString())
	if err2 != nil {
		logger.Println(err2)
	}

	updatedUser, _ := GetUser(uid) // after
	logInformation := map[string]string{
		"uid":       fmt.Sprintf("%v", uid),
		"reason":    log_reason,
		"otherInfo": log_otherInfo,
		"before":    fmt.Sprintf("%v", user.GameCoin),
		"used":      fmt.Sprintf("%v", amount),
		"after":     fmt.Sprintf("%v", updatedUser.GameCoin),
		"task_key":  task_key,
	}
	SendLogInformation(logInformation)

	return updatedUser.GameCoin
}

//add 1007
func UpdateWinGameCoin(user string, amount int64) {
	//uid := user.Uid
	_, err := db.Exec("update `user` set `win_game_coin` = `win_game_coin` + ? where `uid` = ?",
		helper.Int64ToString(amount), user)
	if err != nil {
		logger.Println(err)
	}
	//updatedUser, _ := GetUser(uid) // after
	//return updatedUser.WinGameCoin
}

// func UpdateGameCoin(ty model.AuthType, guestAcc int64, coinNumber int64, operate string, ip string) (*model.User, error) {
// 	logger.Infof("update game coin db side, userAcc:%d, coinNumber:%d , operate:%s", guestAcc, coinNumber, operate)

// 	u := model.User{Uid: guestAcc}
// 	_, _ = db.Get(&u)

// 	switch operate {
// 	case "plus":
// 		u.GameCoin += coinNumber
// 		break
// 	case "minus":
// 		u.GameCoin -= coinNumber
// 		break
// 	}

// 	_, _ = db.ID(guestAcc).Cols("game_coin").Update(&u)

// 	return &u, nil
// }

//change password buggy
func ChangePassword(ty model.AuthType, userAcc int64, password string, ip string) (*model.User, error) {
	logger.Infof("change password db side, userAcc:%d, password:%s", userAcc, password)

	u := model.User{}

	has, err := db.Where("user_acc=?", userAcc).Get(&u)

	if err != nil {
		logger.Println("the error is " + err.Error())
		return nil, err
	}

	if !has {
		logger.Println("try get phone user but it not exist")
		return nil, nil
	} else {
		hash, salt := algoutil.PasswordHash(password)
		u.Password = hash + helper.PASSWORDSEPERATOR + salt
		// u.Password = password

		_, err = db.Where("user_acc=?", userAcc).Omit("uid").Update(u)

		return &model.User{}, nil
	}

}

func PhoneLogin(ty model.AuthType, userAcc int64, password string, ip string) (*model.User, error) {
	logger.Infof("phone login db side, userAcc:%d, password:%s", userAcc, password)

	u := model.User{}

	has, err := db.Where("user_acc=?", userAcc).Get(&u)

	if err != nil {
		logger.Println("the error is " + err.Error())
		return nil, err
	}

	if !has {
		logger.Println("try get phone user but it not exist")
		return nil, nil
	}

	logger.Infof("get password " + u.Password)

	passwordSlice := strings.Split(u.Password, helper.PASSWORDSEPERATOR)

	if !algoutil.VerifyPassword(password, passwordSlice[1], passwordSlice[0]) {
		logger.Println("password incorrect")
		return nil, nil
	}

	u = UpdateLastLogin(u, u.Uid)

	return &u, nil
}

// not used
func BindGuest2(ty model.AuthType, guestAcc int64, phoneNumber int64, password string, ip string) (*model.User, error) {
	logger.Infof("bind guest db side, guestAcc:%d, phoneNumber:%d,password:%s ", guestAcc, phoneNumber, password)

	//check if request phone number is already exisit in record
	checkU := &model.User{}

	has, err := db.Where("user_acc=?", phoneNumber).Get(checkU)

	if err != nil {
		logger.Println("the error is " + err.Error())
	}

	if has {
		// u := &model.User{}
		// _, err := db.ID(guestAcc).Get(u)

		// checkU.UserAcc = phoneNumber
		checkU.Password = password

		_, err = db.Where("user_acc=?", phoneNumber).Update(checkU)

		if err != nil {
			logger.Println("error here ", err)
		}

		return checkU, nil
	} else {
		return nil, nil
	}

}

func BindGuest(ty model.AuthType, guestAcc int64, phoneNumber int64, password string, ip string) (*model.User, error) {
	logger.Infof("bind guest db side, guestAcc:%d, phoneNumber:%d,password:%s ", guestAcc, phoneNumber, password)

	//check if request phone number is already exisit in record
	checkU := model.User{}

	has, err := db.Where("user_acc=?", phoneNumber).Get(&checkU)

	if err != nil {
		logger.Println("the error is " + err.Error())
	}

	if !has {
		u := &model.User{}
		_, err := db.ID(guestAcc).Get(u)
		//https://www.youtube.com/watch?v=el-f84rh4aY&t=2184s 27:31

		u.UserAcc = phoneNumber
		hash, salt := algoutil.PasswordHash(password)
		u.Password = hash + helper.PASSWORDSEPERATOR + salt

		_, err = db.ID(guestAcc).Update(u)
		message, err := db.QueryString("select * from `user` where uid = '" + helper.Int64ToString(guestAcc) + "' LIMIT 1")
		if message[0]["normal_active"] == "0" {
			_, err = db.Exec("Update `user` set `normal_active` = 1 where uid = '" + helper.Int64ToString(guestAcc) + "' ")
		}

		if err != nil {
			logger.Println("error here ", err)
		}

		return u, nil
	} else {
		//return nil, errors.New("已被绑定")
		return nil, errors.New("Ràng buộc")
	}

}

// this is actually an update of user function
func BasicInfo(ty model.AuthType, guestAcc int64, userAcc int64, nickName string, signature string, ip string) (*model.User, error) {
	logger.Infof("basic info db side, guestAcc:%d, userAcc:%s, nickName:%s,signature:%s", guestAcc, userAcc, nickName, signature)

	if userAcc != 0 {
		logger.Println("update use userAcc")
	} else {
		logger.Println("update use guestAcc")
	}

	// sql := "update User set NickName = ? , Signature = ? where GuestAcc = ?"
	// res, err := db.Exec(sql, nickName, signature, guestAcc)

	u := &model.User{
		// NickName:  nickName,
		// Signature: signature,
	}
	_, err := db.ID(guestAcc).Get(u)
	//https://www.youtube.com/watch?v=el-f84rh4aY&t=2184s 27:31

	u.NickName = nickName
	u.Signature = signature

	_, err = db.ID(guestAcc).Update(u)

	if err != nil {
		logger.Println("error here ", err)
	}

	return u, nil
}

func UpdateHeadIcon(uid, faceUri string) {
	_, err := db.Exec("Update `user` set `face_uri` = '" + faceUri + "' WHERE `uid` = '" + uid + "'")
	if err != nil {
		log.Println(err)
	}
}

func UpdateGender(uid, gender string) {
	_, err := db.Exec("Update `user` set `gender` = '" + gender + "' WHERE `uid` = '" + uid + "'")
	if err != nil {
		log.Println(err)
	}
}

func GetGenderByUid(uid string) string {
	message, err := db.QueryString("select gender from `user` " +
		"where uid = '" + uid + "' LIMIT 1")
	if err != nil {
		log.Println(err)
	}

	return message[0]["gender"]
}

func GetUidByIMEI(imei string) (string, error) {
	messages, err := db.QueryString("select uid from `user` " +
		"where imei = '" + imei + "' LIMIT 1")
	if err != nil {
		log.Println(err)
	}
	if len(messages) > 0 {
		return messages[0]["uid"], nil
	}
	return "", err
}

func GetUser(userAcc int64) (*model.User, error) {
	// AuthType not used, ip not
	u := model.User{Uid: userAcc}
	has, err := db.Get(&u)

	if err != nil {
		logger.Println(err)
	}

	if !has {
		logger.Println("try get guest acc but it not exist")
		return nil, err
	}

	return &u, err
}

//CheckLastLogin check last login time
func CheckLastLogin(uid string) int {
	messages, err := db.QueryString("select last_login_at from `user` " +
		"where uid = '" + uid + "' LIMIT 1")
	if err != nil {
		log.Println(err)
	}

	if len(messages) > 0 {
		current := time.Now()
		format := "2006-01-02 15:04:05"
		now := current.Format(format)
		fmt.Println("now:", now)
		diff := getHourDiff(messages[0]["last_login_at"], now)
		if diff >= 2 {
			return 1
		} else {
			return 0
		}
	}
	return 0
}

func getHourDiff(start, end string) int64 {
	var hour int64
	format := "2006-01-02 15:04:05"
	t1, err := time.ParseInLocation(format, start, time.Local)
	t2, err := time.ParseInLocation(format, end, time.Local)
	if err == nil && t1.Before(t2) {
		diff := t2.Unix() - t1.Unix()
		hour = diff / 3600
		fmt.Println(diff)
		return hour
	} else {
		return hour
	}
}

//InsertVCode insert vcode of player with stated uid
func CreateVCode(phoneNumber string, vcode string) {
	_, err := db.Exec("INSERT INTO `v_code`(`phone_number`,"+
		" `v_code`, `create_at`) VALUES (?, ?, ?)", helper.StringToInt64(phoneNumber),
		vcode, helper.GetCurrentShanghaiTimeString())
	if err != nil {
		logger.Println(err)
	}
}

//GetExperienceAndLevel gets the experience and level of a certain uid string
func GetExperienceAndLevel(uid string) map[string]string {
	messages, err := db.QueryString("select * from `user` where `uid` = '" + uid + "' Limit 1")
	if err != nil {
		log.Println(err)
	}
	if len(messages) > 0 {
		payload := map[string]string{
			"level":      messages[0]["level"],
			"experience": messages[0]["experience"],
		}
		return payload
	} else {
		return nil
	}
}

//UpdateLevelAndEXP updates the level and experience of the player with the stated uid
func UpdateLevelAndExpereience(uid string, level string, experience string) {
	_, err := db.Exec("Update `user` set `level` = '" + level + "', `experience` = '" + experience + "' WHERE `uid` = '" + uid + "'")
	if err != nil {
		log.Println(err)
	}
}

func CreateGuest(ip string, reqJSON map[string]string) (*model.User, error) {
	shanghaiTime := helper.GetCurrentShanghaiTime()
	u := &model.User{
		// Uid:         guestAcc, // autoincr
		Username:    "",
		FaceUri:     int64(rand.Intn(10-1) + 1),
		Money:       0,
		CreateAt:    shanghaiTime,
		LastLoginAt: shanghaiTime,
		LastLoginIp: ip,
		//haige added
		// GuestAcc:    guestAcc,
		UserAcc:  1,
		GameCoin: 10000,
		//NickName:    "游客",
		NickName: "Khách",
		//Signature:   "取一个个性签名吧",
		Signature:   "chữ ký",
		Level:       0,
		Experience:  0,
		AccLogin:    1,
		LoginReward: "y",
	}

	if val, ok := reqJSON["myOS"]; ok {
		fmt.Println("val:", val)
		u.CreateOS = val
	}
	if val, ok := reqJSON["imei"]; ok { // 设备id
		fmt.Println("val1:", val)
		u.Imei = val
	}
	_, err := db.Insert(u)
	if err != nil {
		logger.Println("insert error ", err.Error())
		return nil, err
	}

	return u, nil
}

//PlayerObtainedLoginReward updates when player obtained the login reward for the day
func PlayerObtainedLoginReward(uid string) int64 {
	user, err1 := GetUser(helper.StringToInt64(uid))
	if err1 != nil {
		log.Println(err1)
	}
	// TODO
	// logger.Println(user)
	amountToUpdate := getLoginRewardAmountByDay(user.AccLogin)

	updatedAmount := UpdateGameCoin(user, amountToUpdate,
		"登陆奖励", "登陆奖励", fmt.Sprintf("%v", amountToUpdate),
		"")
	//add 1022
	UpdateWinGameCoin(uid,amountToUpdate)

	s := "Update `user` set `login_reward` = 'n', `acc_login` = `acc_login` + 1, `login_reward_claim_time` = '" +
		helper.GetCurrentShanghaiTimeString() + "' WHERE uid = '" + uid + "';"
	//log.Println(s)
	_, err := db.Exec(s)
	if err != nil {
		log.Println(err)
	}

	return updatedAmount
}

//CheckAdmin checks if individual is an admin
func CheckAdmin(playerID string) bool {
	messages, err := db.QueryString("select * from `user` " +
		"where `user_permission` = 'admin' and uid = '" + playerID + "' LIMIT 1")
	if err != nil {
		log.Println(err)
	}
	if len(messages) > 0 {
		return true
	} else {
		return false
	}
}

//CheckLoginRewardStatus checks login reward status
func CheckLoginRewardStatus(playerID string) int {
	messages, err := db.QueryString("select login_reward from `user` " +
		"where uid = '" + playerID + "' LIMIT 1")
	if err != nil {
		log.Println(err)
	}
	if len(messages) > 0 {
		fmt.Println(messages[0]["login_reward"])
		if messages[0]["login_reward"] == "y" {
			return 2
		} else {
			return 1
		}
	}
	return 1
}

//getlevel 0805 added
/*func GetLevel(playerID string) int {
        messages, err := db.QueryString("select level from `user` " +
                "where uid = '" + playerID + "' LIMIT 1")
        if err != nil {
                log.Println(err)
        }
        if len(messages) > 0 {
                fmt.Println(messages[0]["level"])
                return messages[0]["level"]
        }
        return 1
}*/

//get 1007
func GetNode(playerID string) int {
	messages1, err1 := db.QueryString("select level from `user` " +
		"where uid = '" + playerID + "' LIMIT 1")
	if err1 != nil {
		log.Println(err1)
	}

	if len(messages1) > 0 {
		fmt.Println("level " + messages1[0]["level"])
	}

	level, _ := strconv.Atoi(messages1[0]["level"])
	if level > 5 {
		return 0
	}

	//add 1012
	messages2, err2 := db.QueryString("select win_game_coin from `user` " +
		"where uid = '" + playerID + "' ")
	if err2 != nil {
		log.Println(err2)
	}

	if len(messages2) > 0 {
		fmt.Println("win_game_coin " + messages2[0]["win_game_coin"])
	}

	win_game_coin, _ := strconv.Atoi(messages2[0]["win_game_coin"])
	if win_game_coin >= 1000000 {
		sqlStr1 := "update `user` set `count_completed` = 1 " +
			"where uid = '" + playerID + "' "
		_, err := db.Exec(sqlStr1)
		if err != nil {
			log.Println(err)
		}
	}

	messages3, err3 := db.QueryString("select count_completed from `user` " +
		"where uid = '" + playerID + "' LIMIT 1")
	if err3 != nil {
		log.Println(err3)
	}
	count_completed, _ := strconv.Atoi(messages3[0]["count_completed"])
	if len(messages3) > 0 {
		fmt.Println(messages3[0]["count_completed"])
		return count_completed + 1
	}

	fmt.Println("error in GetNode func")
	return -1

}

func GetAssets(playerID string) int {
	messages, err := db.QueryString("select win_game_coin from `user` " +
		"where uid = '" + playerID + "' LIMIT 1")
	if err != nil {
		log.Println(err)
	}
	win_game_coin, _ := strconv.Atoi(messages[0]["win_game_coin"])
	if len(messages) > 0 {
		fmt.Println(messages[0]["win_game_coin"])
		return win_game_coin
	}
	return 0
}

//UpdateLoginRewardStatus checks and update the login reward statuses
func UpdateLoginRewardStatus(playerID string) {
	// allow user to collect update if more than "1" day since login
	sqlStr1 := "Update `user` set `login_reward` = 'y' WHERE DATEDIFF(('" +
		helper.GetCurrentShanghaiTimeString() + "'), (`login_reward_claim_time`)) >= 1" +
		" and `uid` = '" + playerID + "'"
	// update user to the first login reward if more than 2 days since login
	sqlStr2 := "Update `user` set `acc_login` = '1' WHERE DATEDIFF(('" +
		helper.GetCurrentShanghaiTimeString() + "'), (`login_reward_claim_time`)) >= 2 and " +
		"`uid` = '" + playerID + "'"
	// log.Println(sqlStr1)
	// log.Println(sqlStr2)
	_, err := db.Exec(sqlStr1)
	if err != nil {
		log.Println(err)
	}
	_, err = db.Exec(sqlStr2)
	if err != nil {
		log.Println(err)
	}
}

func UpdateLastLogin(u model.User, guestAcc int64) model.User {

	log_reason := "登陆"
	// logger.Println(log_reason)
	logs, err := db.QueryString("SELECT COUNT(*) as num_result FROM `log_information` " +
		"WHERE uid = '" + fmt.Sprintf("%v", u.Uid) + "' AND DATE(`operating_time`)='" +
		helper.GetCurrentShanghaiDateOnlyString() + "' " +
		"AND `reason`='" + log_reason + "'" +
		"ORDER BY log_information_id DESC LIMIT 1")

	// logger.Printf("logs", helper.StringToInt(logs[0]["num_result"]))

	if len(logs) == 0 || helper.StringToInt(logs[0]["num_result"]) == 0 {
		logInformation := map[string]string{
			"uid":    fmt.Sprintf("%v", u.Uid),
			"reason": log_reason,
		}
		SendLogInformation(logInformation)
	}

	u.LastLoginAt = helper.GetCurrentShanghaiTime()

	log.Println("updated login reward：", u.LoginReward)

	affect, err := db.ID(guestAcc).Update(&u)

	if err != nil {
		log.Println("更新登陆时间错误：", err)
	}
	log.Println("更新登陆时间成功：", affect)

	return u

}

func getLoginRewardAmountByDay(day int64) int64 {
	switch day {
	case 1:
		return 10000
	case 2:
		return 12000
	case 3:
		return 14000
	case 4:
		return 16000
	case 5:
		return 18000
	case 6:
		return 20000
	case 7:
		return 30000
	default:
		return 30000
	}
}
