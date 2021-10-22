package controllers

import (
	"admin/db"
	"admin/helper"
	"admin/template"
	"fmt"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/spf13/viper"
)

func (c *controllers) ReportTopup_get(ctx iris.Context) { // GET /report/topup
	//pageNumber := ctx.URLParamDefault("pageNo", "1")
	dateStart := ctx.URLParamDefault("dateStart", "")
	dateEnd := ctx.URLParamDefault("dateEnd", "")
	uid := ctx.URLParamDefault("uid", "")
	topupReportType := ctx.URLParamDefault("topupReportType", "")

	searchParams := map[string]string{
		"dateStart":       strings.TrimSpace(dateStart),
		"dateEnd":         strings.TrimSpace(dateEnd),
		"uid":             strings.TrimSpace(uid),
		"topupReportType": strings.TrimSpace(topupReportType),
	}

	urlParams := ""
	for k, v := range searchParams {
		if strings.TrimSpace(v) != "" {
			urlParams += "&" + k + "=" + v
		}
	}

	/*data := db.GetTopupReport(pageNumber,urlParams,searchParams)
	if data == nil {
		ctx.Redirect("/report/topup", iris.StatusTemporaryRedirect)
		return
	}

	for k, v := range data {
		ctx.ViewData(k, v)
	}*/

	report := db.GetTopupReport(searchParams)
	report_header := []string{
		"ID",
		"注册时间",
		"充值时间",
		"充值金额",
		"充值游戏币",
		"充值状态",
		"是否首充",
	}

	report_data := [][]string{}
	for _, r := range report {
		data := []string{
			r["uid"],
			r["create_at"],
			r["updated_datetime"],
			r["payment_amount"],
			r["game_coin_amount"],
			r["order_status"],
			r["isFirst"],
		}
		report_data = append(report_data, data)
	}

	// topupReportType
	for _, item := range db.TopupReportTypes {
		if fmt.Sprintf("%v", item["val"]) == topupReportType {
			item["selected"] = true
		}
	}
	ctx.ViewData("topupReportType", &template.Params{Name: "topupReportType",
		Title: "充值状态", Options: db.TopupReportTypes})

	// form
	//ctx.ViewData("pagination", helper.PaginationHTML("/report/topup/", urlParams, helper.StringToInt(pageNumber), data["NumPages"].(int)))

	ctx.ViewData("dateStart", &template.Params{Name: "dateStart", Title: "操作时间开始", Value: dateStart})
	ctx.ViewData("dateEnd", &template.Params{Name: "dateEnd", Title: "操作时间结束", Value: dateEnd})
	ctx.ViewData("uid", &template.Params{Name: "uid", Title: "账号ID", Value: uid})

	// report data
	ctx.ViewData("report_header", report_header)
	ctx.ViewData("report_data", report_data)

	ctx.View("report/topup.html")

}

func (c *controllers) Report_get(ctx iris.Context) { // GET /report
	dateStart := ctx.URLParamDefault("dateStart", "")
	dateEnd := ctx.URLParamDefault("dateEnd", "")
	order_status := ctx.URLParamDefault("order_status", "")

	searchParams := map[string]string{
		"dateStart":    strings.TrimSpace(dateStart),
		"dateEnd":      strings.TrimSpace(dateEnd),
		"order_status": strings.TrimSpace(order_status),
	}

	urlParams := ""
	for k, v := range searchParams {
		if strings.TrimSpace(v) != "" {
			urlParams += "&" + k + "=" + v
		}
	}

	report := db.GetStatReport(searchParams)

	current_report := db.GetCurrentReport()
	for k, v := range current_report {
		ctx.ViewData(k, v)
	}

	current_active := db.GetCurrentActive()
	/*for k, v := range current_active {
	        ctx.ViewData(k, v)
	}*/

	current_game_online := db.GetCurrentGameOnline()
	for k, v := range current_game_online {
		ctx.ViewData(k, v)
	}

	//	current_pinshi_online := db.GetCurrentReportPinshiOnline()
	//	current_roulette_online := db.GetCurrentReportRouletteOnline()
	//        current_haochehui_online := db.GetCurrentReportHaochehuiOnline()
	//        current_fruit_online := db.GetCurrentReportFruitOnline()

	current_report_header := []string{
		"登陆人数",
		"注册人数",
		"激活人数",
		"充值金额",
		"总下注",
		"总局数",
		"总输赢",
		"拼十人数",
		"动物人数",
		"豪车人数",
		"水果人数",
	}

	report_header := []string{
		"日期",
		"登录人数",
		"充值金额",
		"充值游戏币",
		"充值数量",
		"注册人数",
		"激活",
		"玩家总游戏币",
		"玩家总下注",
		"系统赠送游戏币",
		"管理员增加游戏币",
		"管理员减少游戏币",
		"水果登陆",
		"豪车汇登陆",
		"动物乐园登陆",
		"拼十登陆",
		"水果下注",
		"豪车汇下注",
		"动物乐园下注",
		"拼十下注",
		"水果输赢",
		"豪车汇输赢",
		"豪车汇税收",
		"动物乐园输赢",
		"拼十输赢",
		"拼十税收",
		"总输赢",
		"游戏总人数",
		"游戏总局数",
		"总税收",
	}

	current_active_num := [][]string{}
	for _, r := range current_active {
		data := []string{
			r["active_num"],
		}
		current_active_num = append(current_active_num, data)
	}

	/*	current_pinshi := [][]string{}
		for _, r := range current_pinshi_online{
		        data := []string{
		                r["pinshi_online"],
		        }
		        current_pinshi = append(current_pinshi, data)
		}
	*/
	/*	current_roulette := [][]string{}
		for _, r := range current_roulette_online{
		        data := []string{
		                r["roulette_online"],
		        }
		        current_roulette = append(current_roulette, data)
		}

		current_haochehui := [][]string{}
		for _, r := range current_haochehui_online{
		        data := []string{
		                r["haochehui_online"],
		        }
		        current_haochehui = append(current_haochehui, data)
		}

		current_fruit := [][]string{}
		for _, r := range current_fruit_online{
		        data := []string{
		                r["fruit_online"],
		        }
		        current_fruit = append(current_fruit, data)
		}

	*/
	report_data := [][]string{}
	for _, r := range report {
		data := []string{
			r["create_at"],
			r["login_num"],
			r["payment_amount"],
			r["game_coin_amount"],
			r["top_up_num"],
			r["register_num"],
			r["is_active"],
			r["totalPlayerGamecoin"],
			r["total_player_bet"],
			r["system_game_coin"],
			r["admin_add_game_coin"],
			r["admin_minus_game_coin"],
			r["fruit_total_login"],
			r["haochehui_total_login"],
			r["roulette_total_login"],
			r["pinshi_total_login"],
			r["fruit_total_bet"],
			r["haochehui_total_bet"],
			r["roulette_total_bet"],
			r["pinshi_total_bet"],
			r["fruit_total_win_lose"],
			r["haochehui_total_win_lose"],
			r["haochehui_tax"],
			r["roulette_total_win_lose"],
			r["pinshi_total_win_lose"],
			r["pinshi_tax"],
			r["total_win_lose"],
			r["total_player"],
			r["total_board"],
			r["total_tax"],
		}
		report_data = append(report_data, data)
	}

	ctx.ViewData("jwt", sess.Get(helper.SESSION_USER).(string))
	ctx.ViewData("wsurl", "ws://"+viper.GetString("webserver.addr")+":12307/coinws/net")

	// form
	ctx.ViewData("dateStart", &template.Params{Name: "dateStart", Title: "操作时间开始", Value: dateStart})
	ctx.ViewData("dateEnd", &template.Params{Name: "dateEnd", Title: "操作时间结束", Value: dateEnd})
	ctx.ViewData("order_status", &template.Params{Name: "order_status", OnChangeSubmit: true,
		Options: db.GetAllOrderStatus(order_status)})

	// reports
	ctx.ViewData("current_report_header", current_report_header)
	ctx.ViewData("current_active_num", current_active_num)

	//        ctx.ViewData("current_pinshi_online", current_pinshi_online)
	//	ctx.ViewData("current_roulette_online", current_roulette_online)
	//       ctx.ViewData("current_haochehui_online", current_haochehui_online)
	//        ctx.ViewData("current_fruit_online", current_fruit_online)

	ctx.ViewData("report_header", report_header)
	ctx.ViewData("report_data", report_data)

	ctx.View("report/index.html")
}
