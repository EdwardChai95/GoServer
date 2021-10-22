package db

import (
	// "gitlab.com/wolfplus/gamespace-lobby/db/model"
	"fmt"
	"log"
	"strconv"

	"gitlab.com/wolfplus/gamespace-lobby/db/model"
	"gitlab.com/wolfplus/gamespace-lobby/helper"
)

type TransactionType string
type UpgradeType string

const (
	CLUB_TRANS_TYPE_DONATE TransactionType = "TRANSACTION_TYPE_DONATE"
	CLUB_TRANS_TYPE_GIFT   TransactionType = "TRANSACTION_TYPE_GIFT"
	CLUB_TRANS_TYPE_FUND   TransactionType = "TRANSACTION_TYPE_FUND"

	CLUB_UPGRADE_VICE_ADMIN   UpgradeType = "UPGRADE_VICE_ADMIN"
	CLUB_DOWNGRADE_VICE_ADMIN UpgradeType = "DOWNGRADE_VICE_ADMIN"
)

func GetValidClubUserByUserId(uid string) map[string]string {
	club_user, err := db.QueryString("select * from `club_user` where `uid` = '" + uid +
		"' AND `type` != 'user0' AND `type` != 'user-1'  LIMIT 1")
	if err != nil {
		log.Println(err)
	}

	if len(club_user) > 0 {
		return club_user[0]
	}
	return nil
}

func ClubDonateToUser(club_id string, uid string, amount string) (bool, string) {
	reason := ""

	// admin, err := db.QueryString("select u.* from `club_user` c LEFT JOIN `user` u ON u.`uid` = c.`uid` where c.`club_id` = '" + club_id + "' AND c.type='admin' LIMIT 1")
	// clubs, err := db.QueryString("select * from `club` where `club_id` = '" + club_id + "' LIMIT 1")
	affected, err := db.QueryString("SELECT u.*, club.`club_name` as club_name " +
		" FROM `club_user`cu" +
		" LEFT JOIN `user` u ON cu.`uid` = u.`uid`" +
		" LEFT JOIN `club` ON club.`club_id` = cu.`club_id` " +
		" WHERE cu.`club_id` = '" + club_id + "' AND cu.`type` = 'admin' LIMIT 1")
	if len(affected) == 0 {
		//return false, "出了问题"
		return false, "lỗi"
	}

	donate_amount, err1 := strconv.Atoi(amount)
	club_game_coin, err1 := strconv.Atoi(affected[0]["game_coin"])

	if err != nil {
		log.Println(err)
	}

	if club_game_coin >= donate_amount && err1 == nil { // sufficient game_coin
		affected1, err := db.Exec("update `user` set `game_coin` = `game_coin` + ? where `uid` = ?", strconv.Itoa(donate_amount), uid)
		if err != nil {
			log.Println(err)
		}
		affected2, err := db.Exec("update `user` set `game_coin` = ? where `uid` = ?", strconv.Itoa(club_game_coin-donate_amount), affected[0]["uid"])
		if err != nil {
			log.Println(err)
		}

		affected3, err := db.Exec("INSERT INTO `game_coin_transaction`(`uid`, `club_id`, `value`, `type`, `comment`, `datetime`) VALUES (?, ?, ?, ?, ?, ?)", uid, club_id, strconv.Itoa(donate_amount), "游戏币", "gift from "+affected[0]["club_name"], helper.GetCurrentShanghaiTimeString())
		if err != nil {
			log.Println(err)
		}

		log.Println(affected1)
		log.Println(affected2)
		log.Println(affected3)
		return true, reason
	} else {
		//reason = "公会游戏币余额不足"
		reason = "Không đủ tiền"
	}
	return false, reason
}

func GetClubTransactions(club_id string, reqJson map[string]string, transactionType TransactionType) []map[string]string {
	where := "`club_id` = '" + club_id + "'"
	if reqJson["Uid"] != "" {
		where = where + " AND `uid` = '" + reqJson["Uid"] + "'"
	}
	if transactionType == (CLUB_TRANS_TYPE_GIFT) {
		where = where + " AND `value` > 0 "
	}
	if transactionType == (CLUB_TRANS_TYPE_DONATE) {
		where = where + " AND `value` < 0 "
	}
	if transactionType == (CLUB_TRANS_TYPE_FUND) {
		where = where + " AND `value` > 0 AND `collected` = 1"
	}
	transactions, err := db.QueryString("select * from `game_coin_transaction` where " + where + " ORDER BY `datetime` DESC LIMIT 10")
	if err != nil {
		log.Println(err)
	}
	return transactions
}

func DonateToClub(club_id string, uid string, amount string) (bool, string) {
	reason := ""

	users, err := db.QueryString("select * from `user` where `uid` = '" + uid + "' LIMIT 1")
	result, err := db.QueryString("SELECT u.*, club.`club_name` as club_name,  club.`game_coin` as club_game_coin" +
		" FROM `club_user`cu" +
		" LEFT JOIN `user` u ON cu.`uid` = u.`uid`" +
		" LEFT JOIN `club` ON club.`club_id` = cu.`club_id` " +
		" WHERE cu.`club_id` = '" + club_id + "' AND cu.`type` = 'admin' LIMIT 1")
	// admin, err := db.QueryString("select u.* from `club_user` c LEFT JOIN `user` u ON u.`uid` = c.`uid` where c.`club_id` = '" + club_id + "' AND c.`type`='admin' LIMIT 1")
	// clubs, err := db.QueryString("select * from `club` where `club_id` = '" + club_id + "' LIMIT 1")

	if err != nil {
		log.Println(err)
	}

	if len(users) == 0 || len(result) == 0 {
		//return false, "出了问题"
		return false, "lỗi"
	}

	game_coin, err1 := strconv.Atoi(users[0]["game_coin"])
	donate_amount, err2 := strconv.Atoi(amount)
	club_game_coin, err2 := strconv.Atoi(result[0]["club_game_coin"])
	if game_coin >= donate_amount && err1 == nil && err2 == nil { // sufficient game_coin
		// update user to new game_coin amount game_coin - donate_amount
		affected1, err := db.Exec("update `user` set `game_coin` = ? where `uid` = ?", strconv.Itoa(game_coin-donate_amount), uid)
		if err != nil {
			log.Println(err)
		}
		// update club to new game_coin club_game_coin + donate_amount
		affected2, err := db.Exec("update `club` set `game_coin` = ? where `club_id` = ?", strconv.Itoa(club_game_coin+donate_amount), club_id)
		if err != nil {
			log.Println(err)
		}
		affected3, err := db.Exec("INSERT INTO `game_coin_transaction`(`uid`, `club_id`, `value`, `type`, `comment`, `datetime`) VALUES (?, ?, ?, ?, ?, ?)", uid, club_id, strconv.Itoa(-donate_amount), "游戏币", "donate to "+result[0]["club_name"], helper.GetCurrentShanghaiTimeString())
		if err != nil {
			log.Println(err)
		}

		log.Println(affected1)
		log.Println(affected2)
		log.Println(affected3)
		return true, reason
	} else {
		//reason = "游戏币余额不足"
		reason = "Không đủ tiền"
	}

	return false, reason
}

func CanNewClub(club_name, uid string) bool {
	result, err := db.QueryString("SELECT COUNT(DISTINCT club.club_id) as num_result FROM `club`" +
		" LEFT JOIN `club_user` u ON u.`club_id` = `club`.`club_id`" +
		" WHERE LOWER(`club_name`) = LOWER('" + club_name + "') " +
		" OR (u.`uid` = '" + uid + "' AND (u.`type` = 'admin' OR u.`type` = 'user1' OR u.`type` = 'user2'))" +
		" LIMIT 1")

	if err != nil {
		logger.Println(err)
	}
	if len(result) > 0 {
		if result[0]["num_result"] == "0" {
			return true
		}
	}

	return false
}

func CreateNewClub(club_name string, uid string) bool {
	var clubId int64

	affected, err := db.Exec("INSERT INTO `club`(`club_name`, `game_coin`, `create_at`, `level`) VALUES (?, ?, ?, ?)", club_name, 0, helper.GetCurrentShanghaiTimeString(), 0)
	if err != nil {
		log.Println(err)
	}
	if club_id, err := affected.LastInsertId(); err == nil {
		clubId = club_id
	} else {
		return false
	}
	affected2, err := db.Exec("INSERT INTO `club_user`(`club_id`, `uid`, `type`, `is_mute`) VALUES (?, ?, ?, ?)", clubId, uid, "admin", 0)
	if err != nil {
		log.Println(err)
	}
	if i, err := affected.RowsAffected(); i > 0 && err == nil {
		return true
	}
	logger.Println(affected)
	logger.Println(affected2)
	return false
}

func CanJoinClub(uid string, club_id string) bool {
	// user0 = 申请
	// user1 = 精英
	// user2 = 管理员
	// user-1=被拉黑成员
	result, err := db.QueryString("SELECT COUNT(DISTINCT club_user.club_id) as num_result FROM `club_user` WHERE `uid` = '" + uid + "' AND (`type` = 'user1' OR `type` = 'user2' OR `type` = 'admin' OR (`club_id` = '" + club_id + "' AND (`type` = 'user-1' OR `type` = 'user0')))")
	if err != nil {
		log.Println(err)
	}
	if len(result) > 0 {
		if result[0]["num_result"] == "0" {
			return true
		}
	}
	return false
}

func PlayerCanJoin(uid string, club_id string) bool {
	affected, err := db.Exec("INSERT INTO `club_user`(`club_id`, `uid`, `type`, `is_mute`) VALUES (?, ?, ?, ?)", club_id, uid, "user0", 0)
	if err != nil {
		log.Println(err)
	}
	log.Println(affected)
	if i, err := affected.RowsAffected(); i > 0 && err == nil {
		return true
	}
	return false
}

// func GetJoinedClubInfo(uid string) map[string]string {
// 	result, err := db.QueryString("SELECT * FROM `club_user` cu LEFT JOIN `club` c ON cu.`club_id` = c.`club_id` WHERE cu.`uid` = '" + uid + "' LIMIT 1")
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	if len(result) > 0 {
// 		return result[0]
// 	}
// 	return nil
// }

func GetClubList(club_name string) ([]map[string]string, error) {

	// err := db.Where("club_name LIKE '%?%'", club_name).Limit(10,0).Find(&c)
	club_id, err := strconv.Atoi(club_name)
	results, err := db.QueryString("SELECT * FROM `club` WHERE `club_name` LIKE '%" + club_name + "%' OR `club_id` LIKE '%" + strconv.Itoa(club_id) + "%' LIMIT 10")
	// convert to int and back to string

	log.Println("results")
	log.Println(results)

	if err != nil {
		log.Println(err)
	}

	return results, err

}

func GetAdminByClubId(club_id string) map[string]string {
	users, err := db.QueryString("select u.* from `club_user` c LEFT JOIN `user` u ON u.`uid` = c.`uid` where c.`club_id` = '" + club_id + "' AND c.`type`='admin' LIMIT 1")
	if err != nil {
		log.Println(err)
	}

	if len(users) > 0 {
		return users[0]
	}
	return nil
}

func GetClubInfoByClubId(club_id string) (map[string]string, error) {
	clubs, err := db.QueryString("select * from `club` where `club_id` = '" + club_id + "' LIMIT 1")
	if err != nil {
		log.Println(err)
	}
	if len(clubs) > 0 {
		return clubs[0], nil
	}
	return nil, err
}

func GetNumMembers(club_id string) (string, error) {
	affected, err := db.QueryString("SELECT COUNT(*) as num_result FROM `club_user` WHERE `club_id` = '" + club_id + "' AND (`type` = 'user1' OR `type` = 'user2' OR `type` = 'admin') Limit 1")
	if err != nil {
		log.Println(err)
	}
	if len(affected) > 0 {
		total := affected[0]["num_result"]
		return total, nil
	}
	return "0", err
}

func UpdateClubContent(reqJson map[string]string) bool {
	set_query := ""
	club_id := ""
	for k, v := range reqJson {
		if k == "club_id" {
			club_id = v
			continue
		}
		set_query += fmt.Sprintf("`%v` = '%v'", k, v)
	}
	affected, err := db.Exec("update `club` set "+set_query+" where `club_id` = ?", club_id)
	if err != nil {
		log.Println(err)
	}
	if i, err := affected.RowsAffected(); i > 0 && err == nil {
		return true
	}
	return false
}

// retarded function
func GetAllClubUsers(club_id string) []map[string]string {
	results, err := db.QueryString("select u.* from `club_user` c LEFT JOIN `user` u ON u.`uid` = c.`uid` where c.`club_id` = '" + club_id + "'LIMIT 200")
	if err != nil {
		log.Println(err)
	}
	return results
}

func RetrieveClubGameCoinFromClub(club_id string, uid string, amount string) (bool, string) {
	reason := ""

	// wk: 这边为什么不用GetClubInfo?
	clubs, err := db.QueryString("select * from `club` where `club_id` = '" + club_id + "' LIMIT 1")

	if err != nil {
		log.Println(err)
	}
	if len(clubs) == 0 {
		//return false, "出了问题"
		return false, "lỗi"
	}

	club_game_coin, err1 := strconv.Atoi(clubs[0]["game_coin"])
	retrieve_amount, err2 := strconv.Atoi(amount)

	if club_game_coin >= retrieve_amount && err1 == nil && err2 == nil { // sufficient game_coin

		// update admin to new game_coin (adminGameCoin + retrieveAmount)
		affected1, err := db.Exec("update `user` set `game_coin` = `game_coin` + ? where `uid` = ?", strconv.Itoa(retrieve_amount), uid)
		if err != nil {
			log.Println(err)
		}
		// update club to new game_coin club_game_coin - retrieveAmount
		affected2, err := db.Exec("update `club` set `game_coin` = ? where `club_id` = ?", strconv.Itoa(club_game_coin-retrieve_amount), club_id)
		if err != nil {
			log.Println(err)
		}
		affected3, err := db.Exec("INSERT INTO `game_coin_transaction`(`uid`, `club_id`, `value`, `type`, `comment`, `datetime`, `collected`) VALUES (?, ?, ?, ?, ?, ?,?)", uid, club_id, strconv.Itoa(retrieve_amount), "游戏币", "retrieve from "+clubs[0]["club_name"], helper.GetCurrentShanghaiTimeString(), 1)
		if err != nil {
			log.Println(err)
		}

		log.Println(affected1)
		log.Println(affected2)
		log.Println(affected3)
		return true, reason
	} else {
		//reason = "公会游戏币余额不足"
		reason = "Không đủ tiền"
	}
	return false, reason
}

func GetClubViceAdminByClubId(club_id string) ([]map[string]string, error) {
	result, err := db.QueryString("SELECT * FROM `club_user` cu LEFT JOIN `user` u ON cu.`uid` = u.`uid` WHERE cu.`club_id` =  '" + club_id + "' AND cu.`type` = 'user2' LIMIT 3")
	if err != nil {
		log.Println(err)
	}
	if len(result) > 0 {
		return result, err
	}
	return nil, err
}

func GetNumsOfViceAdminExistInClub(club_id string) (string, error) {
	affected, err := db.QueryString("SELECT COUNT(*) as num_result FROM `club_user` cu LEFT JOIN `user` u ON cu.`uid` = u.`uid` WHERE cu.`club_id` =  '" + club_id + "' AND cu.`type` = 'user2' LIMIT 1")
	if err != nil {
		log.Println(err)
	}
	if len(affected) > 0 {
		numb := affected[0]["num_result"]
		return numb, nil
	}
	return "0", err
}

func GetUserInSameClub(uid string, club_id string) map[string]string {
	result, err := db.QueryString("SELECT u.* FROM `user` u" +
		" LEFT JOIN `club_user` cu ON cu.`uid` = u.`uid`" +
		"LEFT JOIN `club` c ON c.`club_id` = cu.`club_id`" +
		"WHERE c.`club_id` = '" + club_id + "' AND u.`uid` = '" + uid + "' LIMIT 1")
	if err != nil {
		log.Println(err)
	}
	if len(result) > 0 {
		return result[0]
	}
	return nil
}

func CheckClubUserTypeByUid(club_id string, uid string) bool {
	result, err := db.QueryString("SELECT * FROM `club_user`" +
		"LEFT JOIN `user` ON user.`uid` = club_user.`uid`" +
		"LEFT JOIN `club` ON club.`club_id` = club_user.`club_id`" +
		"WHERE club.`club_id` = '" + club_id + "' AND user.`uid` = '" + uid + "' LIMIT 1")
	if err != nil {
		log.Println(err)
	}
	if len(result) > 0 {
		if result[0]["type"] != "user2" && result[0]["type"] != "admin" {
			return true
		}
	}
	return false
}

func UpdateClubUser(club_id string, uid string, upgradeType string) bool {
	userType := "user2"
	if upgradeType == string(CLUB_DOWNGRADE_VICE_ADMIN) {
		userType = "user1"
	}
	affected, err := db.Exec("update `club_user` set `type` = '"+userType+"' where `club_id` = ? AND `uid` = ?", club_id, uid)
	if err != nil {
		log.Println(err)
	}
	if i, err := affected.RowsAffected(); i > 0 && err == nil {
		return true
	}
	return false
}

// temp function
const (
	//TRANSFERTYPE_GIFT   string = "赠送"
	TRANSFERTYPE_GIFT string = "Phát phần thưởng"
	//TRANSFERTYPE_DONATE string = "捐赠"
	TRANSFERTYPE_DONATE string = "Quyên góp"
)

func TransferGameCoin(jwt_userinfo map[string]string, receiverUid string, amountToUpdate int64) (bool, string, int64, int64) {
	// transaction_comment := "gift from " + jwt_userinfo["uid"]
	log_reason := TRANSFERTYPE_GIFT
	sender_sql := "select * from `user` where `uid` = '" + jwt_userinfo["uid"] + "'"
	receiver_sql := "select * from `user` where `uid` = '" + receiverUid + "'"

	if isAdmin, ok := jwt_userinfo["isAdmin"]; ok && isAdmin == "true" {
		log_reason = TRANSFERTYPE_DONATE
		sender_sql += "  AND `user_permission` = 'proxy_admin'" // check for admin
	} else {
		receiver_sql += "  AND `user_permission` = 'proxy_admin'" // receiver has to be admin if sender is not
	}

	sender_sql += " LIMIT 1"
	receiver_sql += " LIMIT 1"

	sender_results, err1 := db.QueryString(sender_sql)
	if err1 != nil {
		log.Printf("sender err1: %v", err1)
	}
	receiver_results, err2 := db.QueryString(receiver_sql)
	if err2 != nil {
		log.Printf("receiver err2: %v", err2)
	}

	// log.Printf("sender_results: %v", sender_results)
	// log.Printf("receiver_results: %v", receiver_results)

	if len(sender_results) > 0 && len(receiver_results) > 0 {
		sender := sender_results[0]
		receiver := receiver_results[0]

		if sender["user_permission"] == "proxy_admin" {
			if receiver["proxy"] == sender["uid"] {
				proxyadmin, err := db.QueryString("select * from `proxy` where uid = '" + sender["uid"] + "' and date(operating_time) >= curdate() LIMIT 1")
				proxyuser, err := db.QueryString("select * from `proxy_user` where uid = '" + receiver["uid"] + "' and date(operating_time) >= curdate() LIMIT 1")
				code, err := db.QueryString("select * from `exchange_code` where proxy_uid = '" + sender["uid"] + "' LIMIT 1")
				if err != nil {
					log.Println(err)
				}
				if len(proxyadmin) > 0 {
					_, err := db.Exec("`Update `proxy` set `send_num` = `send_num` + 1, `total_amount` = `total_amount` + '" +
						helper.Int64ToString(amountToUpdate) + "' where uid = '" + sender["uid"] + "'")
					if err != nil {
						log.Println(err)
					}
				} else {
					_, err := db.Exec("INSERT INTO `proxy`(`uid`, `promo_code`, `promo_num`, `active_num`, `send_num`, `receive_num`, `total_num`, `total_amount`, `service_tax`, `count_completed`, `operating_time`)"+
						"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", sender["uid"], code[0]["exchange_code"], 0, 0, 1, 0, 1, amountToUpdate, 0, 0, helper.GetCurrentShanghaiDateOnlyString())
					if err != nil {
						log.Println(err)
					}
				}

				if len(proxyuser) > 0 {
					_, err := db.Exec("`Update `proxy_user` set `receive_num` = `receive_num` + 1, `total_amount` = `total_amount` + '" +
						helper.Int64ToString(amountToUpdate) + "' where uid = '" + receiver["uid"] + "'")
					if err != nil {
						log.Println(err)
					}
				} else {
					_, err := db.Exec("INSERT INTO `proxy_user`(`uid`, `proxy_uid`, `total_win`, `total_lose`, `total_win_lose`, `total_broad`, `send_num`, `receive_num`, `total_amount`, `count_completed`, `operating_time`)"+
						"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", receiver["uid"], sender["uid"], 0, 0, 0, 0, 0, 1, amountToUpdate, 0, helper.GetCurrentShanghaiDateOnlyString())
					if err != nil {
						log.Println(err)
					}
				}
			}
		} else {
			if sender["proxy"] == receiver["uid"] {
				proxyadmin, err := db.QueryString("select * from `proxy` where uid = '" + receiver["uid"] + "' and date(operating_time) >= curdate() LIMIT 1")
				proxyuser, err := db.QueryString("select * from `proxy_user` where uid = '" + sender["uid"] + "' and date(operating_time) >= curdate() LIMIT 1")
				code, err := db.QueryString("select * from `exchange_code` where proxy_uid = '" + receiver["uid"] + "' LIMIT 1")
				if err != nil {
					log.Println(err)
				}
				fmt.Println("code:", code[0]["exchange_code"])
				if len(proxyadmin) > 0 {
					_, err := db.Exec("`Update `proxy` set `receive_num` = `receive_num` + 1, `total_amount` = `total_amount` + '" +
						helper.Int64ToString(amountToUpdate) + "' where uid = '" + receiver["uid"] + "'")
					if err != nil {
						log.Println(err)
					}
				} else {
					_, err := db.Exec("INSERT INTO `proxy`(`uid`, `promo_code`, `promo_num`, `active_num`, `send_num`, `receive_num`, `total_num`, `total_amount`, `service_tax`, `count_completed`, `operating_time`)"+
						"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", receiver["uid"], code[0]["exchange_code"], 0, 0, 0, 1, 1, amountToUpdate, 0, 0, helper.GetCurrentShanghaiDateOnlyString())
					if err != nil {
						log.Println(err)
					}
				}

				if len(proxyuser) > 0 {
					_, err := db.Exec("`Update `proxy_user` set `send_num` = `send_num` + 1, `total_amount` = `total_amount` + '" +
						helper.Int64ToString(amountToUpdate) + "' where uid = '" + sender["uid"] + "'")
					if err != nil {
						log.Println(err)
					}
				} else {
					_, err := db.Exec("INSERT INTO `proxy_user`(`uid`, `proxy_uid`, `total_win`, `total_lose`, `total_win_lose`, `total_broad`, `send_num`, `receive_num`, `total_amount`, `count_completed`, `operating_time`)"+
						"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", sender["uid"], receiver["uid"], 0, 0, 0, 0, 1, 0, amountToUpdate, 0, helper.GetCurrentShanghaiDateOnlyString())
					if err != nil {
						log.Println(err)
					}
				}
			}
		}

		// check sender got money
		if helper.StringToInt64(sender["game_coin"]) >= amountToUpdate {
			// 		user *model.User, amount int64,
			// transaction_comment, log_reason, log_otherInfo, task_key string
			senderUpdatedAmount := UpdateGameCoin(
				&model.User{Uid: helper.StringToInt64(sender["uid"]), GameCoin: helper.StringToInt64(sender["game_coin"])},
				-amountToUpdate, log_reason+" 给 "+receiver["uid"], log_reason, helper.Int64ToString(amountToUpdate), "",
				//-amountToUpdate, log_reason+" đưa cho "+receiver["uid"], log_reason, helper.Int64ToString(amountToUpdate), "",
			)
			receiverUpdatedAmount := UpdateGameCoin(
				&model.User{Uid: helper.StringToInt64(receiver["uid"]), GameCoin: helper.StringToInt64(receiver["game_coin"])},
				amountToUpdate, log_reason+" 来自 "+sender["uid"], log_reason, helper.Int64ToString(amountToUpdate), "",
				//amountToUpdate, log_reason+" Từ "+sender["uid"], log_reason, helper.Int64ToString(amountToUpdate), "",
			)

			//return true, log_reason + "成功", senderUpdatedAmount, receiverUpdatedAmount
			return true, log_reason + "Thành công", senderUpdatedAmount, receiverUpdatedAmount
		} else {
			//return false, "余额不足", -1, -1
			return false, "Không đủ tiền", -1, -1
		}
	}

	//return false, log_reason + "出了问题", -1, -1
	return false, log_reason + "Thất bại", -1, -1
}

//
// func UserPermissionAdminGiftToUser(adminUid string, userUid string, amount string) (bool, string) {
// 	reason := ""

// 	query1, err := db.QueryString("select * from `user` where `uid` = '" + adminUid + "' AND `user_permission` = 'admin' LIMIT 1")

// 	if len(query1) == 0 {
// 		return false, "出了问题"
// 	}

// 	gift_amount, _ := strconv.Atoi(amount)
// 	admin_game_coin, _ := strconv.Atoi(query1[0]["game_coin"])

// 	if err != nil {
// 		log.Println(err)
// 	}

// 	if admin_game_coin >= gift_amount && err == nil { // sufficient game_coin
// 		affected1, err := db.Exec("update `user` set `game_coin` = `game_coin` + ? where `uid` = ?", strconv.Itoa(gift_amount), userUid)
// 		if err != nil {
// 			log.Println(err)
// 		}
// 		affected2, err := db.Exec("update `user` set `game_coin` = ? where `uid` = ?", strconv.Itoa(admin_game_coin-gift_amount), adminUid)
// 		if err != nil {
// 			log.Println(err)
// 		}

// 		affected3, err := db.Exec("INSERT INTO `game_coin_transaction`(`uid`, `value`, `type`, `comment`, `datetime`) VALUES (?, ?, ?, ?, ?)", userUid, strconv.Itoa(gift_amount), "游戏币", "gift from "+query1[0]["nick_name"], helper.GetCurrentShanghaiTimeString())
// 		if err != nil {
// 			log.Println(err)
// 		}
// 		log.Println(affected1)
// 		log.Println(affected2)
// 		log.Println(affected3)
// 		return true, reason
// 	} else {
// 		reason = "公会游戏币余额不足"
// 	}
// 	return false, reason
// }

// func UserDonateToUserPermissionAdmin(userUid string, adminUid string, amount string) (bool, string) {
// 	reason := ""

// 	users, err1 := db.QueryString("select * from `user` where `uid` = '" + userUid + "' LIMIT 1")
// 	admins, err2 := db.QueryString("select * from `user` where `uid` = '" + adminUid + "' AND `user_permission` = 'admin' LIMIT 1")

// 	if err1 != nil && err2 != nil {
// 		log.Println(err1)
// 		log.Println(err2)
// 	}

// 	if len(users) == 0 || len(admins) == 0 {
// 		return false, "出了问题"
// 	}

// 	game_coin, err1 := strconv.Atoi(users[0]["game_coin"])
// 	gift_amount, _ := strconv.Atoi(amount)
// 	admin_game_coin, err2 := strconv.Atoi(admins[0]["game_coin"])
// 	if game_coin >= gift_amount && err1 == nil && err2 == nil { // sufficient game_coin
// 		affected1, err := db.Exec("update `user` set `game_coin` = ? where `uid` = ?", strconv.Itoa(game_coin-gift_amount), userUid)
// 		if err != nil {
// 			log.Println(err)
// 		}
// 		affected2, err := db.Exec("update `user` set `game_coin` = ? where `uid` = ?", strconv.Itoa(admin_game_coin+gift_amount), adminUid)
// 		if err != nil {
// 			log.Println(err)
// 		}
// 		affected3, err := db.Exec("INSERT INTO `game_coin_transaction`(`uid`, `value`, `type`, `comment`, `datetime`) VALUES (?, ?, ?, ?, ?)", userUid, strconv.Itoa(-gift_amount), "游戏币", "donate to "+admins[0]["nick_name"], helper.GetCurrentShanghaiTimeString())
// 		if err != nil {
// 			log.Println(err)
// 		}
// 		log.Println(affected1)
// 		log.Println(affected2)
// 		log.Println(affected3)
// 		return true, reason
// 	} else {
// 		reason = "游戏币余额不足"
// 	}

// 	return false, reason
// }
