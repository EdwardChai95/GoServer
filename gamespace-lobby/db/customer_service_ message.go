package db

import (
	// "gitlab.com/wolfplus/gamespace-lobby/db/model"
	"log"
	"math"
	"strconv"

	"gitlab.com/wolfplus/gamespace-lobby/helper"
)

const maxPlayersInAPage = 5
const maxMessages = 30

//GetMessageListingWithPlayerID gets the message listing
func GetMessageListingWithPlayerID(playerID string, isAdmin bool) map[string]interface{} {
	if isAdmin == true {
		var messagePlayerIDAndName []map[string]string
		userInfo, err := db.QueryString("select `nick_name` from `user` where `uid` = '" + playerID + "'")
		if err != nil {
			log.Println(err)
		}
		if len(userInfo) > 0 {
			PlayerIDAndName := make(map[string]string)
			PlayerIDAndName["player_id"] = playerID
			PlayerIDAndName["player_name"] = userInfo[0]["nick_name"]
			messagePlayerIDAndName = append(messagePlayerIDAndName, PlayerIDAndName)
			payload := map[string]interface{}{
				"messageListing": messagePlayerIDAndName,
			}
			return payload
		}
	}
	return nil
}

//GetMessagesWithPlayerID gets the latest 30 messages with the PlayerID given
func GetMessagesWithPlayerID(playerid string, isAdmin bool) map[string]interface{} {
	messages, err := db.QueryString("select * from `customer_service_message` where" +
		" `player_id` = '" + playerid + "' ORDER BY `time_sent` DESC LIMIT " + strconv.Itoa(maxMessages))
	if err != nil {
		log.Println(err)
	}
	// for j := 0; j < len(messages); j++ {
	// 	log.Println(messages[j])
	// }
	if isAdmin == false {
		messages2, err := db.Exec("Update `customer_service_message` set `is_read` = '1' " +
			"WHERE `is_read` = '0' and `player_id` = '" + playerid + "'")
		if err != nil {
			log.Println(err)
		}
		log.Println(messages2)
	} else {
		messages2, err := db.Exec("Update `customer_service_message` set `is_admin_read` = '1' WHERE `is_admin_read` = '0' and `player_id` = '" + playerid + "'")
		if err != nil {
			log.Println(err)
		}
		log.Println(messages2)
	}
	payload := map[string]interface{}{
		"messages": messages,
	}
	return payload
}

//SendMessageToPlayerID sends a message to the group with the same playerID
func SendMessageToPlayerID(playerID string, senderID string, isAdmin bool, message string) map[string]interface{} {
	var isRead = 1
	var isAdminRead = 0
	var adminReplied = 0
	if isAdmin {
		isRead = 0
		isAdminRead = 1
		adminReplied = 1
	}
	affected3, err := db.Exec("INSERT INTO `customer_service_message`(`player_id`, `sender_id`, `message`, `is_read`, `time_sent`, `admin_replied`, `is_admin_read`) VALUES (?, ?, ?, ?, ?, ?, ?)", playerID, senderID, message, isRead, helper.GetCurrentShanghaiTimeString(), adminReplied, isAdminRead)
	if err != nil {
		log.Println(err)
	}
	log.Println(affected3)
	return nil //GetMessagesWithPlayerID(playerID, isAdmin)
}

//GetUnreadMessagesNumber gets the unread messages number
func GetUnreadMessagesNumber(playerID string, isAdmin bool) int {
	searchParams := "`is_read`"
	if isAdmin == true {
		searchParams = "`is_admin_read`"
		messages2, err := db.QueryString("SELECT Count(*) FROM `customer_service_message` a WHERE `time_sent` = (SELECT MAX(`time_sent`) FROM `customer_service_message` WHERE a.`player_id` = `player_id`) and " + searchParams + " = '0'")
		if err != nil {
			log.Println(err)
		}
		maxCountinString := messages2[0]["Count(*)"]
		maxCount, err := strconv.Atoi(maxCountinString)
		return maxCount
	} else {
		adminRepliedBeforeQuery, err := db.QueryString("SELECT Count(*) FROM `customer_service_message` where `player_id` = '" + playerID + "' and `admin_replied` = '1'")
		if err != nil {
			log.Println(err)
		}
		adminRepliedBeforeTimes := adminRepliedBeforeQuery[0]["Count(*)"]
		adminRepliedBefore, err := strconv.Atoi(adminRepliedBeforeTimes)
		if err != nil {
			log.Println(err)
		}
		if adminRepliedBefore > 0 {
			messages2, err := db.QueryString("SELECT Count(*) FROM `customer_service_message` a WHERE `time_sent` = (SELECT MAX(`time_sent`) FROM `customer_service_message` WHERE a.`player_id` = `player_id`) and `player_id` ='" + playerID + "' and " + searchParams + " = '0'")
			if err != nil {
				log.Println(err)
			}
			maxCountinString := messages2[0]["Count(*)"]
			maxCount, err := strconv.Atoi(maxCountinString)
			return maxCount
		}
	}
	return 0
}

//GetMessageListingFromPageNumber gets the message listing
func GetMessageListingFromPageNumber(playerID string, pageNumber int, isAdmin bool,
	status string) map[string]interface{} {
	offsetNumber := (pageNumber - 1) * maxPlayersInAPage
	searchParams := "`is_read`"
	if isAdmin == true {
		searchParams = "`is_admin_read`"
		whereQuery := " AND b.`user_permission` != 'admin'"

		if status == "0" || status == "1" {
			whereQuery += " AND " + searchParams + " = '" + status + "'"
		}

		messageSQL := "SELECT b.nick_name as player_name, a.*, b.`user_permission` FROM `customer_service_message` a " +
			"LEFT JOIN `user` b ON b.uid = a.player_id " +
			"WHERE `time_sent` = (SELECT MAX(`time_sent`) FROM `customer_service_message` " +
			"WHERE a.`player_id` = `player_id` ORDER BY `admin_replied` LIMIT 1) " + whereQuery +
			//"GROUP BY a.`player_id` " +
			"ORDER BY `admin_replied`, `time_sent` DESC " +
			"LIMIT " + strconv.Itoa(maxPlayersInAPage) +
			" OFFSET " + strconv.Itoa(offsetNumber)

		// logger.Println(messageSQL)

		messages, err := db.QueryString(messageSQL)
		if err != nil {
			log.Println(err)
		}
		messages2, err := db.QueryString("SELECT Count(*) as countMessages, b.`user_permission` FROM `customer_service_message` a " +
			"LEFT JOIN `user` b ON b.uid = a.player_id " +
			"WHERE `time_sent` = (SELECT MAX(`time_sent`) FROM `customer_service_message` " +
			"WHERE a.`player_id` = `player_id` LIMIT 1)" + whereQuery +
			"GROUP BY a.`player_id` " +
			" LIMIT 1")
		if err != nil {
			log.Println(err)
		}
		// var messagePlayerIDAndName []map[string]string
		// for i := 0; i < len(messages); i++ {
		// 	userInfo, err := db.QueryString("select `nick_name` from `user` " +
		// 		"where `uid` = '" + messages[i]["player_id"] + "'")
		// 	if err != nil {
		// 		log.Println(err)
		// 		continue
		// 	}
		// 	PlayerIDAndName := make(map[string]string)
		// 	PlayerIDAndName["player_id"] = messages[i]["player_id"]
		// 	if len(userInfo) > 0 {
		// 		PlayerIDAndName["player_name"] = userInfo[0]["nick_name"]
		// 	} else {
		// 		PlayerIDAndName["player_name"] = "游客"
		// 	}
		// 	messagePlayerIDAndName = append(messagePlayerIDAndName, PlayerIDAndName)
		// }
		maxCountinString := messages2[0]["countMessages"]
		pageWanted, err := strconv.Atoi(maxCountinString)
		if err != nil {
			log.Println(err)
		}
		maxPageCount := math.Ceil(float64(pageWanted) / float64(maxPlayersInAPage))
		payload := map[string]interface{}{
			"count":          maxPageCount,
			"messageListing": messages, // messagePlayerIDAndName,
		}
		return payload
	} else {
		adminRepliedBeforeQuery, err := db.QueryString("SELECT Count(*) FROM `customer_service_message` where `player_id` = '" + playerID + "' and `admin_replied` = '1'")
		if err != nil {
			log.Println(err)
		}
		adminRepliedBeforeTimes := adminRepliedBeforeQuery[0]["Count(*)"]
		adminRepliedBefore, err := strconv.Atoi(adminRepliedBeforeTimes)
		if err != nil {
			log.Println(err)
		}
		if adminRepliedBefore > 0 {
			messages, err := db.QueryString("SELECT * FROM `customer_service_message` a WHERE `time_sent` = (SELECT MAX(`time_sent`) FROM `customer_service_message` WHERE a.`player_id` = `player_id`) and `player_id` ='" + playerID + "' and " + searchParams + " = '" + status + "' ORDER BY `time_sent` LIMIT " + strconv.Itoa(maxPlayersInAPage) + " OFFSET " + strconv.Itoa(offsetNumber))
			if err != nil {
				log.Println(err)
			}
			messages2, err := db.QueryString("SELECT Count(*) FROM `customer_service_message` a WHERE `time_sent` = (SELECT MAX(`time_sent`) FROM `customer_service_message` WHERE a.`player_id` = `player_id`) and `player_id` ='" + playerID + "' and " + searchParams + " = '" + status + "'")
			if err != nil {
				log.Println(err)
			}
			payload := map[string]interface{}{
				"count":          messages2,
				"messageListing": messages,
			}
			return payload
		} else {
			payload := map[string]interface{}{
				"count":          0,
				"messageListing": nil,
			}
			return payload
		}
	}
	// return nil
}
