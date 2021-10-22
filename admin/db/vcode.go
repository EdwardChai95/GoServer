package db

import (
	"fmt"
)

func GetVcode(searchParams map[string]string) []map[string]string {

	queryStr := "SELECT * "
	queryStr += "FROM `v_code` "
	queryStr += "WHERE 1=1 "

	if searchParams["phone_number"] != "" {
		queryStr += "AND `phone_number` = '" + searchParams["phone_number"] + "' "
	} else {
		queryStr += "AND `phone_number` = 0"
	}

	fmt.Println(queryStr)
	reports, err := db.QueryString(queryStr)
	if err != nil {
		logger.Printf(queryStr)
		logger.Error(err)
	}

	if len(reports) > 0 {
		return reports
	}

	return nil
}
