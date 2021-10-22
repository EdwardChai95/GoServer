package game

import (
	"fmt"
	"strconv"
	"sync"

	"gitlab.com/wolfplus/gamespace-lhd/db"
)

func (r *Room) NewLogInformation(logInformation map[string]string) {
	r.logInformations = append(r.logInformations, logInformation)
}

func (r *Room) InsertAllLogInformations(gameLogInformation map[string]string, gamblersLog string) {
	var wg sync.WaitGroup
	paramsInt := db.SendLogInformation(gameLogInformation)
	params := strconv.FormatInt(paramsInt, 10)

	if paramsInt == -1 {
		return
	}

	for _, logInformation := range r.logInformations {
		logInformation["otherInfo"] = "[开奖：" + gamblersLog + "]" + "[参数：" + params + "]"
		logInformation["params"] = params
		wg.Add(1)
		fmt.Println("log:", logInformation)
		go func(logInformation map[string]string) {
			db.SendLogInformation(logInformation)
			wg.Done()
		}(logInformation)
	}

	wg.Wait()
	r.logInformations = []map[string]string{}
}
