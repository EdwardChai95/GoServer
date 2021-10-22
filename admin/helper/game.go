package helper

import (
	"encoding/json"
	"html/template"
)

var (
	fruitBetZones = []string{"bar", "77", "双星", "西瓜", "铃铛", "芒果", "橘子", "苹果"}

	haocheBetzones = []string{"保大", "宝大", "奔大", "众大", "保小", "宝小", "奔小", "众小"}

	rouletteColors     = map[string]string{"red": "红", "yellow": "黄", "green": "绿"}
	rouletteTexts      = map[string]string{"red": "庄", "yellow": "闲", "green": "和"}
	rouletteTexts2     = map[string]string{"zhuang": "庄", "xian": "闲", "he": "和"}
	rouletteAnimals    = []string{"狮", "猫", "猴", "兔"}
	rouletteAnimalsMap = map[string]string{"lion": "狮", "panda": "熊", "monkey": "猴", "rabit": "兔"}
)

// game specific helper
func GetSuit(suitKey string) string {
	var suits = map[string]string{
		"s": "&#x02660;",
		"h": "<span style=\"color:red;\">&#x02665;</span>",
		"c": "&#x02663;",
		"d": "<span style=\"color:red;\">&#x02666;</span>",
	}
	return suits[suitKey]
}

func GetGambler(gambler map[string]json.RawMessage) template.HTML {
	var title = JsonObjToStr(gambler, "title")
	var htmlstring = title + "<br/>"
	var orderedCards = JsonObjToArr(gambler, "orderedCardValues")
	var winLose = "赢"
	if !JsonObjToBool(gambler, "isWin") {
		winLose = "输"
	}

	for _, card := range orderedCards {
		htmlstring += GetSuit(JsonObjToStr(card, "suit"))
		htmlstring += JsonObjToStr(card, "face")
		htmlstring += " "
	}
	htmlstring += "<br/>"
	if title != "banker" {
		htmlstring += "总押注：" + Int64ToString(JsonObjToInt(gambler, "totalbetting"))
		htmlstring += "输赢：" + winLose + "<br/>"
	}

	return template.HTML(htmlstring)
}

func CustomPlayerLog(v map[string]interface{}) map[string]interface{} {
	// roundInfo := JsonStrToMap(v["other_info"].(string))
	// i := StringToInt(v["result"].(string))
	if v["game"] == "水果" {
		// {"SpecialPrizeChinese":"普通","SelectedWinningItems":
		// [{"winningBetZone":2,"probability":1100,"odds":2,"image":"reel_slot_20","isBigFruit":false}]}

		// {"SpecialPrizeChinese":"九莲宝灯","SelectedWinningItems":
		// [{"winningBetZone":6,"probability":300,"odds":10,"image":"reel_slot_0","isBigFruit":true},
		// {"winningBetZone":4,"probability":150,"odds":20,"image":"reel_slot_1","isBigFruit":true},
		// ]}

		// betZone := fruitBetZones[i]
		// v["other_info1"] = "押注：" + betZone

		// v["other_info4"] = "奖项：" + JsonObjToStr(roundInfo, "SpecialPrizeChinese")

		// selectedWinningItems := JsonObjToArr(roundInfo, "SelectedWinningItems")

		// v["other_info5"] = "开奖："
		// if len(selectedWinningItems) == 0 {
		// 	v["other_info5"] = v["other_info5"].(string) + "-"
		// }
		// for _, winningItem := range selectedWinningItems {
		// 	index := int(JsonObjToInt(winningItem, "winningBetZone"))
		// 	if index != -1 {
		// 		odds := Int64ToString(JsonObjToInt(winningItem, "odds"))
		// 		probability := fmt.Sprintf("%.2f", float32(JsonObjToInt(winningItem, "probability"))/100)
		// 		v["other_info5"] = v["other_info5"].(string) + "<br/>" + fruitBetZones[index] + " 概率：" + odds + " 几率：" + probability
		// 	}

		// }
		// v["other_info5"] = template.HTML(v["other_info5"].(string))
	}

	if v["game"] == "豪车汇" {
		// betZone := haocheBetzones[i]
		// v["other_info1"] = "押注：" + betZone
	}

	if v["game"] == "动物乐园" {
		// v["other_info4"] = v["other_info"]
		// v["other_info"] = v["before"]
		// v["other_info2"] = v["used"]
		// v["other_info3"] = v["after"]
	}

	// v["other_info"] = ""

	return v
}

func CustomServerLog(v map[string]interface{}) map[string]interface{} {

	// bet_total := "[总押注：" + v["bet_total"].(string) + "， 总输赢：" + v["win_total"].(string) + "] "
	params1 := "[参数：" + v["log_information_id"].(string) + "]"

	if v["game"] == "豪车汇" {
		// i := StringToInt(v["result"].(string))
		// betZone := haocheBetzones[i]
		v["other_info"] = v["other_info"].(string) + params1
	}

	if v["game"] == "动物乐园" {
		// roundInfo := JsonStrToMap(v["other_info"].(string))

		// html, html2 := animalRoundInfo(roundInfo)

		// v["other_info"] = template.HTML(html + html2 + bet_total + params1)
	}
	return v
}

func htmltextShadow(str, color string) string {
	return "<span style=\"text-shadow: " + color + " 0px 0px 1px, " + color + " 0px 0px 1px, " + color + " 0px 0px 1px,#000 0px 0px 1px, " + color + " 0px 0px 1px, " + color + " 0px 0px 1px;-webkit-font-smoothing: antialiased;\">" + str + "</span>"
}

func animalRoundInfo(roundInfo map[string]json.RawMessage) (string, string) {
	winningBetZone := JsonAttributeToMap(roundInfo, "winningBetZone")
	winningTextBetZone := JsonAttributeToMap(roundInfo, "winningTextBetZone")

	if winningBetZone == nil || winningTextBetZone == nil {
		return "[-，", "-]"
	}

	betZoneColor := JsonObjToStr(winningBetZone, "Bg")
	textIcon := rouletteTexts2[JsonObjToStr(winningTextBetZone, "Icon")]
	textColor := betZoneColor
	for k, v := range rouletteTexts {
		if v == textIcon {
			textColor = k
		}
	}

	html := "开奖：[" + htmltextShadow(rouletteColors[betZoneColor]+rouletteAnimalsMap[JsonObjToStr(winningBetZone, "Icon")], betZoneColor) +
		"｜" + Int64ToString(JsonObjToInt(winningBetZone, "Odds")) + "，"

	html2 := htmltextShadow(textIcon, textColor) + "｜" + Int64ToString(JsonObjToInt(winningTextBetZone, "Odds")) + "] "

	return html, html2
}
