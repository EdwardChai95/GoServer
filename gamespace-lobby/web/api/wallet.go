package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lonng/nex"
	"gitlab.com/wolfplus/gamespace-lobby/db"
	"gitlab.com/wolfplus/gamespace-lobby/db/model"
	"gitlab.com/wolfplus/gamespace-lobby/helper"
	proto "gitlab.com/wolfplus/gamespace-lobby/proto"
)

func MakeWalletService() http.Handler {
	router := mux.NewRouter()

	router.Handle("/wallet/createGmaeCoinLocker", nex.Handler(createGameCoinLocker)).Methods("POST")
	router.Handle("/wallet/setPassword", nex.Handler(setPassword)).Methods("POST")
	router.Handle("/wallet/getGameCoinTR", nex.Handler(getGameCoinTR)).Methods("POST")
	router.Handle("/wallet/depositGameCoin", nex.Handler(depositGameCoin)).Methods("POST")
	router.Handle("/wallet/getGameCoinLB", nex.Handler(getGameCoinLB)).Methods("POST") //LB for locker balance
	router.Handle("/wallet/withdrawGameCoin", nex.Handler(withdrawGameCoin)).Methods("POST")

	return router
}

//处理 前端 创建游戏币保险箱 POST请求
func createGameCoinLocker(r *http.Request, req *proto.CreateGameCoinLocker) (map[string]interface{}, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}
	logger.Println("creating game coin locker ", uid)

	locker, _ := db.CreateGameCoinLocker(model.AuthType(req.Type), helper.StringToInt64(uid), r.RemoteAddr)

	if locker == nil {
		payload := map[string]interface{}{
			"code": "create game coin locker failed",
		}
		return payload, nil
	}

	payload := map[string]interface{}{
		"balance":  locker.Balance,
		"uid":      locker.Uid,
		"password": locker.Password,
	}

	return payload, nil

	/*return &proto.CreateGameCoinLocker{
		Password: locker.Password,
	}, nil*/

}

//处理 前端 创建游戏币保险箱密码 POST请求
func setPassword(r *http.Request, req *proto.SetPassword) (*proto.SetPassword, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}
	logger.Println("set game locker password ", uid)

	locker, _ := db.SetPassword(model.AuthType(req.Type), helper.StringToInt64(uid), req.Password, r.RemoteAddr)

	if locker == nil {
		return &proto.SetPassword{Code: "set game locker password failed"}, nil
	}

	return &proto.SetPassword{}, nil

}

//处理 前端 提取游戏币存取记录 POST请求
func getGameCoinTR(r *http.Request, req *proto.GetGameCoinTR) (*proto.GetGameCoinTR, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}
	logger.Println("wallet get transaciton history ", uid)

	records, _ := db.GetGameCoinTR(model.AuthType(req.Type), helper.StringToInt64(uid), r.RemoteAddr)

	if records == nil {
		return &proto.GetGameCoinTR{Code: "get history failed"}, nil
	}

	fmt.Println("what ur length outside:", len(records))

	c := make([]*proto.GameCoinLockerHistory, 0)

	for _, v := range records {
		c = append(c, &proto.GameCoinLockerHistory{
			Operate: v.Operate,
			Amount:  v.Amount,
			Date:    v.Date.Unix(),
			Balance: v.Balance,
		})
	}

	res := &proto.GetGameCoinTR{
		Records: c,
	}
	return res, nil

}

//处理 前端 提取游戏币 POST请求
func withdrawGameCoin(r *http.Request, req *proto.WithdrawGameCoin) (map[string]interface{}, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}
	logger.Println("wallet withdraw game coin ", uid)

	locker, _ := db.WithdrawGameCoin(model.AuthType(req.Type), helper.StringToInt64(uid), req.CoinNumber, req.Password, r.RemoteAddr)

	/*if locker == nil {
		return &proto.WithdrawGameCoin{Code: "withdraw failed"}, nil
	}

	return &proto.WithdrawGameCoin{Balance: locker.Balance}, nil*/

	if locker == nil {
		payload := map[string]interface{}{
			"code": "locker don't exist,make a deposit first",
		}
		return payload, nil
	}

	payload := map[string]interface{}{
		"balance": locker.Balance,
	}
	user, _ := db.GetUser(helper.StringToInt64(uid))
	logInformation := map[string]string{
		"uid":       uid,
		"reason":    "保险箱帐变",
		"otherInfo": fmt.Sprintf("%v", req.CoinNumber),
		"before":    fmt.Sprintf("%v", user.GameCoin-req.CoinNumber),
		"used":      fmt.Sprintf("%v", req.CoinNumber),
		"after":     fmt.Sprintf("%v", user.GameCoin),
	}
	db.SendLogInformation(logInformation)
	return payload, nil
}

//处理 前端 提取游戏币余额 POST请求
func getGameCoinLB(r *http.Request, req *proto.GetGameCoinLB) (map[string]interface{}, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}
	logger.Println("check game coin locker balance", uid)

	locker, _ := db.GetGameCoinLB(model.AuthType(req.Type), helper.StringToInt64(uid), r.RemoteAddr)

	/*if locker == nil {
		return &proto.GetGameCoinLB{Code: "locker don't exist,make a deposit first"}, nil
	}

	return &proto.GetGameCoinLB{Balance: locker.Balance}, nil*/

	if locker == nil {
		payload := map[string]interface{}{
			"code": "locker don't exist,make a deposit first",
		}
		return payload, nil
	}

	payload := map[string]interface{}{
		"balance": locker.Balance,
	}

	return payload, nil

}

//处理 前端 存入游戏币 POST请求
func depositGameCoin(r *http.Request, req *proto.DepositGameCoin) (map[string]interface{}, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}
	logger.Println("wallet deposit game coin ", uid)

	locker, _ := db.DepositGameCoin(model.AuthType(req.Type), helper.StringToInt64(uid), req.CoinNumber, r.RemoteAddr)

	if locker == nil {
		payload := map[string]interface{}{
			"code": "locker don't exist,make a deposit first",
		}
		return payload, nil
	}

	payload := map[string]interface{}{
		"balance": locker.Balance,
	}
	user, _ := db.GetUser(helper.StringToInt64(uid))
	logInformation := map[string]string{
		"uid":       uid,
		"reason":    "保险箱帐变",
		"otherInfo": fmt.Sprintf("%v", req.CoinNumber),
		"before":    fmt.Sprintf("%v", user.GameCoin+req.CoinNumber),
		"used":      fmt.Sprintf("%v", -req.CoinNumber),
		"after":     fmt.Sprintf("%v", user.GameCoin),
	}
	db.SendLogInformation(logInformation)

	return payload, nil
	/*if locker == nil {
		return &proto.DepositGameCoin{Code: "deposit failed"}, nil
	}

	return &proto.DepositGameCoin{Balance: locker.Balance}, nil*/
}

// func UpdateTransactionSuccessTime() {

// }
