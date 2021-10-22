package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lonng/nex"
	"gitlab.com/wolfplus/gamespace-lobby/db"
	"gitlab.com/wolfplus/gamespace-lobby/helper"
)

func completeTaskHandler(r *http.Request) (map[string]string, error) {
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}
	reqJson := helper.ReadParameters(r)
	if myTaskItem, ok := tasks[reqJson["taskKey"]]; ok {
		if success, gamecoin := db.CompleteTask(myTaskItem, uid); success {
			data := map[string]interface{}{
				"uid":       uid,
				"game_coin": gamecoin, //need giftuser.gamecoin to update
			}
			LobbyCoinUpdate(data)
			return map[string]string{
				//"success": "领取成功",
				"success": "sự thành công",
			}, nil
		}
	}
	return nil, errors.New("unable to complete task")
}

func getMyTasksHandler(r *http.Request) ([]*db.TaskItem, error) { // /task/get
	uid, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}

	// get progress
	progressStats := db.GetTaskProgress(uid)

	for i, task := range tasks {
		task.Progress = fmt.Sprintf("%v/%v", progressStats[i], task.Guage)
		if progressStats[i] >= task.Guage {
			task.IsAchived = true
			// check if tasks awarded
			task.IsAwarded = db.IsTaskAwarded(task, uid)
		} else {
			task.IsAchived = false
			task.IsAwarded = false
		}
	}

	myTasks := []*db.TaskItem{}

	myTasks = append(myTasks, tasks[db.TASK_TIMEPLAYED])
	myTasks = append(myTasks, tasks[db.TASK_ROUNDSPLAYED])
	myTasks = append(myTasks, tasks[db.TASK_WONAMOUNT1])
	//myTasks = append(myTasks, tasks[db.TASK_WONAMOUNT2])
	//myTasks = append(myTasks, tasks[db.TASK_WONAMOUNT3])

	// tasks[0].Progress = fmt.Sprintf("%v/%v", progress[db.TASK_TIMEPLAYED], tasks[0].Guage)
	// tasks[1].Progress = fmt.Sprintf("%v/%v", progress[db.TASK_ROUNDSPLAYED], tasks[1].Guage)
	// tasks[2].Progress = fmt.Sprintf("%v/%v", progress[db.TASK_PROGRESS_WIN_TOTAL], tasks[2].Guage)
	// tasks[3].Progress = fmt.Sprintf("%v/%v", progress[db.TASK_PROGRESS_WIN_TOTAL], tasks[3].Guage)
	// tasks[4].Progress = fmt.Sprintf("%v/%v", progress[db.TASK_PROGRESS_WIN_TOTAL], tasks[4].Guage)

	return myTasks, nil

}

func initTasks() {
	tasks[db.TASK_TIMEPLAYED].Key = db.TASK_TIMEPLAYED
	tasks[db.TASK_TIMEPLAYED].Name = db.GetConfigValueByKey(db.TASK_TIMEPLAYED)
	tasks[db.TASK_TIMEPLAYED].Reward = helper.StringToInt64(db.GetConfigValueByKey(db.TASK_TIMEPLAYED_REWARD))
	tasks[db.TASK_TIMEPLAYED].Guage = helper.StringToInt64(db.GetConfigValueByKey(db.TASK_TIMEPLAYED_GUAGE))

	tasks[db.TASK_ROUNDSPLAYED].Key = db.TASK_ROUNDSPLAYED
	tasks[db.TASK_ROUNDSPLAYED].Name = db.GetConfigValueByKey(db.TASK_ROUNDSPLAYED)
	tasks[db.TASK_ROUNDSPLAYED].Reward = helper.StringToInt64(db.GetConfigValueByKey(db.TASK_ROUNDSPLAYED_REWARD))
	tasks[db.TASK_ROUNDSPLAYED].Guage = helper.StringToInt64(db.GetConfigValueByKey(db.TASK_ROUNDSPLAYED_GUAGE))

	tasks[db.TASK_WONAMOUNT1].Key = db.TASK_WONAMOUNT1
	tasks[db.TASK_WONAMOUNT1].Name = db.GetConfigValueByKey(db.TASK_WONAMOUNT1)
	tasks[db.TASK_WONAMOUNT1].Reward = helper.StringToInt64(db.GetConfigValueByKey(db.TASK_WONAMOUNT1_REWARD))
	tasks[db.TASK_WONAMOUNT1].Guage = helper.StringToInt64(db.GetConfigValueByKey(db.TASK_WONAMOUNT1_GUAGE))

	//tasks[db.TASK_WONAMOUNT2].Key = db.TASK_WONAMOUNT2
	//tasks[db.TASK_WONAMOUNT2].Name = db.GetConfigValueByKey(db.TASK_WONAMOUNT2)
	//tasks[db.TASK_WONAMOUNT2].Reward = helper.StringToInt64(db.GetConfigValueByKey(db.TASK_WONAMOUNT2_REWARD))
	//tasks[db.TASK_WONAMOUNT2].Guage = helper.StringToInt64(db.GetConfigValueByKey(db.TASK_WONAMOUNT2_GUAGE))

	//tasks[db.TASK_WONAMOUNT3].Key = db.TASK_WONAMOUNT3
	//tasks[db.TASK_WONAMOUNT3].Name = db.GetConfigValueByKey(db.TASK_WONAMOUNT3)
	//tasks[db.TASK_WONAMOUNT3].Reward = helper.StringToInt64(db.GetConfigValueByKey(db.TASK_WONAMOUNT3_REWARD))
	//tasks[db.TASK_WONAMOUNT3].Guage = helper.StringToInt64(db.GetConfigValueByKey(db.TASK_WONAMOUNT3_GUAGE))
}

func MakeTaskService() http.Handler {
	initTasks()

	router := mux.NewRouter()
	router.Handle("/task/get", nex.Handler(getMyTasksHandler)).Methods("GET")
	router.Handle("/task/complete", nex.Handler(completeTaskHandler)).Methods("POST")

	return router
}

var (
	tasks = map[string]*db.TaskItem{
		//db.TASK_TIMEPLAYED:   {Name: "游戏时间累计达1小时", Reward: 1000, Guage: 30, Description: "每日重复"},
		db.TASK_TIMEPLAYED: {Name: "Cần chơi trong 1 giờ", Reward: 1000, Guage: 30, Description: "Cần chơi trong 100 phút"},
		//db.TASK_ROUNDSPLAYED: {Name: "游戏局数累计达100局", Reward: 1000, Guage: 100, Description: "每日重复"},
		db.TASK_ROUNDSPLAYED: {Name: "Cần chơi 100 vòng", Reward: 1000, Guage: 100, Description: "Cần chơi 10 trò chơi"},
		//db.TASK_WONAMOUNT1:   {Name: "任意游戏单次赢取达100万游戏币", Reward: 2000, Guage: 1000000, Description: "每日重复"},
		db.TASK_WONAMOUNT1: {Name: "Giành được 100V cùng một lúc", Reward: 5000, Guage: 3000000, Description: "Thắng 500.000 trong một lần chơi"},
		//db.TASK_WONAMOUNT2:   {Name: "任意游戏单次赢取达300万游戏币", Reward: 5000, Guage: 3000000, Description: "每日重复"},
		//db.TASK_WONAMOUNT2: {Name: "Giành được 300V cùng một lúc", Reward: 5000, Guage: 3000000, Description: "Cần giành được 2 triệu trong một vòng"},
		//db.TASK_WONAMOUNT3:   {Name: "任意游戏单次赢取达500万游戏币", Reward: 10000, Guage: 5000000, Description: "每日重复"},
		//db.TASK_WONAMOUNT3: {Name: "Giành được 500V cùng một lúc", Reward: 10000, Guage: 5000000, Description: "Cần giành được 5 triệu trong một vòng"},
	}
)
