package db

import "log"

func CheckProxyByUid(uid string) int {
	messages, err := db.QueryString("select uid from `proxy` " +
		"where uid = '" + uid + "' LIMIT 1")
	if err != nil {
		log.Println(err)
	}
	if len(messages) > 0 {
		return 1
	}
	return 0
}

func GetTodayByUid(uid string) []map[string]string {
	select_query := "select uid, coalesce(sum(promo_num), 0) as promo_num, " +
		"coalesce(sum(active_num), 0) as active_num, " +
		"coalesce(sum(count_completed), 0) as count_completed " +
		"from `proxy` " +
		"where uid = '" + uid + "' " +
		"and date(operating_time) >= curdate() "

	today, err := db.QueryString(select_query)
	if err != nil {
		log.Println(err)
	}

	return today
}

func GetTotalByUid(uid string) []map[string]string {
	select_query := "select uid, coalesce(sum(promo_num), 0) as promo_num, " +
		"coalesce(sum(active_num), 0) as active_num, " +
		"coalesce(sum(count_completed), 0) as count_completed " +
		"from `proxy` " +
		"where uid = '" + uid + "' " +
		"and date(operating_time) <= curdate() "

	total, err := db.QueryString(select_query)
	if err != nil {
		log.Println(err)
	}

	return total
}

func GetListByUid(uid string) []map[string]string {
	select_query := "select uid, coalesce(sum(promo_num), 0) as promo_num, " +
		"coalesce(sum(active_num), 0) as active_num, " +
		"coalesce(sum(count_completed), 0) as count_completed " +
		"from `proxy` " +
		"where uid = '" + uid + "' "

	lists, err := db.QueryString(select_query)
	if err != nil {
		log.Println(err)
	}

	return lists
}
