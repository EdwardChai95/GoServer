package db

import (
	"log"

	"gitlab.com/wolfplus/gamespace-lobby/db/model"
	"gitlab.com/wolfplus/gamespace-lobby/helper"
)

func PlayerObtainedNewYearReward(uid string) {
	now := helper.GetCurrentShanghaiTime()

	log.Println("month: ", int(now.Month()))
	log.Println("day: ", now.Day())
	// 2 12 - 17
	// date := now.Unix()
	// println("date:", date)
	println("uid:", uid)
	var day1, day2, day3, day4, day5 string
	if int(now.Month()) == 2 && now.Day() < 17 { // day6
		_, err := db.Exec("Update `new_year_event` set `day6` = 1 WHERE uid = '" + uid + "'")
		if err != nil {
			log.Println(err)
		}
	}
	if int(now.Month()) == 2 && now.Day() < 12 { // day1
		day, err1 := db.QueryString("select `day1` from `new_year_event` where uid = '" + uid + "'")
		if err1 != nil {
			log.Println(err1)
		}
		if len(day) > 0 {
			day1 = day[0]["day5"]
			if day1 != "1" {
				_, err2 := db.Exec("Update `new_year_event` set `day1` = 2 WHERE uid = '" + uid + "'")
				if err2 != nil {
					log.Println(err2)
				}
			}
		}
		_, err3 := db.Exec("Update `new_year_event` set `day2` = 1 WHERE uid = '" + uid + "'")
		if err3 != nil {
			log.Println(err3)
		}
	}
	if int(now.Month()) == 2 && now.Day() < 13 { // day2
		day, err1 := db.QueryString("select `day2` from `new_year_event` where uid = '" + uid + "'")
		if err1 != nil {
			log.Println(err1)
		}
		if len(day) > 0 {
			day2 = day[0]["day5"]
			if day2 != "1" {
				_, err2 := db.Exec("Update `new_year_event` set `day2` = 2 WHERE uid = '" + uid + "'")
				if err2 != nil {
					log.Println(err2)
				}
			}
		}
		_, err3 := db.Exec("Update `new_year_event` set `day3` = 1 WHERE uid = '" + uid + "'")
		if err3 != nil {
			log.Println(err3)
		}
	}
	if int(now.Month()) == 2 && now.Day() < 14 { // day3
		day, err1 := db.QueryString("select `day3` from `new_year_event` where uid = '" + uid + "'")
		if err1 != nil {
			log.Println(err1)
		}
		if len(day) > 0 {
			day3 = day[0]["day5"]
			if day3 != "1" {
				_, err2 := db.Exec("Update `new_year_event` set `day3` = 2 WHERE uid = '" + uid + "'")
				if err2 != nil {
					log.Println(err2)
				}
			}
		}
		_, err3 := db.Exec("Update `new_year_event` set `day4` = 1 WHERE uid = '" + uid + "'")
		if err3 != nil {
			log.Println(err3)
		}
	}
	if int(now.Month()) == 2 && now.Day() < 15 { // day4
		day, err1 := db.QueryString("select `day4` from `new_year_event` where uid = '" + uid + "'")
		if err1 != nil {
			log.Println(err1)
		}
		if len(day) > 0 {
			day4 = day[0]["day4"]
			if day4 != "1" {
				_, err2 := db.Exec("Update `new_year_event` set `day4` = 2 WHERE uid = '" + uid + "'")
				if err2 != nil {
					log.Println(err2)
				}
			}
		}
		_, err3 := db.Exec("Update `new_year_event` set `day5` = 1 WHERE uid = '" + uid + "'")
		if err3 != nil {
			log.Println(err3)
		}
	}
	if int(now.Month()) == 2 && now.Day() < 16 { // day5
		day, err1 := db.QueryString("select `day5` from `new_year_event` where uid = '" + uid + "'")
		if err1 != nil {
			log.Println(err1)
		}
		if len(day) > 0 {
			day5 = day[0]["day5"]
			if day5 != "1" {
				_, err2 := db.Exec("Update `new_year_event` set `day5` = 2 WHERE uid = '" + uid + "'")
				if err2 != nil {
					log.Println(err2)
				}
			}
		}
		_, err3 := db.Exec("Update `new_year_event` set `day6` = 1 WHERE uid = '" + uid + "'")
		if err3 != nil {
			log.Println(err3)
		}
	}
}

func GetEvent(ty model.AuthType, guestAcc int64) (*model.NewYearEvent, error) {
	n := model.NewYearEvent{Uid: guestAcc}
	has, err := db.Get(&n)

	if err != nil {
		logger.Println(err)
	}

	if !has {
		logger.Println("try get guest acc but it not exist")
		return nil, err
	}

	return &n, err
}

func CreateEvent(ty model.AuthType, guestAcc int64) (*model.NewYearEvent, error) {
	n := &model.NewYearEvent{
		Uid:  guestAcc,
		Day1: 0,
		Day2: 0,
		Day3: 0,
		Day4: 0,
		Day5: 0,
		Day6: 0,
	}
	_, err := db.Insert(n)
	if err != nil {
		logger.Println("insert error ", err.Error())
		return nil, err
	}
	return n, nil
}
