package game

import "math/rand"

type (
	Room struct {
		// players map[int64]*Player // room 的所有玩家押注
		// limitedConfig
		SmallRaise int64 // 前2轮加注金额
		BigRaise   int64 // 后2轮加注金额
		IsLimited  bool  // 是否有限
		// to config:
		Title           string // 场次名字 e.g. 初级场
		GameCoinToEnter int64  // 准入金额
		SmallBlind      int64  // 小盲注金额
		BigBlind        int64  // 大盲注金额
		manager         *Manager
		tables          []*Table // dynamic
		maxTableSize    int      // start at 20
	}
)

func (r *Room) ChooseBestTable() *Table {
	var thisTable *Table = nil
	currentTotalTable := len(r.tables)
	maxTableSize := r.maxTableSize

	if currentTotalTable == 0 { // no table yet
		thisTable = r.makeTable()
		return thisTable
	}

	tablesSmallerThree := []int{} // 人数少过3优先进入
	tablesBiggerThree := []int{}  // 3<=人数<6 60%进入游戏中牌桌 40%进入全新的牌桌
	tablesEmpty := []int{}

	for i := 0; i < currentTotalTable; i++ {
		thisTable = r.tables[i]
		thisNumSeatedPlayers := thisTable.numOfSeatedPlayers()
		if thisNumSeatedPlayers < 3 && thisNumSeatedPlayers > 0 {
			tablesSmallerThree = append(tablesSmallerThree, i)
			continue
		}
		if thisNumSeatedPlayers >= 3 && thisNumSeatedPlayers < 6 {
			tablesBiggerThree = append(tablesBiggerThree, i)
			continue
		}
		if thisNumSeatedPlayers == 0 {
			tablesEmpty = append(tablesEmpty, i)
			continue
		}
	}

	if len(tablesSmallerThree) > 0 {
		randomIndex := rand.Intn(len(tablesSmallerThree))
		if thisTable = r.tables[tablesSmallerThree[randomIndex]]; thisTable != nil {
			return thisTable
		}
	}

	if len(tablesBiggerThree) > 0 {
		if randomRange(1, 10) >= 6 || currentTotalTable == maxTableSize { // 60%
			randomIndex := rand.Intn(len(tablesBiggerThree))
			if thisTable = r.tables[tablesBiggerThree[randomIndex]]; thisTable != nil {
				return thisTable
			}
		}
	}

	if len(tablesEmpty) > 0 {
		randomIndex := rand.Intn(len(tablesEmpty))
		if thisTable = r.tables[tablesEmpty[randomIndex]]; thisTable != nil {
			return thisTable
		}
	}

	thisTable = r.makeTable() // new table if cannot meet any of the aforementioned criteria

	if r.maxTableSize < len(r.tables) {
		r.maxTableSize += 5
	}

	if thisTable == nil {
		thisTable = r.tables[0] // defensive programming
	}

	return thisTable
}

func (r *Room) makeTable() *Table {
	var table *Table = nil

	if r.IsLimited {
		table = newLimitedTable(r.Title, r.GameCoinToEnter, r.SmallBlind, r.BigBlind, r.SmallRaise, r.BigRaise)
	} else {
		table = newTable(r.Title, r.GameCoinToEnter, r.SmallBlind, r.BigBlind)
	}

	r.tables = append(r.tables, table)

	return table
}

// init :

func NewLimitedRoom(title string, gameCoinToEnter, smallBlind, bigBlind,
	smallRaise, bigRaise int64) *Room {
	// 有限加注金额是固定的
	room := NewRoom(title, gameCoinToEnter, smallBlind, bigBlind)

	room.IsLimited = true
	room.SmallRaise = smallRaise
	room.BigRaise = bigRaise

	return room
}

func NewRoom(title string, gameCoinToEnter, smallBlind, bigBlind int64) *Room {
	return &Room{
		Title:           title,
		GameCoinToEnter: gameCoinToEnter,
		SmallBlind:      smallBlind,
		BigBlind:        bigBlind,
		IsLimited:       false,
		manager:         defaultManager,
		tables:          []*Table{}, // init
		maxTableSize:    20,
	}
}
