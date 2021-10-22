package game

import (
	"strconv"
	"sync"

	"gitlab.com/wolfplus/gamespace-lhd/db"
)

func (m *Manager) NewLogInformation(logInformation map[string]string) {
	m.Lock()
	m.logInformations = append(m.logInformations, logInformation)
	m.Unlock()
}

func (m *Manager) InsertAllLogInformations(gameLogInformation map[string]string, winningItemLogInfo string) {
	var wg sync.WaitGroup
	paramsInt := db.SendLogInformation(gameLogInformation)
	params := strconv.FormatInt(paramsInt, 10)

	if paramsInt == -1 {
		return
	}

	for _, logInformation := range m.logInformations {
		logInformation["otherInfo"] = "[开奖：" + winningItemLogInfo + "]" + logInformation["otherInfo"] + "[参数：" + params + "]"
		logInformation["params"] = params
		wg.Add(1)
		go func(logInformation map[string]string) {
			db.SendLogInformation(logInformation)
			wg.Done()
		}(logInformation)
	}

	wg.Wait()
	m.logInformations = []map[string]string{}
}
