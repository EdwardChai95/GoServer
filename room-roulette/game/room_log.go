package game

import (
	"strconv"
	"sync"

	"gitlab.com/wolfplus/gamespace-lhd/db"
)

func (r *Room) NewLogInformation(logInformation map[string]string) {
	r.LogInformations = append(r.LogInformations, logInformation)
}

func (r *Room) InsertAllLogInformations(gameLogInformation map[string]string, resultInfo string) {
	var wg sync.WaitGroup
	paramsInt := db.SendLogInformation(gameLogInformation)
	params := strconv.FormatInt(paramsInt, 10)

	if paramsInt == -1 {
		return
	}

	// logger.Println("InsertAllLogInformations")
	// logger.Println(gameLogInformation)

	for _, logInformation := range r.LogInformations {
		logInformation["otherInfo"] = "[开奖：" + resultInfo + "]" +
			logInformation["otherInfo"] + "[参数：" + params + "]"
		logInformation["params"] = params
		wg.Add(1)
		go func(logInformation map[string]string) {
			db.SendLogInformation(logInformation)
			wg.Done()
		}(logInformation)
	}

	wg.Wait()
	r.LogInformations = []map[string]string{}
}
