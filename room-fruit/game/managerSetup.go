package game

import (
	"sync"

	"gitlab.com/wolfplus/gamespace-lhd/protocol"
)

func (m *Manager) addToHistoryList() {
	// add to history list
	if len(m.historyList) >= 50 { // wk add from 20 to 50
		m.historyList = m.historyList[:len(m.historyList)-1]
	}
	m.historyList = append([]*protocol.WinningItem{m.selectedWinningItem}, m.historyList...)
}

func (m *Manager) resetForBetting() {
	var wg sync.WaitGroup
	for i := range RoomRobotPlayers {
		wg.Add(1)
		// reset robot players
		go func(rp *RobotPlayer) {
			if rp.gameCoin <= 1000 {
				rp = newRobotPlayer()
			} else {
				rp.currentBettings = map[int]int64{}
				rp.awarded = 0
			}
			wg.Done()
		}(RoomRobotPlayers[i])
	}

	wg.Wait()
	players := m.getPlayers()

	for uid := range players {
		wg.Add(1)
		go func(uid int64) {
			if p, ok := m.players[uid]; ok {
				p.betResult = protocol.BetResult{}
				p.currentBettings = map[int]int64{}
				p.awarded = 0
			}
			wg.Done()
		}(uid)
	}

	wg.Wait()
}
