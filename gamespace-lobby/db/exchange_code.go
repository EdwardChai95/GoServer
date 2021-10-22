package db

import (
	// "gitlab.com/wolfplus/gamespace-lobby/db/model"
	"log"
	"strconv"

	"gitlab.com/wolfplus/gamespace-lobby/helper"
)

//ClaimCodeIfValid uses up the code provided if valid
func ClaimCodeIfValid(uid string, code string) map[string]interface{} {
	validCodes, err := db.QueryString("select * from `exchange_code` where `exchange_code` = '" + code + "' and `valid_until_time` > UNIX_TIMESTAMP(NOW()) LIMIT 1")
	if err != nil {
		log.Println(err)
	}
	if len(validCodes) <= 0 {
		payload := map[string]interface{}{
			"la_ba":     0,
			"game_coin": 0,
			//"failed_reason": "~无效礼包码~",
			"failed_reason": "~Mã gói không hợp lệ~",
		}
		return payload
	}
	timesCodeUsed, err := strconv.Atoi(validCodes[0]["is_used"])
	maxUsage, err := strconv.Atoi(validCodes[0]["max_usage"])
	usedBefore, err := db.QueryString("select * from `log_information` where `reason` = '使用兑换码' and `other_info` = '" + code + "' and `uid` ='" + uid + "'")
	proxycheck, err := db.QueryString("select * from `user` where proxy != 0 and uid = '" + uid + "'")
	if err != nil {
		log.Println(err)
	}

	if validCodes[0]["code_type"] == "0" {
		if timesCodeUsed >= maxUsage {
			payload := map[string]interface{}{
				"la_ba":     0,
				"game_coin": 0,
				//"failed_reason": "~该礼包码已经使用~",
				"failed_reason": "~Đã được sử dụng~",
			}
			return payload
		}
		if timesCodeUsed > 0 {
			if len(usedBefore) > 0 {
				payload := map[string]interface{}{
					"la_ba":     0,
					"game_coin": 0,
					//"failed_reason": "~该礼包码已经使用~",
					"failed_reason": "~Đã được sử dụng~",
				}
				return payload
			}
			return returnSuccessfulClaim(validCodes[0], uid)
		}
		return returnSuccessfulClaim(validCodes[0], uid)
	}
	//validCodes[0]["code_type"] == "1"
	if maxUsage == 1 {
		if timesCodeUsed > 0 {
			if len(usedBefore) <= 0 {
				payload := map[string]interface{}{
					"la_ba":     0,
					"game_coin": 0,
					//"failed_reason": "~该礼包码已经使用~",
					"failed_reason": "~Đã được sử dụng~",
				}
				return payload
			}
			payload := map[string]interface{}{
				"la_ba":     0,
				"game_coin": 0,
				//"failed_reason": "~该礼包码不可重复使用~",
				"failed_reason": "~Không thể tái sử dụng~",
			}
			return payload
		}
		return returnSuccessfulClaim(validCodes[0], uid)
	}
	if timesCodeUsed >= maxUsage {
		payload := map[string]interface{}{
			"la_ba":     0,
			"game_coin": 0,
			//"failed_reason": "~该礼包码已经使用~",
			"failed_reason": "~Đã được sử dụng~",
		}
		return payload
	}
	if len(proxycheck) > 0 {
		if validCodes[0]["proxy_uid"] != "0" {
			payload := map[string]interface{}{
				"la_ba":         0,
				"game_coin":     0,
				"failed_reason": "~不可重复领取代理礼包码~",
			}
			return payload
		}
	}

	if len(usedBefore) > 0 {
		payload := map[string]interface{}{
			"la_ba":     0,
			"game_coin": 0,
			//"failed_reason": "~该礼包码不可重复使用~",
			"failed_reason": "~Không thể tái sử dụng~",
		}
		return payload
	}
	return returnSuccessfulClaim(validCodes[0], uid)
}

func returnSuccessfulClaim(validCodes map[string]string, uid string) map[string]interface{} {
	laba, err := strconv.Atoi(validCodes["la_ba"])
	if err != nil {
		log.Println(err)
	}
	gamecoin, err := strconv.Atoi(validCodes["game_coin"])
	if err != nil {
		log.Println(err)
	}
	proxy, err := strconv.Atoi(validCodes["proxy_uid"])
	if err != nil {
		log.Println(err)
	}
	reqJSON := map[string]string{
		"uid":       uid,
		"reason":    "使用兑换码",
		"otherInfo": validCodes["exchange_code"],
	}
	SendLogInformation(reqJSON)
	SendCodeInformation(reqJSON)
	if laba > 0 {
		_, err := db.Exec("Update `item` set `la_ba` = `la_ba` + " + validCodes["la_ba"] + " WHERE `uid` = '" + uid + "'")
		if err != nil {
			log.Println(err)
		}
		reqJSON := map[string]string{
			"uid":       uid,
			"reason":    "新手卡",
			"otherInfo": "获得道具" + validCodes["la_ba"],
		}
		SendLogInformation(reqJSON)
		SendCodeInformation(reqJSON)
	}
	if gamecoin > 0 {
		_, err := db.Exec("Update `user` set `game_coin` = `game_coin` + '" + validCodes["game_coin"] + "' WHERE `uid` = '" + uid + "'")
		if err != nil {
			log.Println(err)
		}
		_, err = db.Exec("INSERT INTO `game_coin_transaction`(`uid`, `value`, `type`, `comment`, `datetime`) VALUES (?, ?, ?, ?, ?)", uid, validCodes["game_coin"], "游戏币", "新手卡", helper.GetCurrentShanghaiTimeString())
		if err != nil {
			log.Println(err)
		}
		reqJSON := map[string]string{
			"uid":       uid,
			"reason":    "新手卡",
			"otherInfo": "获得数量" + validCodes["game_coin"],
		}
		SendLogInformation(reqJSON)
		SendCodeInformation(reqJSON)
	}
	if proxy > 0 {
		_, err := db.Exec("Update `user` set `proxy` = '" + validCodes["proxy_uid"] + "' WHERE `uid` = '" + uid + "'")
		if err != nil {
			log.Println(err)
		}
		proxyadmin, err := db.QueryString("select * from `proxy` where uid = '" + validCodes["proxy_uid"] + "' and operating_time >= '" + helper.GetCurrentShanghaiDateOnlyString() + "' LIMIT 1")
		proxyuser, err := db.QueryString("select * from `proxy_user` where operating_time <= '" + helper.GetCurrentShanghaiDateOnlyString() + "' and uid = '" + uid + "'")
		if err != nil {
			log.Println(err)
		}
		if len(proxyuser) == 0 {
			_, err := db.Exec("INSERT INTO `proxy_user`(`uid`, `proxy_uid`, `total_win`, `total_lose`, `total_win_lose`, `total_broad`, `send_num`, `receive_num`, `total_amount`, `count_completed`, `operating_time`)"+
				"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", uid, validCodes["proxy_uid"], 0, 0, 0, 0, 0, 0, 0, 0, helper.GetCurrentShanghaiDateOnlyString())
			if err != nil {
				log.Println(err)
			}
		}
		if len(proxyadmin) == 0 {
			_, err := db.Exec("INSERT INTO `proxy`(`uid`, `promo_code`, `promo_num`, `active_num`, `send_num`, `receive_num`, `total_num`, `total_amount`, `service_tax`, `count_completed`, `operating_time`)"+
				"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", validCodes["proxy_uid"], validCodes["exchange_code"], 1, 0, 0, 0, 0, 0, 0, 0, helper.GetCurrentShanghaiDateOnlyString())
			if err != nil {
				log.Println(err)
			}
		} else {
			_, err := db.Exec("`Update `proxy` set `promo_num` = `promo_num` + 1 where uid = '" + validCodes["proxy_uid"] + "'")
			if err != nil {
				log.Println(err)
			}
		}

	}
	messages2, err := db.Exec("Update `exchange_code` set `is_used` = `is_used` + 1 WHERE `exchange_code_id` = '" + validCodes["exchange_code_id"] + "'")
	if err != nil {
		log.Println(err)
	}
	log.Println(messages2)
	message3, err1 := db.Exec("INSERT INTO `code_information` (`exchange_code`, `uid`, `game_coin`, `la_ba`, `operating_time`, `is_used`, `max_usage`, `valid_until_time`) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		validCodes["exchange_code"], uid, validCodes["game_coin"], validCodes["la_ba"], helper.GetCurrentShanghaiTimeString(), validCodes["is_used"], validCodes["max_usage"], validCodes["valid_until_time"])
	if err != nil {
		log.Println(err1)
	}
	log.Println(message3)
	payload := map[string]interface{}{
		"la_ba":     laba,
		"game_coin": gamecoin,
	}
	return payload
}
