package db

import (
	"strconv"

	"gitlab.com/wolfplus/gamespace-lobby/db/model"
	// "reflect"
	// "fmt"
)

func GetItem(ty model.AuthType, guestAcc int64, ip string) (*model.Item, error) {
	logger.Println("get guest id ", guestAcc)

	i := model.Item{}
	_, err := db.Where("uid = ?", guestAcc).Get(&i)
	if err != nil {
		logger.Println("insert error ", err.Error())
		return nil, err
	}

	return &i, nil
}

func CreateItem(guestAcc int64, ip string) (*model.Item, error) {
	logger.Println("create guest id ", guestAcc)

	i := model.Item{
		Uid:  guestAcc,
		LaBa: 1,
	}
	_, err := db.Insert(&i)
	if err != nil {
		logger.Println("insert error ", err.Error())
		return nil, err
	}

	return &i, nil
}

func UpdateItem(ty model.AuthType, guestAcc int64, itemName string, operate string, number int64, ip string) (*model.Item, error) {
	logger.Println("db side here yo ")

	i := model.Item{}
	// _, _ = db.Get(&i)

	has, err := db.Where("uid=?", guestAcc).Get(&i)

	if err != nil {
		logger.Println("the error is " + string(err.Error()))
		return nil, err
	}

	if has {
		// logger.Println("can get that item ")
		logger.Println("see item id " + strconv.Itoa(int(i.Uid)))
		// logger.Println("see pass in guest acocunt "+ strconv.Itoa(int(guestAcc)))
		// fmt.Println(reflect.TypeOf(i.Uid))
		// fmt.Println(reflect.TypeOf(guestAcc))

		switch itemName {
		case "LaBa":
			switch operate {
			case "plus":
				i.LaBa += number
			case "minus":
				i.LaBa -= number
			}
		}
		// affect, err := db.ID(guestAcc).Update(i)

		affect, err := db.ID(i.Uid).Cols("la_ba").Update(i)

		if err != nil {
			logger.Println("error here ", err)
			return nil, err
		}

		if affect != 0 {
			logger.Println("can update that item ")
		} else {
			logger.Println("item didn't update ")
		}
	}

	return &model.Item{}, nil

}
