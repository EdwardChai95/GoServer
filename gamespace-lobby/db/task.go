package db

import (
	"fmt"

	"gitlab.com/wolfplus/gamespace-lobby/helper"
)

func IsTaskAwarded(task *TaskItem, uid string) bool { // means collected for the day
	logs, err := db.QueryString("SELECT COUNT(*) as num_result FROM `log_information` " +
		"WHERE uid = '" + uid + "' AND DATE(`operating_time`)='" + helper.GetCurrentShanghaiDateOnlyString() + "' " +
		"AND `task_key`='" + task.Key + "'" +
		"ORDER BY log_information_id DESC LIMIT 1")

	if err != nil {
		logger.Error(err)
	}

	if len(logs) > 0 {
		if helper.StringToInt(logs[0]["num_result"]) > 0 {
			return true
		}
	}

	return false
}

func CompleteTask(task *TaskItem, uid string) (bool, int64) {
	// logger.Println("complete task")
	// check task progress for today
	progress := GetTaskProgress(uid)
	// logger.Println(progress[task.Key])
	// logger.Println(task.Guage)

	if progress[task.Key] >= task.Guage {
		logger.Println("complete task")
		// 任务领取
		user, _ := GetUser(helper.StringToInt64(uid)) // before
		updatedAmount := UpdateGameCoin(user, task.Reward,
			task.Name, "任务领取", task.Name+"任务领取"+fmt.Sprintf("%v", task.Reward), task.Key)

		return true, updatedAmount
	}
	return false, 0
}

func getRoundsPlayedTimeByUid(uid string) int64 { // return in minutes
	ref := map[string]int64{ // reference seconds of each game
		"动物乐园":   57,
		"豪车汇":    36,
		"水果":     45,
		"拼十":     19,
		"others": 15,
	}

	var timePlayedSeconds int64 = 0
	gamelogs, err := db.QueryString("SELECT DISTINCT game from `log_information` where game != ''")

	if err != nil {
		logger.Error(err)
	}

	if len(gamelogs) > 0 {
		for _, gamelog := range gamelogs {
			game := gamelog["game"]
			gameSeconds := ref["others"]
			if gs, ok := ref[game]; ok {
				gameSeconds = gs
			}
			sql := "SELECT count(*) as num_result FROM `log_information` " +
				"WHERE `game`='" + game + "' AND `uid` = '" + uid + "' " +
				"AND DATE(`operating_time`)='" + helper.GetCurrentShanghaiDateOnlyString() + "' LIMIT 1"
			logs, err := db.QueryString(sql)
			if err != nil {
				logger.Error(err)
			}

			if len(logs) > 0 {
				timePlayedSeconds += helper.StringToInt64(logs[0]["num_result"]) * gameSeconds
				logger.Printf("game: %v", game)
				logger.Printf("logs[0][\"num_result\"]: %v", logs[0]["num_result"])
				logger.Printf("gameSeconds: %v", gameSeconds)
				logger.Printf("timePlayedSeconds: %v", timePlayedSeconds)
			}
		}
		if timePlayedSeconds >= 60 {
			return timePlayedSeconds / 60 // minutes
		}
	}

	return 0
}

func getTaskStatsByUid(uid, task_cond_type string) int64 {
	sql := "SELECT "

	if task_cond_type == TASKCOND_type_roundsPlayed ||
		task_cond_type == TASKCOND_type_roundsPlayedWithParam {
		sql += TASKCOND_type_roundsPlayed // count(*)
	} else {
		sql += TASKCOND_type_winTotal
	}

	sql += " as num_result FROM `log_information` " +
		"WHERE uid = '" + uid + "' AND DATE(`operating_time`)='" + helper.GetCurrentShanghaiDateOnlyString() + "' " +
		"AND game !='' "

	if task_cond_type == TASKCOND_type_roundsPlayedWithParam {
		sql += " AND `bet_total` >= '100' " // hard coded 100
	}

	sql += " ORDER BY `used` DESC, log_information_id DESC LIMIT 1"

	logger.Printf("getTaskStatsByUid sql: %v", sql)

	logs, err := db.QueryString(sql)

	if err != nil {
		logger.Error(err)
	}

	if len(logs) > 0 {

		// logger.Printf("logs: %v", logs)
		if progress := helper.StringToInt64(logs[0]["num_result"]); progress > 0 {
			return progress
		}
	}
	return 0
}

func GetTaskProgress(uid string) map[string]int64 {

	progress := map[string]int64{}

	// rounds played
	rounds_played := getRoundsPlayedTimeByUid(uid) // minutes
	rounds_played_with_param := getTaskStatsByUid(uid, TASKCOND_type_roundsPlayedWithParam)
	// logger.Printf("rounds_played: %v", rounds_played)
	progress[TASK_TIMEPLAYED] = rounds_played // minutes
	progress[TASK_ROUNDSPLAYED] = rounds_played_with_param

	// win total
	win_total := getTaskStatsByUid(uid, TASKCOND_type_winTotal)
	// logger.Printf("win_total: %v", win_total)
	progress[TASK_WONAMOUNT1] = win_total
	progress[TASK_WONAMOUNT2] = win_total
	progress[TASK_WONAMOUNT3] = win_total

	return progress
}

func GetConfigValueByKey(key string) string {
	if config := GetConfigByKey(key); config != nil {
		return config["value"]
	}
	return "0"
}

func GetConfigByKey(key string) map[string]string {

	configs, err := db.QueryString("select * from `config` WHERE `key` = '" + key +
		"' LIMIT 1")

	if err != nil {
		logger.Error(err)
	}
	if len(configs) > 0 {
		return configs[0]
	}

	return nil
}

type TaskItem struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Reward      int64  `json:"reward"`
	Guage       int64  `json:"guage"`
	Progress    string `json:"progress"` // varies
	Description string `json:"description"`
	IsAchived   bool   `json:"isAchived"` // varies
	IsAwarded   bool   `json:"isAwarded"` // varies
}

const (
	TASK_TIMEPLAYED        string = "task_timePlayed"
	TASK_TIMEPLAYED_REWARD string = TASK_TIMEPLAYED + "_reward"
	TASK_TIMEPLAYED_GUAGE  string = TASK_TIMEPLAYED + "_guage"

	TASK_ROUNDSPLAYED        string = "task_roundsPlayed"
	TASK_ROUNDSPLAYED_REWARD string = TASK_ROUNDSPLAYED + "_reward"
	TASK_ROUNDSPLAYED_GUAGE  string = TASK_ROUNDSPLAYED + "_guage"

	TASK_WONAMOUNT1        string = "task_wonAmount1"
	TASK_WONAMOUNT1_REWARD string = TASK_WONAMOUNT1 + "_reward"
	TASK_WONAMOUNT1_GUAGE  string = TASK_WONAMOUNT1 + "_guage"

	TASK_WONAMOUNT2        string = "task_wonAmount2"
	TASK_WONAMOUNT2_REWARD string = TASK_WONAMOUNT2 + "_reward"
	TASK_WONAMOUNT2_GUAGE  string = TASK_WONAMOUNT2 + "_guage"

	TASK_WONAMOUNT3        string = "task_wonAmount3"
	TASK_WONAMOUNT3_REWARD string = TASK_WONAMOUNT3 + "_reward"
	TASK_WONAMOUNT3_GUAGE  string = TASK_WONAMOUNT3 + "_guage"

	// TASK_PROGRESS_WIN_TOTAL string = "win_total"

	TASKCOND_type_roundsPlayed          string = "COUNT(*)"                            // used in sql
	TASKCOND_type_roundsPlayedWithParam string = "TASKCOND_type_roundsPlayedWithParam" // not used in sql
	TASKCOND_type_winTotal              string = "`used`"                              //"IFNULL(SUM(`used`), 0)"// used in sql
)
