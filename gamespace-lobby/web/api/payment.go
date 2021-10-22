package api

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	b64 "encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
	"unsafe"

	"github.com/gorilla/mux"
	"github.com/lonng/nex"
	"gitlab.com/wolfplus/gamespace-lobby/db"
	"gitlab.com/wolfplus/gamespace-lobby/define"
	"gitlab.com/wolfplus/gamespace-lobby/helper"
)

// return url
// /payment/updateuser

// 打款功能
// https://payment.funpay.asia/fun/payment/offlinePay

const (
	API_VCDESTROY    string = "https://payment.funpay.asia/fun/payment/virtualCard/destroy"
	API_VCCREATE     string = "https://payment.funpay.asia/fun/payment/virtualCard/create"
	API_VCOFFLINEPAY string = "https://payment.funpay.asia/fun/payment/offlinePay"
	//secretkey        string = "2u1u62zwm8UV911J6W1yUI6o1K277p6okk6c8pFo"
	//merchantID       int16  = 1586
	//businessID       int16  = 273
	//feeID            int16  = 243
	//feeName          string = "testing1"
	secretkey  string = "aHc7j6vY5aZbRe088mUT69X90Ac9GUGv3OoZz4T0"
	merchantID int16  = 1899
	businessID int16  = 960
	feeID      int16  = 930
	feeName    string = "test" // 计费点名称
	currency   string = "VND"
	// vc 10 mins

	// ivnpay const
	IVNPAY_API_PAYPORTAL string = "/comm/v1/pay_portal?param="
	IVNPAY_ORDERTYPE            = "ivnpay"
)

var (
	giftPacks = []*giftPack{
		&giftPack{GameCoinAmount: 23000, PurchaseAmount: 20000},
		&giftPack{GameCoinAmount: 57500, PurchaseAmount: 50000},
		&giftPack{GameCoinAmount: 115000, PurchaseAmount: 100000},
		&giftPack{GameCoinAmount: 575000, PurchaseAmount: 500000},
		&giftPack{GameCoinAmount: 1150000, PurchaseAmount: 1000000},
		&giftPack{GameCoinAmount: 2300000, PurchaseAmount: 2000000},
	}
)

var (
	giftPacksOnce = []*giftPack{
		&giftPack{GameCoinAmount: 100000, PurchaseAmount: 5000},
	}
)

type giftPack struct {
	PurchaseAmount int64 `json:"purchaseAmount"`
	GameCoinAmount int64 `json:"gameCoinAmount"`
}

type response1 struct { // what the api returns
	Code   int         `json:"code"`
	Msg    string      `json:"msg"`
	Result interface{} `json:"result"`
}

/*ivnpay funcs*/

// checkout page
func updateIvnpayOrder2Handler(r *http.Request) (map[string]string, error) {
	r.Header.Set("Content-Type", "application/json;charset=UTF-8")

	// 	GET params were 2: map[param:[YW1vdW50PTEwMDAwMCZhdHRhY2g9JmVycm9yX2Rlc2NyPSZoYXNoPTUxMUJBMzA3NjkxMjYxRTBCOTE4Q0E3NUQzOTExMTI3Jm1jaF9pZD0xMDIyNDAmbWNoX29yZGVyX2lkPTIxMyZwYXlfdHlwZT00JnN0YXR1cz0xJnN2cl90cmFuc2FjdGlvbl9pZD1WMjIwMjEwNjAyMDExMDIyNDA0Mjg5NDExMjYxNDIxMjg5NzI4JnRzPTIwMjEtMDYtMDIrMTAlM0E1NyUzQTE3LjEwNzE5MjcxMyslMkIwMDAwK1VUQw==]]
	// time="2021-06-02T06:57:17-04:00" level=info msg="return param 2: YW1vdW50PTEwMDAwMCZhdHRhY2g9JmVycm9yX2Rlc2NyPSZoYXNoPTUxMUJBMzA3NjkxMjYxRTBCOTE4Q0E3NUQzOTExMTI3Jm1jaF9pZD0xMDIyNDAmbWNoX29yZGVyX2lkPTIxMyZwYXlfdHlwZT00JnN0YXR1cz0xJnN2cl90cmFuc2FjdGlvbl9pZD1WMjIwMjEwNjAyMDExMDIyNDA0Mjg5NDExMjYxNDIxMjg5NzI4JnRzPTIwMjEtMDYtMDIrMTAlM0E1NyUzQTE3LjEwNzE5MjcxMyslMkIwMDAwK1VUQw==" component=http service=login
	// "amount=100000&attach=&error_descr=&hash=511BA307691261E0B918CA75D3911127&mch_id=102240&mch_order_id=213&pay_type=4&status=1&svr_transaction_id=V220210602011022404289411261421289728&ts=2021-06-02+10%3A57%3A17.107192713+%2B0000+UTC"

	fmt.Println("GET params were 2:", r.URL.Query())

	param := r.URL.Query().Get("param")
	logger.Printf("return param 2: %v", param)
	if param == "" {
		return nil, errors.New("404")
	}

	// 将param结算base64 url decode得到字符串string1
	data, err := b64.StdEncoding.DecodeString(param)
	if err != nil {
		logger.Fatal("error:", err)
	}
	fmt.Printf("%q\n", data)
	// 切分出所有参数，提出hash后，剩下参数集合做签名校验（参考签名规则）

	decodedValue, err := url.ParseQuery(string(data))
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(decodedValue)

	if !_verifyIvnpayHash(decodedValue) {
		return nil, errors.New("404")
	}

	orderID := decodedValue["mch_order_id"][0]

	error1 := define.ORDERSTATUSFAILED
	if val, ok := decodedValue["error_no"]; ok {
		error1 = val[0]
	}

	if decodedValue["status"][0] == "1" { // success
		// get and update order
		order := db.GetOrderByOrderId(orderID)
		if order != nil {
			// 订单号，支付时间，支付金额，游戏币数额，备注
			return map[string]string{
				"标题":    define.ORDERSTATUSPAID,
				"订单号":   order["order_id"],
				"支付时间":  order["updated_datetime"],
				"支付金额":  order["payment_amount"],
				"游戏币数额": order["game_coin_amount"],
			}, nil
		} else {
			return nil, errors.New(fmt.Sprintf("结果码：%v: %v", decodedValue["status"][0], error1))
		}
	} else {
		return nil, errors.New(fmt.Sprintf("结果码：%v: %v", decodedValue["status"][0], error1))
	}
}

func updateIvnpayOrderHandler(r *http.Request) (map[string]interface{}, error) {
	r.Header.Set("Content-Type", "application/json;charset=UTF-8")

	// GET params were: map[param:[YW1vdW50PTEwMDAwMCZhdHRhY2g9Jmhhc2g9QUI1NDhBMjQ3NkY1NzEyMkQzODRDRTAyNTM0MUU4NDUmbWNoX2lkPTEwMjI0MCZtY2hfb3JkZXJfaWQ9MjEzJnBheV90eXBlPTQmc3RhdHVzPTEmc3ZyX3RyYW5zYWN0aW9uX2lkPVYyMjAyMTA2MDIwMTEwMjI0MDQyODk0MTEyNjE0MjEyODk3MjgmdHM9MTYyMjYzMTQyOQ==]]
	// time="2021-06-02T06:57:09-04:00" level=info msg="return param: YW1vdW50PTEwMDAwMCZhdHRhY2g9Jmhhc2g9QUI1NDhBMjQ3NkY1NzEyMkQzODRDRTAyNTM0MUU4NDUmbWNoX2lkPTEwMjI0MCZtY2hfb3JkZXJfaWQ9MjEzJnBheV90eXBlPTQmc3RhdHVzPTEmc3ZyX3RyYW5zYWN0aW9uX2lkPVYyMjAyMTA2MDIwMTEwMjI0MDQyODk0MTEyNjE0MjEyODk3MjgmdHM9MTYyMjYzMTQyOQ==" component=http service=login

	fmt.Println("GET params were:", r.URL.Query())

	param := r.URL.Query().Get("param")
	logger.Printf("return param: %v", param)

	// 将param结算base64 url decode得到字符串string1
	data, err := b64.StdEncoding.DecodeString(param)
	if err != nil {
		logger.Fatal("error:", err)
	}
	fmt.Printf("%q\n", data)
	// "amount=100000&attach=&hash=AB548A2476F57122D384CE025341E845&mch_id=102240&mch_order_id=213&pay_type=4&status=1&svr_transaction_id=V220210602011022404289411261421289728&ts=1622631429"
	// 切分出所有参数，提出hash后，剩下参数集合做签名校验（参考签名规则）
	// status 0   未支付1   支付成功2   支付失败
	decodedValue, err := url.ParseQuery(string(data))
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(decodedValue)

	// ************** 切分出所有参数，提出hash后，剩下参数集合做签名校验（参考签名规则）
	if !_verifyIvnpayHash(decodedValue) {
		return nil, nil
	}
	// ************** 切分出所有参数，提出hash后，剩下参数集合做签名校验（参考签名规则）

	orderID := decodedValue["mch_order_id"][0]

	error1 := ""
	if val, ok := decodedValue["error_no"]; ok {
		error1 = val[0]
	}

	if decodedValue["status"][0] == "1" { // success
		// get and update order
		order, updatedAmount := db.UpdateOrder(orderID,
			map[string]string{
				"order_status":           define.ORDERSTATUSPAID, // important only update to paid if status is 1
				"callbackPurchaseAmount": decodedValue["amount"][0],
				"callback_data":          "",
			})
		logger.Printf("order: %v, updatedAmount: %v", order, updatedAmount)

		firstDeposit := db.GetUidForFirstDeposit(order["uid"])
		fmt.Println(firstDeposit)
		first := 0
		if firstDeposit == 1 {
			first = 1
		}

		if order != nil {
			if updatedAmount != 0 {
				// user, _ := db.GetUser(helper.StringToInt64(order["uid"]))
				data := map[string]interface{}{
					"uid":       order["uid"],
					"game_coin": updatedAmount,
				}
				data1 := map[string]interface{}{
					"uid":    order["uid"],
					"amount": decodedValue["amount"][0],
					"first":  first,
				}
				LobbyCoinUpdate(data)
				LobbyPurchaseUpdate(data1)
			}
		} else {
			db.UpdateOrder(orderID, map[string]string{
				"comment":      fmt.Sprintf("结果码：%v: %v", decodedValue["status"][0], error1),
				"order_status": define.ORDERSTATUSFAILED,
			})
		}
	} else {
		db.UpdateOrder(orderID, map[string]string{
			"comment": fmt.Sprintf("结果码：%v: %v", decodedValue["status"][0], error1),
		})
	}

	return nil, nil
}

func _verifyIvnpayHash(decodedValue url.Values) bool {
	hash := decodedValue["hash"][0]
	delete(decodedValue, "hash")

	var keys = []string{}
	for k := range decodedValue {
		keys = append(keys, k)
	}
	sort.Sort(Alphabetic(keys)) // sort by ascii chart
	string1 := ""
	for i, k := range keys {
		string1 += k + "=" + fmt.Sprintf("%v", decodedValue[k][0])
		if i+1 != len(keys) {
			string1 += "&"
		}
	}
	string1 += define.IVNPAY_SECRET
	decodedHash := md5.Sum([]byte(string1))
	sign := strings.ToUpper(hex.EncodeToString(decodedHash[:]))
	fmt.Println(sign)
	if sign != hash {
		return false
	}
	return true
}

// params giftPackId
// returns link for the front end to display
func createIvnpayOrderHandler(r *http.Request) (map[string]string, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}
	reqJson := helper.ReadParameters(r)

	selectedGiftPack := giftPacks[helper.StringToInt(reqJson["giftPackId"])]
	amount := selectedGiftPack.PurchaseAmount

	orderID := ""

	if existingOrder := db.CheckWaitingOrder(map[string]string{
		"uid":              uid,
		"type":             IVNPAY_ORDERTYPE,
		"payment_amount":   helper.Int64ToString(amount),
		"game_coin_amount": helper.Int64ToString(selectedGiftPack.GameCoinAmount),
	}); existingOrder != nil {
		orderID = existingOrder["order_id"]
	} else {
		orderIDInt64, _ := db.NewOrder(map[string]string{
			"uid":              uid,
			"payment_amount":   helper.Int64ToString(amount),
			"game_coin_amount": helper.Int64ToString(selectedGiftPack.GameCoinAmount),
			"type":             IVNPAY_ORDERTYPE,
		})
		orderID = helper.Int64ToString(orderIDInt64)
	}

	array1 := map[string]interface{}{
		"mch_id":          define.IVNPAY_MCHID,
		"mch_uid":         uid,
		"mch_order_id":    orderID,
		"equipment_type":  2,
		"expected_amount": amount,
		"mch_url":         define.PAYMENT_RETURNURL + "/payment/update_ivnpay_order2",
		"show_types":      "4", // "41+4+36+38+40+18+42",
		"attach":          "{\"auto_redirect\":\"true\"}",
	}

	/********* START SIGNING DATA **********/
	var keys = []string{}
	for k := range array1 {
		keys = append(keys, k)
	}
	sort.Sort(Alphabetic(keys)) // sort by ascii chart
	string1 := ""
	for i, k := range keys {
		string1 += k + "=" + fmt.Sprintf("%v", array1[k])
		if i+1 != len(keys) {
			string1 += "&"
		}
	}
	string1 += define.IVNPAY_SECRET
	hash := md5.Sum([]byte(string1))
	sign := strings.ToUpper(hex.EncodeToString(hash[:]))
	/********* END SIGNING DATA **********/

	array1["hash"] = sign // 将hash值添加到集合M生成新的集合M1

	jsonString, _ := json.Marshal(array1) // 集合M1 json序列化生成新的字符串string1
	param := b64.StdEncoding.EncodeToString([]byte(jsonString))

	logger.Println(array1)
	logger.Println(jsonString)
	logger.Println(param)

	payload := map[string]string{
		"param": param,
		"link":  define.IVNPAY_URL + IVNPAY_API_PAYPORTAL + param,
	}
	return payload, nil
}

func createIvnpayOrderOnceHandler(r *http.Request) (map[string]string, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}
	//reqJson := helper.ReadParameters(r)

	selectedGiftPackOnce := giftPacksOnce[0]
	amount := selectedGiftPackOnce.PurchaseAmount
	gamecoin := selectedGiftPackOnce.GameCoinAmount
	orderID := ""

	if existingOrder := db.CheckWaitingOrder(map[string]string{
		"uid":              uid,
		"type":             IVNPAY_ORDERTYPE,
		"payment_amount":   helper.Int64ToString(amount),
		"game_coin_amount": helper.Int64ToString(gamecoin),
	}); existingOrder != nil {
		orderID = existingOrder["order_id"]
	} else {
		/*if firstFail := db.CheckFirstDeposit(uid, helper.Int64ToString(amount)); firstFail != "1" {
			updateOrderData := map[string]string{}
			updateOrderData["order_status"] = define.ORDERSTATUSFAILED
			updateOrderData["comment"] = fmt.Sprintf("结果码：首充失效")
			db.UpdateOrder(firstFail, updateOrderData)
		}*/
		orderIDInt64, _ := db.NewOrder(map[string]string{
			"uid":              uid,
			"payment_amount":   helper.Int64ToString(amount),
			"game_coin_amount": helper.Int64ToString(gamecoin),
			"type":             IVNPAY_ORDERTYPE,
		})
		orderID = helper.Int64ToString(orderIDInt64)
	}

	array1 := map[string]interface{}{
		"mch_id":          define.IVNPAY_MCHID,
		"mch_uid":         uid,
		"mch_order_id":    orderID,
		"equipment_type":  2,
		"expected_amount": amount,
		"mch_url":         define.PAYMENT_RETURNURL + "/payment/update_ivnpay_order2",
		"show_types":      "4", // "41+4+36+38+40+18+42",
		"attach":          "{\"auto_redirect\":\"true\"}",
	}

	/********* START SIGNING DATA **********/
	var keys = []string{}
	for k := range array1 {
		keys = append(keys, k)
	}
	sort.Sort(Alphabetic(keys)) // sort by ascii chart
	string1 := ""
	for i, k := range keys {
		string1 += k + "=" + fmt.Sprintf("%v", array1[k])
		if i+1 != len(keys) {
			string1 += "&"
		}
	}
	string1 += define.IVNPAY_SECRET
	hash := md5.Sum([]byte(string1))
	sign := strings.ToUpper(hex.EncodeToString(hash[:]))
	/********* END SIGNING DATA **********/

	array1["hash"] = sign // 将hash值添加到集合M生成新的集合M1

	jsonString, _ := json.Marshal(array1) // 集合M1 json序列化生成新的字符串string1
	param := b64.StdEncoding.EncodeToString([]byte(jsonString))

	logger.Println(array1)
	logger.Println(jsonString)
	logger.Println(param)

	payload := map[string]string{
		"param": param,
		"link":  define.IVNPAY_URL + IVNPAY_API_PAYPORTAL + param,
	}
	return payload, nil
}
func _destroy_vc(orderNo string) { // orderNo is order_id in order table
	array1 := map[string]interface{}{
		"merchantID": merchantID,
		"businessID": businessID,
		"feeID":      feeID,
		"timestamp":  time.Now().Unix(),
		"orderNo":    orderNo,
		// "bankType":nil,
		"version": 1.3,
	}
	_, err := postData(array1,
		API_VCDESTROY)

	if err != nil {
		logger.Printf("destroy err: %v", err)
	}
}

func getOrdersByUidHandler(r *http.Request) ([]map[string]string, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}

	orders := db.GetOrdersByUid(uid)

	if len(orders) > 0 {
		return orders, nil
	}

	return nil, nil
}

func GetGiftPacksHandler(r *http.Request) ([]*giftPack, error) {
	// logger.Printf("giftPacks: %v", giftPacks)
	return giftPacks, nil
}

func GetFloatPointHandler(r *http.Request) (map[string]interface{}, error) {
	logger.Println("GetFloatPointHandler")
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}

	payload := map[string]interface{}{
		"firstDeposit": db.GetUidForFirstDeposit(uid),
	}
	return payload, nil
}

// 商家传入的returnUrl POST
func updateOrderHandler(r *http.Request) (map[string]interface{}, error) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}
	var reqJson response1
	err = json.Unmarshal(reqBody, &reqJson)
	if err != nil {
		log.Println(err)
	}

	// logger.Print("updateOrderHandler")
	// logger.Printf("reqJson: %v", reqJson)

	if resultObj, ok := reqJson.Result.(map[string]interface{}); ok {
		orderID := resultObj["orderNo"].(string)
		if reqJson.Code == 10000 { // success
			// get and update order
			order, updatedAmount := db.UpdateOrder(orderID,
				map[string]string{
					"order_status":           define.ORDERSTATUSPAID, // important only update to paid if code is 10000
					"callbackPurchaseAmount": helper.Int64ToString(int64(resultObj["purchaseAmount"].(float64))),
					"callback_data":          "",
				})
			logger.Printf("order: %v, updatedAmount: %v", order, updatedAmount)
			if order != nil {
				if updatedAmount != 0 {
					// user, _ := db.GetUser(helper.StringToInt64(order["uid"]))
					data := map[string]interface{}{
						"uid":       order["uid"],
						"game_coin": updatedAmount,
					}
					data1 := map[string]interface{}{
						"uid":    order["uid"],
						"amount": helper.Int64ToString(int64(resultObj["purchaseAmount"].(float64))),
					}
					LobbyCoinUpdate(data)
					LobbyPurchaseUpdate(data1)
				}

			} else {
				db.UpdateOrder(orderID, map[string]string{
					"comment": fmt.Sprintf("结果码：%v: %v，付款数额：%v", reqJson.Code, reqJson.Msg,
						int64(resultObj["purchaseAmount"].(float64))),
					"order_status": define.ORDERSTATUSFAILED,
				})
			}

			_, vcok := r.URL.Query()["vc"]
			if vcok {
				_destroy_vc(resultObj["orderNo"].(string))
			}
		} else {
			db.UpdateOrder(orderID, map[string]string{
				"comment": fmt.Sprintf("结果码：%v: %v", reqJson.Code, reqJson.Msg),
			})
		}
	}
	// 回调内容：{"code":10000,"msg":"the request is succeed.","result":{"accountName":"VAP001 ASDF","accountNo":"902000228252","amount":50000,"bankLink":"","bankName":"WOORIBANK","branchBankName":"Hoan kiem","businessID":273,"cityName":"HCM","currency":"VND","expireDate":"","feeID":243,"merchantID":1586,
	// "orderNo":"1","purchaseAmount":50000,"purchaseCurrency":"VND","purchaseTime":"20210414152122","remark":"test","serviceFee":0,"sign":"61154CDF05462AE9B6DD0ADD981D27B9","tradeNo":"003000130020210414162122582971","userName":"asdf"}}
	// orderNo, tradeNo
	// accountNo, accountName, bankName, branchBankName, bankLink 空, cityName, serviceFee,
	// sign, remark
	return nil, nil
}

// params: giftPackId, userName, phoneNumber
func vcPayHandler(r *http.Request) (map[string]interface{}, error) {

	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}
	reqJson := helper.ReadParameters(r)

	selectedGiftPack := giftPacks[helper.StringToInt(reqJson["giftPackId"])]
	amount := selectedGiftPack.PurchaseAmount

	if existingOrder := db.CheckWaitingOrder(map[string]string{
		"uid":  uid,
		"type": "vc",
		//"payment_amount":   helper.Int64ToString(amount),
		//"game_coin_amount": helper.Int64ToString(selectedGiftPack.GameCoinAmount),
	}); existingOrder != nil {

		jsonObj := response1{}
		json.Unmarshal([]byte(existingOrder["callback_data"]), &jsonObj)

		logger.Printf("created_datetime: %v", existingOrder["created_datetime"])

		layout := "2006-01-02 15:04:05"
		t, _ := time.Parse(layout, existingOrder["created_datetime"]) // order created date time
		// logger.Printf("t: %v", t)
		// logger.Printf("current shanghai time: %v", helper.GetCurrentShanghaiTime())
		t2, _ := time.Parse(layout, helper.GetCurrentShanghaiTimeString()) // current date time
		// logger.Printf("t2: %v", t2)

		diff := t2.UTC().Sub(t.UTC())

		if diff.Minutes() <= 30 { //更新为30分钟失效
			return map[string]interface{}{
				"jsonObj": jsonObj,
			}, nil
		} else {
			updateOrderData := map[string]string{
				"order_status": define.ORDERSTATUSFAILED,
			}
			db.UpdateOrder(existingOrder["order_id"], updateOrderData)
			_destroy_vc(existingOrder["order_id"])
		}

	}

	orderID, _ := db.NewOrder(map[string]string{
		"uid":              uid,
		"payment_amount":   helper.Int64ToString(amount),
		"game_coin_amount": helper.Int64ToString(selectedGiftPack.GameCoinAmount),
		"type":             "vc",
	})

	array1 := map[string]interface{}{
		"merchantID": merchantID,
		"businessID": businessID,
		"feeID":      feeID,
		// "clientID":     reqJson["clientID"], //get clientID 终端用户的ID或者合同号等唯一标记
		"timestamp":  time.Now().Unix(),
		"version":    1.3,
		"amount":     amount,
		"currency":   currency,
		"orderNo":    helper.Int64ToString(orderID),
		"expireDate": helper.GetExpiryDateHanoiTime(), //"20221230120000", // 以河内时间为准。此值暂时无实际意义
		"returnUrl":  define.PAYMENT_RETURNURL + "/payment/updateorder?uid=" + uid + "&vc=true",
		// "bankType":   "", // 此值现在已无意义，Funpay会自动选择当前最好的虚拟卡类型返回给用户。
		"accountBase": uid,
		"userName":    "BBH",        // 用户真实姓名，用于线下便利店跟用户确认核验订单，避免误操作
		"phoneNumber": "0123456789", // 用户手机号码，用于线下便利店跟用户确认核验订单，避免误操作
		// "IDNo":         reqJson["IDNo"],        // 用户真实身份证号，支持9位和12位身份证号[新增参数]
	}

	jsonObj, err := postData(array1,
		API_VCCREATE)

	payload := map[string]interface{}{
		"jsonObj": jsonObj,
	}
	logger.Printf("jsonObj: %v", jsonObj)
	if err == nil {

		updateOrderData := map[string]string{}
		if jsonObj.Code != 10000 {
			updateOrderData["order_status"] = define.ORDERSTATUSFAILED
			updateOrderData["comment"] = fmt.Sprintf("结果码：%v: %v", jsonObj.Code, jsonObj.Msg)
			if jsonObj.Code == 60005 {

			}
		} else {
			bytes, err := json.Marshal(jsonObj)
			if err != nil {
				logger.Printf("error converting to string: %v", err)
			}
			updateOrderData["callback_data"] = string(bytes)
		}
		db.UpdateOrder(helper.Int64ToString(orderID), updateOrderData)

		// jsonObj->result is an interface
		if resultObj, ok := jsonObj.Result.(map[string]interface{}); ok {
			// 商家需要将此信息展示给用户: accountNo, accountName, bankName, branchBankName
			// amount 注意此值不是用户实际存入金额
			// accountNo 为该用户分配的 VC号码,
			// accountName 为该用户分配的VC号码户主名称,
			// bankName 该VC归属银行名称,
			// branchBankName 该VC归属银行支行名称
			// bankLink 空, cityName, serviceFee, sign
			logger.Printf("result: %v", resultObj)
		}
	} else {
		logger.Printf("error: %v", err)
	}
	payload["err"] = err
	return payload, err

}

// params: phoneNumber, userName, IDNo, giftPackId
func offlinePayHandler(r *http.Request) (map[string]interface{}, error) {
	logger.Printf("offline pay expire time: %v", helper.GetExpiryDateHanoiTime())

	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}

	reqJson := helper.ReadParameters(r)
	logger.Printf("reqJson: %v", reqJson)

	selectedGiftPack := giftPacks[helper.StringToInt(reqJson["giftPackId"])]
	amount := selectedGiftPack.PurchaseAmount

	orderID, _ := db.NewOrder(map[string]string{
		"uid":              uid,
		"payment_amount":   helper.Int64ToString(amount),
		"game_coin_amount": helper.Int64ToString(selectedGiftPack.GameCoinAmount),
		"type":             "offline",
	})

	array1 := map[string]interface{}{
		"merchantID": merchantID,
		"businessID": businessID,
		"feeID":      feeID,
		// "clientID":     reqJson["clientID"], //get clientID 终端用户的ID或者合同号等唯一标记
		"timestamp":    time.Now().Unix(),
		"amount":       amount,
		"currency":     currency,
		"name":         feeName,                         // 计费点名称，必须与申请计费点的时候输入的一致
		"orderNo":      helper.Int64ToString(orderID),   // for testing
		"expireDate":   helper.GetExpiryDateHanoiTime(), //"20221230120000",// yyyyMMddHHmmss。超过指定时间后用户将不能再对此订单进行支付。
		"returnUrl":    define.PAYMENT_RETURNURL + "/payment/updateorder?uid=" + uid,
		"version":      1.3,                    // 此处是固定值“1.3”，代表API版本V1.3
		"purchaseType": 2,                      // 选择使用的支付方式，此处使用固定值2
		"phoneNumber":  reqJson["phoneNumber"], // 用户手机号码，用于线下便利店跟用户确认核验订单，避免误操作
		"userName":     reqJson["userName"],    // 用户真实姓名，用于线下便利店跟用户确认核验订单，避免误操作
		"IDNo":         reqJson["IDNo"],        // 用户真实身份证号，支持9位和12位身份证号[新增参数]
	}

	if val, ok := reqJson["clientID"]; ok && reqJson["clientID"] != "" {
		array1["clientID"] = val
	}

	jsonObj, err := postData(array1, API_VCOFFLINEPAY)

	payload := map[string]interface{}{
		"jsonObj": jsonObj,
	}
	if err == nil {
		updateOrderData := map[string]string{}
		if jsonObj.Code != 10000 {
			updateOrderData = map[string]string{
				"order_status": define.ORDERSTATUSFAILED,
				"comment":      fmt.Sprintf("结果码：%v: %v", jsonObj.Code, jsonObj.Msg),
			}
		} else {
			bytes, err := json.Marshal(jsonObj)
			if err != nil {
				logger.Printf("error converting to string: %v", err)
			}
			updateOrderData["callback_data"] = string(bytes)
		}
		db.UpdateOrder(helper.Int64ToString(orderID), updateOrderData)

		// jsonObj->result is an interface
		if resultObj, ok := jsonObj.Result.(map[string]interface{}); ok {
			logger.Printf("result: %v", resultObj)
			return payload, nil
		}
	}
	payload["err"] = err
	logger.Printf("error: %v", err)
	return payload, err
}

// no use
func onlinePayHandler(r *http.Request) (map[string]interface{}, error) {
	// uid, isValid := helper.VerifyJWT(r)
	// if !isValid {
	// 	return nil, errors.New("Invalid token")
	// }
	logger.Print("online pay start")
	uid := "111"

	// reqJson := helper.ReadParameters(r)

	array1 := map[string]interface{}{
		"merchantID": 1586,
		"businessID": 273,
		"feeID":      243,
		"timestamp":  time.Now().Unix(),
		"amount":     50000, // for testing
		"currency":   "VND",
		"name":       "testing1", // 计费点名称，必须与申请计费点的时候输入的一致
		"orderNo":    "1",        // for testing
		"returnUrl":  define.PAYMENT_RETURNURL + "/payment/updateorder?uid=" + uid,
		"version":    1.3, // 此处是固定值“1.3”，代表API版本V1.3
	}

	var keys = []string{}
	for k := range array1 {
		keys = append(keys, k)
	}
	sort.Sort(Alphabetic(keys))

	string1 := ""
	for i, k := range keys {
		string1 += k + "=" + fmt.Sprintf("%v", array1[k])
		if i+1 != len(keys) {
			string1 += "&"
		}
	}
	string3 := string1
	string1 += secretkey

	logger.Printf("array2: %v", string1)

	hash := md5.Sum([]byte(string1))
	sign := strings.ToUpper(hex.EncodeToString(hash[:]))
	array1["sign"] = sign

	string3 += "&sign=" + sign
	param := b64.StdEncoding.EncodeToString([]byte(string3))

	logger.Printf("array1: %v", array1)
	logger.Printf("string3: %v", string3)
	logger.Printf("param: %v", param)

	bytesData, err := json.Marshal(array1)
	if err != nil {
		logger.Println(err.Error())
		return nil, err
	}

	// logger.Printf("bytesData: %v", bytesData)
	reader := bytes.NewBuffer(bytesData)
	request, err := http.NewRequest("POST",
		"https://sandbox.funpay.asia/fun/payment/onlinePay",
		reader)

	if err != nil {
		logger.Println(err.Error())
		return nil, err
	}

	q := request.URL.Query()
	q.Add("param", param)
	request.URL.RawQuery = q.Encode()
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{Transport: tr}
	resp, err := client.Do(request)

	if err != nil {
		logger.Println(err.Error())
		return nil, err
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Println(err.Error())
		return nil, err
	}

	str := (*string)(unsafe.Pointer(&respBytes))
	logger.Printf("res: %v", *str)

	payload := map[string]interface{}{
		"ret": str,
	}

	return payload, nil
}

// helpers

func postData(array1 map[string]interface{}, apiString string) (response1, error) {
	signFunPayData(array1)
	logger.Printf("array1: %v", array1)

	jsonObj := response1{}

	bytesData, err := json.Marshal(array1)
	if err != nil {
		logger.Println(err.Error())
		return jsonObj, err
	}

	// logger.Printf("bytesData: %v", bytesData)
	reader := bytes.NewBuffer(bytesData)
	request, err := http.NewRequest("POST",
		apiString,
		reader)

	if err != nil {
		logger.Println(err.Error())
		return jsonObj, err
	}

	request.Header.Set("Content-Type", "application/json;charset=utf-8")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{Transport: tr}
	resp, err := client.Do(request)

	if err != nil {
		logger.Println(err.Error())
		return jsonObj, err
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Println(err.Error())
		return jsonObj, err
	}

	str := (*string)(unsafe.Pointer(&respBytes))
	logger.Printf("res: %v", *str)

	json.Unmarshal(respBytes, &jsonObj)

	return jsonObj, nil
}

func signFunPayData(array1 map[string]interface{}) {
	var keys = []string{}
	for k := range array1 {
		keys = append(keys, k)
	}
	sort.Sort(Alphabetic(keys)) // sort by ascii chart

	string1 := ""
	for i, k := range keys {
		string1 += k + "=" + fmt.Sprintf("%v", array1[k])
		if i+1 != len(keys) {
			string1 += "&"
		}
	}
	string1 += secretkey

	logger.Printf("string1: %v", string1)

	hash := md5.Sum([]byte(string1))
	sign := strings.ToUpper(hex.EncodeToString(hash[:]))
	array1["sign"] = sign

}

type Alphabetic []string

func (list Alphabetic) Len() int { return len(list) }

func (list Alphabetic) Swap(i, j int) { list[i], list[j] = list[j], list[i] }

func (list Alphabetic) Less(i, j int) bool {
	var si string = list[i]
	var sj string = list[j]
	var si_lower = []rune(si)
	var sj_lower = []rune(sj)
	// var si_lower = strings.ToLower(si)
	// var sj_lower = strings.ToLower(sj)
	if si_lower[0] == sj_lower[0] {
		return si < sj
	}
	return si_lower[0] < sj_lower[0]
}

func MakePaymentService() http.Handler {
	router := mux.NewRouter()

	// http://199.115.228.247:12307/payment/update_ivnpay_order
	router.Handle("/payment/update_ivnpay_order2", nex.Handler(updateIvnpayOrder2Handler)).Methods("GET")
	router.Handle("/payment/update_ivnpay_order", nex.Handler(updateIvnpayOrderHandler)).Methods("GET")
	router.Handle("/payment/create_ivnpay_order", nex.Handler(createIvnpayOrderHandler)).Methods("POST")
	router.Handle("/payment/create_ivnpay_orderOnce", nex.Handler(createIvnpayOrderOnceHandler)).Methods("POST")
	router.Handle("/payment/getMyOrders", nex.Handler(getOrdersByUidHandler)).Methods("GET")
	router.Handle("/payment/getGiftPacks", nex.Handler(GetGiftPacksHandler)).Methods("GET")
	router.Handle("/payment/updateorder", nex.Handler(updateOrderHandler)).Methods("POST")
	router.Handle("/payment/vcPay", nex.Handler(vcPayHandler)).Methods("POST")
	router.Handle("/payment/offlinePay", nex.Handler(offlinePayHandler)).Methods("POST")
	router.Handle("/payment/onlinePay", nex.Handler(onlinePayHandler)).Methods("POST") // not used
	router.Handle("/payment/firstDeposit", nex.Handler(GetFloatPointHandler)).Methods("POST")
	// fs := http.FileServer(http.Dir("./static/views/offlinePay/"))
	// router.PathPrefix("/payment/").Handler(http.StripPrefix("/payment/", fs)) // for web test only

	return router
}
