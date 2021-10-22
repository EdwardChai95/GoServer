package pokersolver

import (
	"sort"

	"gitlab.com/wolfplus/gamespace-lhd/protocol"
)

const (
	ROYAL_STRAIGHT_FLUSH = 10
	STRAIGHT_FLUSH       = 9
	QUAD                 = 8
	BOAT                 = 7
	FLUSH                = 6
	STRAIGHT             = 5
	SET                  = 4
	TWO_PAIR             = 3
	PAIR                 = 2
	HIGH_CARD            = 1
)

var RANK_TITLE = map[int]string{
	ROYAL_STRAIGHT_FLUSH: "皇家同花顺",
	STRAIGHT_FLUSH:       "同花顺",
	QUAD:                 "四 条",
	BOAT:                 "葫 芦",
	FLUSH:                "同 花",
	STRAIGHT:             "顺 子",
	SET:                  "三 条",
	TWO_PAIR:             "两 对",
	PAIR:                 "对 子",
	HIGH_CARD:            "高 牌",
}

type PlayerHand struct {
	Hand          []protocol.Card
	Rank          int
	SeatNumber    int
	myHighestCard protocol.Card
}

func CompareHands(playerHands []PlayerHand) []int {
	if len(playerHands) == 0 {
		return nil
	}
	// var playerHands []PlayerHand = nil
	// for i := range highestHands {
	// 	playerHands = append(playerHands, &PlayerHand{
	// 		Hand:       highestHands[i],
	// 		Rank:       ranks[i],
	// 		SeatNumber: seatNumbers[i],
	// 	})
	// }

	sort.Slice(playerHands, func(i, j int) bool {
		return playerHands[i].Rank > playerHands[j].Rank
	})

	highestRank := playerHands[0].Rank
	var handsToCompare []PlayerHand = nil

	for _, Hand := range playerHands {
		if Hand.Rank == highestRank {
			handsToCompare = append(handsToCompare, Hand)
		}
	}

	if len(handsToCompare) > 1 {
		return compareByRank(handsToCompare, highestRank)
	} else {
		return []int{handsToCompare[0].SeatNumber}
	}
	// return compareTopCard(handsToCompare) // compare top card

}

func compareByRank(handsToCompare []PlayerHand, rank int) []int {
	// for i, playerHand := range handsToCompare {
	// 	playerHandPtr := &handsToCompare[i]
	// 	sort.Slice(playerHandPtr.Hand, func(i, j int) bool {
	// 		return playerHand.Hand[i].Value > playerHand.Hand[j].Value
	// 	})
	// }

	switch {
	// case rank == FLUSH: // compare top card
	// 	return compareTopCard(handsToCompare)
	case rank == QUAD: // compare set of 4 cards
		return compareQuads(handsToCompare)
	case rank == SET || rank == BOAT: // compare set of 3 cards
		return compareTrips(handsToCompare)
	case rank == STRAIGHT_FLUSH || rank == STRAIGHT: // compare top card
		return compareStraight(handsToCompare)
	case rank == PAIR: // compare biggest pair
		return comparePair(handsToCompare)
	case rank == TWO_PAIR:
		return compareTwoPair(handsToCompare)

	}
	return compareTopCard(handsToCompare) // compare top card
}

func compareQuads(handsToCompare []PlayerHand) []int {
	return compareSet(handsToCompare, 4)
}

func compareTrips(handsToCompare []PlayerHand) []int {
	return compareSet(handsToCompare, 3)
}

func comparePair(handsToCompare []PlayerHand) []int {
	return compareSet(handsToCompare, 2)
}

func compareTwoPair(handsToCompare []PlayerHand) []int {
	return compareSet(handsToCompare, 22)
}

func compareSet(handsToCompare []PlayerHand, highestCardType int) []int {
	var highestCard protocol.Card = protocol.Card{
		Value: 2,
		Suit:  "d",
		Face:  "2",
	}

	for x, playerHand := range handsToCompare {
		playerHandPtr := &handsToCompare[x]
		playerHandPtr.myHighestCard = highestCard
		// sort.Slice(playerHandPtr.Hand, func(i, j int) bool {
		// 	return playerHand.Hand[i].Value > playerHand.Hand[j].Value
		// })
		if highestCardType == 2 || highestCardType == 22 {
			for i := 0; i < 4; i++ { // highest card means a card to determine highest set value of a player
				if playerHand.Hand[i].Value == playerHand.Hand[i+1].Value {
					if playerHand.Hand[i].Value >= playerHandPtr.myHighestCard.Value {
						playerHandPtr.myHighestCard = playerHand.Hand[i]
					}
				}
			}
		}

		if highestCardType == 3 {
			for i := 0; i < 3; i++ {
				if playerHand.Hand[i].Value == playerHand.Hand[i+1].Value &&
					playerHand.Hand[i+1].Value == playerHand.Hand[i+2].Value {
					if playerHand.Hand[i].Value >= playerHandPtr.myHighestCard.Value {
						playerHandPtr.myHighestCard = playerHand.Hand[i]
					}
				}
			}
		}

		if highestCardType == 4 {
			for i := 0; i < 2; i++ {
				if playerHand.Hand[i].Value == playerHand.Hand[i+1].Value &&
					playerHand.Hand[i+1].Value == playerHand.Hand[i+2].Value &&
					playerHand.Hand[i+2].Value == playerHand.Hand[i+3].Value {
					if playerHand.Hand[i].Value >= playerHandPtr.myHighestCard.Value {
						playerHandPtr.myHighestCard = playerHand.Hand[i]
					}
				}
			}
		}
	}

	var winnerSeats []int = nil
	highestCard = handsToCompare[0].myHighestCard
	// var highestFace = highestCard.Face
	// sameFaceCount := 0
	highCardCount := 0

	for _, playerHand := range handsToCompare {
		if playerHand.myHighestCard.Value > highestCard.Value {
			highestCard = playerHand.myHighestCard
			// highestFace = playerHand.myHighestCard.Face
		}
	}

	for _, playerHand := range handsToCompare {
		if playerHand.myHighestCard.Value >= highestCard.Value {
			highCardCount++
		}
	}

	if highCardCount > 1 && highestCardType == 22 { // this means there is a tie for the first pair
		newHandsToCompare := []PlayerHand{}
		for _, playerHand := range handsToCompare {
			if playerHand.myHighestCard.Value == highestCard.Value {
				newHandsToCompare = append(newHandsToCompare, playerHand)
			}
		}
		handsToCompare = newHandsToCompare
		for x, playerHand := range handsToCompare {
			playerHandPtr := &handsToCompare[x]

			// sort.Slice(playerHandPtr.Hand, func(i, j int) bool {
			// 	return playerHand.Hand[i].Value > playerHand.Hand[j].Value
			// })

			for i := 0; i < 4; i++ { // highest card means a card to determine highest set value of a player
				if playerHand.Hand[i].Value == playerHand.Hand[i+1].Value &&
					playerHand.Hand[i].Value != highestCard.Value { // it has to *not* be the same as first set
					playerHandPtr.myHighestCard = playerHand.Hand[i]
				}
			}
		}

		highestCard = handsToCompare[0].myHighestCard
		highCardCount = 0

		for _, playerHand := range handsToCompare {
			if playerHand.myHighestCard.Value > highestCard.Value {
				highestCard = playerHand.myHighestCard
			}
		}

		for _, playerHand := range handsToCompare {
			if playerHand.myHighestCard.Value >= highestCard.Value {
				highCardCount++
			}
		}

	}

	if highCardCount > 1 { // sameFaceCount ==
		newHandsToCompare := []PlayerHand{}
		for _, playerHand := range handsToCompare {
			if playerHand.myHighestCard.Value == highestCard.Value {
				newHandsToCompare = append(newHandsToCompare, playerHand)
			}
		}
		handsToCompare = newHandsToCompare
		return compareTopCard(handsToCompare)
	}

	// for _, playerHand := range handsToCompare {
	// 	if playerHand.myHighestCard.Face == highestFace {
	// 		// sameFaceCount++
	// 	}
	// }

	for _, playerHand := range handsToCompare {
		if playerHand.myHighestCard.Value == highestCard.Value {
			winnerSeats = append(winnerSeats, playerHand.SeatNumber)
		}
	}
	return winnerSeats
}

func compareStraight(handsToCompare []PlayerHand) []int {
	// for i, playerHand := range handsToCompare {
	// 	playerHandPtr := &handsToCompare[i]
	// 	sort.Slice(playerHandPtr.Hand, func(i, j int) bool {
	// 		return playerHand.Hand[i].Value > playerHand.Hand[j].Value
	// 	})
	// }

	var winnerSeats []int = nil
	var highestCard protocol.Card = handsToCompare[0].Hand[0]

	for i, playerHand := range handsToCompare {
		// 	cards[0].Value-cards[4].Value == 4 ||
		var myHighestCard = playerHand.Hand[0]
		if playerHand.Hand[0].Value == 14 && playerHand.Hand[1].Value == 5 {
			myHighestCard = playerHand.Hand[1]
		}
		playerHandPtr := &handsToCompare[i]
		playerHandPtr.myHighestCard = myHighestCard

		if myHighestCard.Value > highestCard.Value {
			highestCard = playerHand.Hand[0]
		}
	}

	for _, playerHand := range handsToCompare {
		if playerHand.myHighestCard.Value == highestCard.Value || (highestCard.Value == 14 && playerHand.Hand[0].Value == 14) {
			winnerSeats = append(winnerSeats, playerHand.SeatNumber)
		}
	}

	return winnerSeats

}

func compareTopCard(handsToCompare []PlayerHand) []int {
	// for i, playerHand := range handsToCompare {
	// 	playerHandPtr := &handsToCompare[i]
	// 	sort.Slice(playerHandPtr.Hand, func(i, j int) bool {
	// 		return playerHand.Hand[i].Value > playerHand.Hand[j].Value
	// 	})
	// }

	var winnerSeats []int = nil

	var cardIndex = 0
	var highestCard protocol.Card = handsToCompare[0].Hand[cardIndex]
	var count = 0

	for {
		if cardIndex >= 5 {
			break
		}
		highestCard = handsToCompare[0].Hand[cardIndex]

		for _, playerHand := range handsToCompare {
			if playerHand.Hand[cardIndex].Value > highestCard.Value {
				highestCard = playerHand.Hand[cardIndex]
			}
		}

		for _, playerHand := range handsToCompare {
			if playerHand.Hand[cardIndex].Value == highestCard.Value {
				count++
			}
		}

		if count > 1 { // == len(handsToCompare) {// tie
			count = 0
			cardIndex++
		} else {
			for _, playerHand := range handsToCompare {
				if playerHand.Hand[cardIndex].Value == highestCard.Value {
					winnerSeats = append(winnerSeats, playerHand.SeatNumber)
				}
			}
			break
		}
	}

	if cardIndex == 5 { // all is a tie
		for _, playerHand := range handsToCompare {
			winnerSeats = append(winnerSeats, playerHand.SeatNumber)
		}
	}

	return winnerSeats
}

func GetHighestHand(heldCards, publicCards []protocol.Card) ([]protocol.Card, int) { // len has to be more than 5

	allCards := append(heldCards, publicCards...)
	// fmt.Printf("allCards: %v", allCards)
	potentialHands := combinationUtil(allCards,
		[][]protocol.Card{}, map[int]protocol.Card{}, 0, len(allCards)-1, 0, 5)
	// fmt.Println(potentialHands)
	Rank := HIGH_CARD
	var highestHand []protocol.Card = allCards
	var potentialHandIndex = 0

	for x, Hand := range potentialHands {
		// if len(publicCards) == 5 && equalHands(Hand, publicCards) {
		// 	// make sure Hand not same as publicCards
		// 	continue
		// }

		if thisRank := handRank(Hand); thisRank >= Rank {
			if thisRank == Rank {
				handsToCompare := []PlayerHand{
					{Hand: highestHand, SeatNumber: potentialHandIndex}, {Hand: Hand, SeatNumber: x}}
				indexes := compareByRank(handsToCompare, Rank)
				if !equalHands(potentialHands[indexes[0]], Hand) {
					continue
				}
			}

			Rank = thisRank
			highestHand = Hand
			potentialHandIndex = x
		}
	}

	// fmt.Printf("Rank: %v", Rank)
	// fmt.Println()
	// hh2B, _ := json.Marshal(highestHand)
	// fmt.Println("highesthand: " + string(hh2B))

	return highestHand, Rank
}

func handRank(cards []protocol.Card) int { // assuming len is 5
	sort.Slice(cards, func(i, j int) bool {
		return cards[i].Value > cards[j].Value
	})

	isFlush := false
	isStraight := false

	histogram_count := 1
	histogram_counts := []int{}
	cardValueToCompare := -1

	for i, card := range cards {
		if cardValueToCompare == -1 {
			cardValueToCompare = card.Value
		} else {
			if cardValueToCompare == card.Value {
				histogram_count++
				if i+1 == len(cards) {
					histogram_counts = append(histogram_counts, histogram_count)
				}
			} else {
				histogram_counts = append(histogram_counts, histogram_count)
				histogram_count = 1
			}
			cardValueToCompare = card.Value
		}
	} // end of for

	sort.Slice(histogram_counts, func(i, j int) bool {
		return histogram_counts[i] > histogram_counts[j]
	})

	// fmt.Printf("histogram_counts: %v", histogram_counts)

	switch {
	case histogram_counts[0] == 4: // quad 四条
		return QUAD
	case len(histogram_counts) >= 2 && histogram_counts[0] == 3 && histogram_counts[1] == 2: // boat 葫芦
		return BOAT
	case histogram_counts[0] == 3: // set 三条
		return SET
	case len(histogram_counts) >= 2 && histogram_counts[0] == 2 && histogram_counts[1] == 2: // two pair 两对
		// see which pair bigger which should be the first pair since its already sorted by Value
		return TWO_PAIR
	case histogram_counts[0] == 2: // one pair 对子
		return PAIR
	}

	if cards[0].Suit == cards[1].Suit &&
		cards[1].Suit == cards[2].Suit &&
		cards[2].Suit == cards[3].Suit &&
		cards[3].Suit == cards[4].Suit {
		isFlush = true // 同花
	}

	if cards[0].Value-cards[4].Value == 4 ||
		(cards[0].Value == 14 && cards[1].Value == 5) { //  top card is an ace and the 2nd to top card is a 5
		isStraight = true // 顺子
	}

	if isStraight && isFlush {
		if cards[0].Value == 14 && cards[1].Value == 13 {
			return ROYAL_STRAIGHT_FLUSH
		}
		return STRAIGHT_FLUSH // 同花顺
	} else if isStraight {
		return STRAIGHT
	} else if isFlush {
		return FLUSH
	} else {
		return HIGH_CARD
	}
}

func equalHands(hand1, hand2 []protocol.Card) bool {
	if len(hand1) == len(hand2) {
		numSameCards := 0

		for i := range hand1 {
			if hand1[i] == hand2[i] {
				numSameCards++
			}
		}

		if numSameCards == len(hand2) {
			return true
		}
	}
	return false
}

func combinationUtil(allCards []protocol.Card, potentialHands [][]protocol.Card,
	data map[int]protocol.Card, start, end, index, r int) [][]protocol.Card { // where r = Size of a combination
	if index == r {
		potentialHand := []protocol.Card{}
		for _, val := range data {
			potentialHand = append(potentialHand, val)
		}
		potentialHands = append(potentialHands, potentialHand)
		return potentialHands
	}

	for i := start; i <= end && end-i+1 >= r-index; i++ {
		data[index] = allCards[i]
		potentialHands = combinationUtil(allCards, potentialHands, data, i+1, end, index+1, r)
	}

	return potentialHands
}
