package db

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"gitlab.com/wolfplus/gamespace-lobby/define"
	"gitlab.com/wolfplus/gamespace-lobby/helper"
)

func GetOrdersByUid(uid string) []map[string]string {
	orders, err := db.QueryString("select * from `order` " +
		"where `uid` = '" + uid + "' ORDER BY `updated_datetime` DESC LIMIT 20")
	if err != nil {
		log.Println(err)
	}

	return orders
}

func GetUidForFirstDeposit(uid string) int {
	firstDeposit, err := db.QueryString("select first_deposit from `user` " +
		"where `uid` = '" + uid + "' ORDER BY `create_at` DESC LIMIT 1")
	if err != nil {
		log.Println(err)
	}
	first_deposit_str := firstDeposit[0]["first_deposit"]
	log.Println("first_deposit_str: ")
	log.Println(first_deposit_str)

	first_deposit, err := strconv.Atoi(first_deposit_str)
	log.Println("first_deposit: ")
	log.Println(first_deposit)
	return first_deposit
}

func GetUidForActive(uid string) []map[string]string {
	active, err := db.QueryString("select * from `user` " +
		"where `uid` = '" + uid + "' AND `normal_active` = 0 LIMIT 1")
	if err != nil {
		log.Println(err)
	}
	return active
}

func CheckWaitingOrder(data map[string]string) map[string]string {
	sql := "select * from `order` WHERE"

	data["order_status"] = define.ORDERSTATUSWAITING
	for col, val := range data {
		sql += fmt.Sprintf("`%v` = '%v' ", col, val)
		sql += " AND "
	}

	sql += " `callback_data` <> '' ORDER BY `order_id` DESC LIMIT 1"
	//sql += " `order_status` = '" + data["order_status"] + "' ORDER BY `order_id` DESC LIMIT 1"

	orders, err := db.QueryString(sql)
	if err != nil {
		log.Println(err)
	}

	if len(orders) > 0 {
		return orders[0]
	}

	return nil
}

func CheckFirstDeposit(uid, amount string) string {
	firstDeposit, err := db.QueryString("select order_id from `order` " +
		"where `uid` = '" + uid + "' AND payment_amount = '" + amount + "' ORDER BY `created_datetime` DESC LIMIT 1")
	if err != nil {
		log.Println(err)
	}
	if len(firstDeposit) > 0 {
		firstDeposit_str := firstDeposit[0]["order_id"]
		return firstDeposit_str
	}
	return "1"
}

func GetOrderByOrderId(orderId string) map[string]string {
	orders, err := db.QueryString("select * from `order` " +
		"where `order_id` = '" + orderId + "'  LIMIT 1")
	if err != nil {
		log.Println(err)
	}

	return orders[0]
}

func UpdateOrder(orderID string, data map[string]string) (map[string]string, int64) {
	toUpdateUserGameCoin := false
	// userCoinUpdated := false
	var updatedAmount int64 = 0 // new user game coin
	sql := "UPDATE `order` "
	columns := "SET "

	data["updated_datetime"] = helper.GetCurrentShanghaiTimeString()

	var callbackPurchaseAmount int64 = 0
	if val, ok := data["callbackPurchaseAmount"]; ok {
		callbackPurchaseAmount = helper.StringToInt64(val)
		delete(data, "callbackPurchaseAmount")
	}

	colCount := len(data)
	i := 1

	for col, val := range data {
		// if col == "amount" {
		// 	i++
		// 	continue
		// }

		columns += fmt.Sprintf("`%v` = '%v' ", col, val)

		if i != colCount {
			columns += ","
		}

		i++
	}

	sql += columns + " WHERE `order_id` = '" + orderID + "'"
	if val, ok := data["order_status"]; ok && val == define.ORDERSTATUSPAID {
		// check to prevent double update
		sql += fmt.Sprintf(" AND `order_status` = '%v'", define.ORDERSTATUSWAITING)
		toUpdateUserGameCoin = true
	}
	if callbackPurchaseAmount != 0 {
		sql += fmt.Sprintf(" AND `payment_amount` = '%v'", callbackPurchaseAmount)
	}
	// if val, ok := data["amount"]; ok {
	// 	sql += fmt.Sprintf(" AND `payment_amount` = '%v'", val)
	// }
	affected1, err := db.Exec(sql)
	if err != nil {
		logger.Print(err)
	}
	if count, err := affected1.RowsAffected(); count > 0 && err == nil {
		order := GetOrderByOrderId(orderID)
		if toUpdateUserGameCoin {
			// affected1, err := db.Exec("update `user` set `game_coin` = `game_coin` + ? "+
			// 	"where `uid` = ?", order["game_coin_amount"], order["uid"])
			// if err != nil {
			// 	logger.Printf("err: %v", err)
			// }
			// logger.Printf("update game coin query: %v", affected1)

			// affected2, err := db.Exec("INSERT INTO `game_coin_transaction` "+
			// 	"(`uid`, `value`, `type`, `comment`, `datetime`) VALUES (?, ?, ?, ?, ?)",
			// 	order["uid"], order["game_coin_amount"], "游戏币",
			// 	"购买礼包，订单号："+orderID, helper.GetCurrentShanghaiTimeString())

			// if err != nil {
			// 	logger.Printf("err: %v", err)
			// }
			// logger.Printf("new game coin transaction: %v", affected2)

			user, err1 := GetUser(helper.StringToInt64(order["uid"]))
			if err1 != nil {
				log.Println(err1)
			}

			firstDeposit := GetUidForFirstDeposit(order["uid"])
			fmt.Println("uid2:", order["uid"])
			fmt.Println("purchase:", callbackPurchaseAmount)
			if firstDeposit == 0 && callbackPurchaseAmount == 5000 {
				db.Exec("UPDATE `user` SET `first_deposit` = 1 WHERE `uid` = '" + order["uid"] + "'")
			}

			active := GetUidForActive(order["uid"])
			if len(active) > 0 {
				db.Exec("UPDATE `user` set `normal_active` = 1 WHERE `uid` = '" + order["uid"] + "'")
			}

			amountToUpdate := helper.StringToInt64(order["game_coin_amount"])
			if firstDeposit == 1 && callbackPurchaseAmount == 5000 {
				amountToUpdate = 5000
			}
			updatedAmount = UpdateGameCoin(user, amountToUpdate,
				"购买礼包，订单号："+orderID, "充值", fmt.Sprintf("%v", amountToUpdate),
				"")
			// userCoinUpdated = true
		}
		return order, updatedAmount
	}
	return nil, updatedAmount
}

func NewOrder(data map[string]string) (int64, error) {
	// check if there is existing vc order

	sql := "INSERT INTO `order`"
	columns := "("
	values := "("

	data["created_datetime"] = helper.GetCurrentShanghaiTimeString()
	data["updated_datetime"] = data["created_datetime"]
	data["order_status"] = define.ORDERSTATUSWAITING

	colCount := len(data)
	i := 1

	for col, val := range data {
		columns += fmt.Sprintf("`%v`", col)
		values += fmt.Sprintf("'%v'", val)

		if i != colCount {
			columns += ","
			values += ","
		}

		i++
	}

	columns += ")"
	values += ")"

	sql += columns + " VALUES " + values

	affected, err := db.Exec(sql)
	if err != nil {
		logger.Println(err)
		return -1, err
	}
	if orderId, err := affected.LastInsertId(); err == nil {
		return orderId, nil
	}
	return -1, errors.New("付款失败")
}
