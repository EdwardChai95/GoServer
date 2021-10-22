package db

import (
	"fmt"
	"strings"

	"gitlab.com/wolfplus/gamespace-lobby/algoutil"
	"gitlab.com/wolfplus/gamespace-lobby/db/model"
	"gitlab.com/wolfplus/gamespace-lobby/helper"
)

func CreateGameCoinLocker(ty model.AuthType, guestAcc int64, ip string) (*model.GameCoinLocker, error) {
	logger.Infof("set game locker password dbside, guestAcc:%d ", guestAcc)

	l := model.GameCoinLocker{
		Uid: guestAcc,
	}

	has, _ := db.Get(&l)

	if has {
		return &l, nil
	} else {

		//create locker
		l := model.GameCoinLocker{
			Uid:      guestAcc,
			Balance:  0,
			Password: "n",
		}

		db.Insert(&l)

		return &l, nil
	}

}

func SetPassword(ty model.AuthType, guestAcc int64, password string, ip string) (*model.GameCoinLocker, error) {
	logger.Infof("set game locker password dbside, guestAcc:%d ,password:%s ", guestAcc, password)

	//update locker
	l := model.GameCoinLocker{
		Uid: guestAcc,
	}

	db.Get(&l)

	hash, salt := algoutil.PasswordHash(password)
	l.Password = hash + helper.PASSWORDSEPERATOR + salt
	// l.Password = password

	db.ID(guestAcc).Update(&l)

	return &l, nil

}

func GetGameCoinTR(ty model.AuthType, guestAcc int64, ip string) ([]*model.GameCoinLockerHistory, error) {
	logger.Infof("get game coin transaction record dbside, guestAcc:%d ", guestAcc)

	c := make([]*model.GameCoinLockerHistory, 0)

	err := db.Where("uid = ?", guestAcc).Desc("date").Limit(10).Find(&c)

	if err != nil {
		logger.Infof("something wrong in find ")
		return nil, err
	}

	fmt.Println("what ur length inside:", len(c))

	return c, nil
}

func WithdrawGameCoin(ty model.AuthType, guestAcc int64, coinNumber int64, password string, ip string) (*model.GameCoinLocker, error) {
	logger.Infof("wallet withdraw game coin dbside, guestAcc:%d, coinNumber:%d ,password:%s", guestAcc, coinNumber, password)

	//update locker
	l := model.GameCoinLocker{
		Uid: guestAcc,
	}

	has, _ := db.Get(&l)

	if !has {
		logger.Infof("record is gone")
		return nil, nil
	} else {
		passwordSlice := strings.Split(l.Password, helper.PASSWORDSEPERATOR)
		if !algoutil.VerifyPassword(password, passwordSlice[1], passwordSlice[0]) {
			logger.Infof("password is wrong")
			return nil, nil
		}

		logger.Infof("record is here")

		l.Balance -= coinNumber

		affected, _ := db.ID(guestAcc).AllCols().Update(&l)

		if affected != 0 {
			logger.Infof("some affect ")
		} else {
			logger.Infof("0 affect")
		}

		//update user game coin
		u := model.User{
			Uid: guestAcc,
		}

		db.Get(&u)

		u.GameCoin += coinNumber

		db.ID(guestAcc).AllCols().Update(&u)

		//update history
		record := model.GameCoinLockerHistory{
			Uid:     guestAcc,
			//Operate: "提取",
			Operate: "đưa ra",
			Amount:  coinNumber,
			Date:    helper.GetCurrentShanghaiTime(),
			Balance: l.Balance,
		}

		db.Insert(&record)

		return &l, nil
	}
}

func GetGameCoinLB(ty model.AuthType, guestAcc int64, ip string) (*model.GameCoinLocker, error) {
	logger.Infof("get game coin locker balance dbside , guestAcc:%d ", guestAcc)

	l := model.GameCoinLocker{
		Uid: guestAcc,
	}

	has, _ := db.Get(&l)

	if !has {
		return nil, nil
	} else {
		return &l, nil
	}

}

func DepositGameCoin(ty model.AuthType, guestAcc int64, coinNumber int64, ip string) (*model.GameCoinLocker, error) {
	logger.Infof("wallet deposit game coin dbside, guestAcc:%d, coinNumber:%d ", guestAcc, coinNumber)

	//update user game coin
	u := model.User{
		Uid: guestAcc,
	}

	db.Get(&u)

	u.GameCoin -= coinNumber

	db.ID(guestAcc).AllCols().Update(&u)

	//update locker
	l := model.GameCoinLocker{
		Uid: guestAcc,
	}

	has, _ := db.Get(&l)

	if !has {
		l := model.GameCoinLocker{
			Uid:     guestAcc,
			Balance: coinNumber, //和游戏币一样初始化值为1,前端提取时-1
		}
		db.Insert(&l)

	} else {
		l.Balance += coinNumber
		db.ID(guestAcc).AllCols().Update(&l)
	}

	//update history
	record := model.GameCoinLockerHistory{
		Uid:     guestAcc,
		//Operate: "存入",
		Operate: "Tiền gửi",
		Amount:  coinNumber,
		Date:    helper.GetCurrentShanghaiTime(),
		Balance: l.Balance,
	}

	db.Insert(&record)

	return &l, nil
}
