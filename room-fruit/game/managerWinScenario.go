package game

import (
	"math/rand"
	"time"

	"gitlab.com/wolfplus/gamespace-lhd/protocol"
)

func (m *Manager) resultSpecial() {
	randomInt := randomRange(1, 100)
	var winningBetZoneIndexes = []int{}

	specialPrize := ""
	musicFileName := "LUCK"
	specialPrizeImage := ""
	selectedWinningItems := []*protocol.WinningItem{}

	var goodluckKey int = 21
	if randomRange(1, 2) > 1 {
		goodluckKey = 9
	}

	if randomInt <= 10 { //true { //
		// 九莲宝灯bigfruit
		specialPrize = "bigFruit"
		m.specialPrizeChinese = "九莲宝灯"
		specialPrizeImage = "history_special_1"
		// 九莲宝灯：除BAR外其他水果均计压中，且都为大水果，
		// 开奖时亮9个大水果，可重复但需要都包含；
		var i int = 0
		for x, winningItem := range m.winningItems {
			if winningItem.IsBigFruit == true {
				selectedWinningItems = append(selectedWinningItems, winningItem)
				winningBetZoneIndexes = append(winningBetZoneIndexes, x)
				i++
			}
		}
		// random all isBigFruitItem, current length = 13,
		rand.Shuffle(len(winningBetZoneIndexes), func(i, j int) {
			winningBetZoneIndexes[i], winningBetZoneIndexes[j] = winningBetZoneIndexes[j], winningBetZoneIndexes[i]
			selectedWinningItems[i], selectedWinningItems[j] = selectedWinningItems[j], selectedWinningItems[i]
		})
		winningBetZoneIndexes = winningBetZoneIndexes[0:9] // slice to 9 items
		selectedWinningItems = selectedWinningItems[0:9]
	} else if randomInt <= 40 {
		// 超级火车train
		specialPrize = "train"
		m.specialPrizeChinese = "超级火车"
		specialPrizeImage = "history_special_2"
		numToSelect := randomRange(2, 6)
		for i := 0; i < numToSelect; i++ {
			winningItem, winningkey := m.chooseWinningItem()
			selectedWinningItems = append(selectedWinningItems, winningItem)
			winningBetZoneIndexes = append(winningBetZoneIndexes, winningkey)
		}
	} else {
		// 仙女散花fairy
		specialPrize = "fairy"
		m.specialPrizeChinese = "仙女散花"
		specialPrizeImage = "history_special_3"
		numToSelect := randomRange(3, 4)
		for i := 0; i < numToSelect; i++ {
			winningItem, winningkey := m.chooseWinningItem()
			selectedWinningItems = append(selectedWinningItems, winningItem)
			winningBetZoneIndexes = append(winningBetZoneIndexes, winningkey)
		}
	}

	m.selectedWinningItem = &protocol.WinningItem{WinningBetZoneIndex: -1,
		Image: specialPrizeImage, Music: musicFileName}

	m.addToHistoryList() // wk add
	m.group.Broadcast("animationPhase", &protocol.AnimationPhaseResponse{
		Deadline:     m.deadline,
		SpecialPrize: specialPrize,
		// SelectedWinningItem:  &protocol.WinningItem{Image: specialPrizeImage},
		SelectedWinningItem:   m.selectedWinningItem,
		SelectedWinningItems:  selectedWinningItems,
		WinningKey:            goodluckKey,
		WinningBetZoneIndexes: winningBetZoneIndexes,
	})

	m.selectedWinningItemKey = goodluckKey // 9 or 21
	m.selectedWinningItemOdd = 1
	m.selectedWinningItems = selectedWinningItems
	m.specialPrize = specialPrize
}

func (m *Manager) resultCommon() {
	// 普通开奖跑灯时间10
	m.deadline = time.Now().UTC().Add(time.Second * time.Duration(10))

	// select winning item based on probability 普通开奖
	winningKey := m.selectedWinningItemKey
	m.selectedWinningItem, winningKey = m.chooseWinningItem()
	// winningKey = 0                                     // testing
	// m.selectedWinningItem = m.winningItems[winningKey] // testing

	// tell to the slave when to expect next phase
	m.group.Broadcast("animationPhase", &protocol.AnimationPhaseResponse{
		Deadline:            m.deadline,
		SelectedWinningItem: m.selectedWinningItem,
		WinningKey:          winningKey,
	})

	m.selectedWinningItemOdd = m.selectedWinningItem.Odds
	// add to history list
	m.addToHistoryList()

	m.calculateWinLoseCoin()
}

func (m *Manager) resultPasskill() { // 通杀passkill
	m.deadline = time.Now().UTC().Add(time.Second * time.Duration(13))

	// glarray = []{9,21}
	// max len(glarray)
	// glarray[randomrange-1] < -- winningkey
	var goodluckKey int = 21
	if randomRange(1, 2) > 1 {
		goodluckKey = 9
	}
	m.selectedWinningItem = &protocol.WinningItem{WinningBetZoneIndex: -1,
		Image: "history_special_no", Music: "LUCK"}
	m.selectedWinningItems = []*protocol.WinningItem{m.selectedWinningItem}
	m.specialPrizeChinese = "通杀"
	m.selectedWinningItemKey = goodluckKey

	m.group.Broadcast("animationPhase", &protocol.AnimationPhaseResponse{
		Deadline:            m.deadline,
		SpecialPrize:        "passkill",
		WinningKey:          goodluckKey,
		SelectedWinningItem: m.selectedWinningItem,
	})

	m.addToHistoryList()
	m.calculateWinLoseCoin()
	m.resetForBetting()          // because no result phase
	m.setGameSystemLoggingInfo() // because no result phase
	// count down to next game status
	// if m.sessionCount() == 0 {
	// 	m.gameStatus = "nogame"
	// } else {
	// 	go func() {
	s := m.deadline.Sub(time.Now()).Seconds()
	time.Sleep(time.Duration(s) * time.Second)
	m.bettingPhase()
	// 	}()
	// }
}

func (m *Manager) chooseWinningItem() (*protocol.WinningItem, int) {
	maxProbability := 0 // max range
	var probabilityRange = []int{}
	for i := 0; i < len(m.winningItems); i++ {
		maxProbability += m.winningItems[i].Probability
		probabilityRange = append(probabilityRange, maxProbability)
	}
	randProbability := randomRange(0, maxProbability) // chance of getting selected winning item 0 - max
	winningKey := len(m.winningItems) - 1
	// winningProbability := m.winningItems[winningKey].Probability // for logging later, maybe no need
	selectedWinningItem := m.winningItems[winningKey] // key 9 & key 21 = goodluck
	for i, v := range probabilityRange {              // i = index, v = probability threshold
		if randProbability <= v { // inside of probability threshold
			winningKey = i
			if m.winningItems[winningKey].Probability == 0 {
				winningKey += 1 // goodluck cannot win
			}
			// winningProbability = m.winningItems[winningKey].Probability
			selectedWinningItem = m.winningItems[winningKey]
			break
		}
	}
	m.selectedWinningItemKey = winningKey
	// return m.winningItems[4], 4 // FOR TESTING
	return selectedWinningItem, winningKey
}
