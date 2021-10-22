package api

import (
	"errors"
	"net/http"

	// "time"

	// "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/lonng/nex"
	"gitlab.com/wolfplus/gamespace-lobby/db"
	"gitlab.com/wolfplus/gamespace-lobby/helper"
	// "gitlab.com/wolfplus/gamespace-lobby/db/model"
	// "gitlab.com/wolfplus/gamespace-lobby/define"
	// "gitlab.com/wolfplus/gamespace-lobby/errutil"
)

func MakeTempAnnouncementService() http.Handler {
	router := mux.NewRouter()

	router.Handle("/tempannouncement/gettempannouncement", nex.Handler(getTempAnnouncement)).Methods("POST")
	router.Handle("/tempannouncement/getannouncement", nex.Handler(getAnnouncement)).Methods("POST")
	return router
}

func getTempAnnouncement(r *http.Request) ([]map[string]string, error) {
	_, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}
	return db.GetTempAnnouncement(), nil
}

func getAnnouncement(r *http.Request) (map[string]interface{}, error) { // value为""即是隐藏
	_, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}
	var text1, text2 string
	text1 = `Nội dung:
	Đăng nhập nhận Code: 666666/888888
	Nhập code nhận ngay 10000 Xu và 66 Loa. Thêm Zalo CSKH: 0569894904 để nhận thêm nhiều ưu đãi hấp dẫn khác; mời bạn cùng tham gia nào!!!
	Code 666666: Loa*66
	Code 888888: Xu*10000`
	text2 = `Nội dung:
		Thêm Zalo CSKH, trong group Zalo sẽ gửi code bất kỳ để nhận thưởng ưu đãi
		Zalo CSKH: 0569894904
		Thời gian: 8:30 - 17:30
		Nội dung:
		Thêm Zalo CSKH, trong group Zalo sẽ gửi code bất kỳ để nhận thưởng ưu đãi
		Zalo CSKH: 0569894904
		Thời gian: 8:30 - 17:30`

	data := map[string]interface{}{
		"title1":  "Đăng nhập để nhận code tân thủ",
		"title2":  "Thêm Zalo CSKH nhận code thường ưu đãi",
		"title3":  "", //"测试",
		"title4":  "",
		"title5":  "",
		"title6":  "",
		"title7":  "Nạp càng nhiều, thưởng càng nhiều",
		"title8":  "ĐIỀU KHOẢN VÀ CHÍNH SÁCH",
		"announ1": text1,
		"announ2": text2,
		"announ3": "", //"公告，是指政府、团体对重大事件当众正式公布或者公开宣告，宣布。国务院2012年4月16日发布、2012年7月1日起施行的《党政机关公文处理工作条例》，对公告的使用表述为：“适用于向国内外宣布重要事项或者法定事项”。其中包含两方面的内容：一是向国内外宣布重要事项，公布依据政策、法令采取的重大行动等；二是向国内外宣布法定事项，公布依据法律规定告知国内外的有关重要规定和重大行动等。公告，是指政府、团体对重大事件当众正式公布或者公开宣告，宣布。国务院2012年4月16日发布、2012年7月1日起施行的《党政机关公文处理工作条例》，对公告的使用表述为：“适用于向国内外宣布重要事项或者法定事项”。其中包含两方面的内容：一是向国内外宣布重要事项，公布依据政策、法令采取的重大行动等；二是向国内外宣布法定事项，公布依据法律规定告知国内外的有关重要规定和重大行动等。",
		"announ4": "",
		"announ5": "",
		"announ6": "",
		"announ7": "Mỗi ngày nạp tiền thành công 01 lần, có thể liên hệ CSKH đăng ký thưởng 20% tiền nạp thành tiền mặt, có thể nạp nhiều lần để đăng ký nhận thưởng không giới hạn!",
		"announ8": "Vui lòng nhấp vào URL sau để xem thỏa thuận người dùng và chính sách bảo mật:",
	}

	return data, nil
}
