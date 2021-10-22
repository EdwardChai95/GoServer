package api

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"unsafe"

	"github.com/gorilla/mux"
	"github.com/lonng/nex"
	"gitlab.com/wolfplus/gamespace-lobby/db"
	"gitlab.com/wolfplus/gamespace-lobby/helper"
	proto "gitlab.com/wolfplus/gamespace-lobby/proto"
)

func MakePhoneService() http.Handler {
	router := mux.NewRouter()
	router.Handle("/phone/verificationCode", nex.Handler(vCode)).Methods("POST")

	return router
}

//处理 前端 发送验证码 POST请求
func vCode(r *http.Request, req *proto.VCodeReq) (*proto.VCodeRes, error) {
	_, isValid := helper.VerifyJWT(r)
	if !isValid {
		return nil, errors.New("Invalid token")
	}
	vCode := generateVcode(6)

	accesskey := "602200c2b81e61422665a65b"
	secretkey := "c7f3a228bdcc4966b484794b4591bf61"

	random := generateVcode(6)
	wholeurl := "https://live.moduyun.com/sms/v1/sendsinglesms?accesskey=" + accesskey +
		"&random=" + random
	currTime := helper.GetCurrentShanghaiTime().Unix()

	//请登录zz.253.com获取API账号、密码以及短信发送的URL
	params := make(map[string]interface{})
	tel := map[string]string{
		"nationcode": "84",
		"mobile":     req.PhoneNumber,
	}
	params["tel"] = tel
	params["type"] = 0
	// params["msg"] = url.QueryEscape("【北京易思迈】亲爱的用户，您的短信验证码为 " + vCode + ",5分钟内有效，若非本人操作请忽略。")
	//params["msg"] = "【自云咨讯】尊敬的用户：您的验证码" + vCode + "工作人员不会索取，请勿泄漏。"
	params["msg"] = "Your verification code is:" + vCode //+ ""

	h := sha256.New()
	h.Write([]byte("secretkey=" + secretkey + "&random=" + random +
		"&time=" + strconv.Itoa(int(currTime)) + "&mobile=" + req.PhoneNumber))
	params["sig"] = hex.EncodeToString(h.Sum(nil))

	params["time"] = currTime
	params["extend"] = ""
	params["ext"] = ""

	bytesData, err := json.Marshal(params)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	/* TEST */

	// fmt.Println((params))
	// fmt.Println(string(bytesData))
	/* TEST */

	reader := bytes.NewBuffer(bytesData)
	request, err := http.NewRequest("POST", wholeurl, reader)

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{Transport: tr}
	resp, err := client.Do(request)

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	str := (*string)(unsafe.Pointer(&respBytes))
	fmt.Println(*str)

	logger.Infof("the return vCode and phone number is " + vCode + " phone " + req.PhoneNumber)

	db.CreateVCode(req.PhoneNumber, vCode)

	res := &proto.VCodeRes{
		PhoneNumber: req.PhoneNumber,
		VCode:       vCode,
	}

	return res, nil
}

// func vCodeService(request *http.Request ) <-(chan *proto.VCodeRes, chan error){
// 	r := make(chan int32)

// 	client := http.Client{}
// 	resp, err := client.Do(request)
// }

// 辅助函数
func generateVcode(max int) string {
	b := make([]byte, max)
	n, err := io.ReadAtLeast(rand.Reader, b, max)
	if n != max {
		panic(err)
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b)
}

var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
