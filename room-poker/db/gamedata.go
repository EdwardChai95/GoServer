package db

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

var logInformations = []map[string]string{}

// update the game coin of a uid
// e.g. uid 123 gamecoin += 10000
func UpdateGameCoinByUid(uid, updateAmt int64, transaction_comment string) {
	if updateAmt != 0 {
		_, err := db.Exec("update `user` set `game_coin` = `game_coin` + ? where `uid` = ?",
			updateAmt, uid)
		if err != nil {
			logger.Println(err)
		}
		_, err2 := db.Exec("INSERT INTO `game_coin_transaction`(`uid`,"+
			" `value`, `type`, `comment`, `datetime`) VALUES (?, ?, ?, ?, ?)", uid,
			updateAmt, "游戏币", transaction_comment, getCurrentShanghaiTimeString())
		if err2 != nil {
			logger.Println(err2)
		}
	}
}

// get the gamecoin of a uid
func GetGameCoinByUid(uid int64) int64 {
	users, err := db.QueryString("select * from user where uid = '" + strconv.Itoa(int(uid)) + "' LIMIT 1")
	if err != nil {
		log.Println(err)
	}
	if len(users) > 0 {
		gamecoin, err := strconv.ParseInt(users[0]["game_coin"], 10, 64)
		if err != nil {
			log.Println(err)
			return -1
		}
		return gamecoin
	}
	return -1
}

// log for whenever you ADD or SUBTRACT game coin
// e.g. + 100, - 200 etc
func NewGameCoinTransaction(uid int64, value int64) {
	if value != 0 {
		affected, err := db.Exec("INSERT INTO `game_coin_transaction`(`uid`, `value`, `type`, `comment`, `datetime`) "+
			"VALUES (?, ?, ?, ?, ?)",
			strconv.Itoa(int(uid)),
			strconv.Itoa(int(value)),
			"游戏币",
			"德州",
			getCurrentShanghaiTimeString())
		if err != nil {
			log.Println(err)
		}
		log.Infof("NewGameCoinTransaction: &v", affected)
	}
}

func NewLogInformation(logInformation map[string]string) {
	logInformations = append(logInformations, logInformation)
}

// log information of a single round
func InsertAllLogInformations(gameLogInformation map[string]string) {
	var wg sync.WaitGroup
	paramsInt := InsertLogInformation(gameLogInformation)
	params := strconv.FormatInt(paramsInt, 10)

	if paramsInt == -1 {
		return
	}

	for _, logInformation := range logInformations {
		logInformation["other_info"] = logInformation["other_info"] + " [参数：" + params + "]"
		logInformation["params"] = params
		wg.Add(1)
		go func(logInformation map[string]string) {
			InsertLogInformation(logInformation)
			wg.Done()
		}(logInformation)
	}

	wg.Wait()
	logInformations = []map[string]string{}
}

// function accepts a map where key is the name of the *column* in table of log_information
func InsertLogInformation(data map[string]string) int64 {
	cols := ""
	vals := ""

	for col, val := range data {
		cols += fmt.Sprintf("`%v`, ", col)
		vals += fmt.Sprintf("'%v', ", val)
	}

	cols += "`operating_time`"
	vals += fmt.Sprintf("'%v'", getCurrentShanghaiTimeString())

	sql := fmt.Sprintf("INSERT INTO `log_information` (%v) VALUES (%v)", cols, vals)
	// logger.Printf("sql: %v", sql)
	affected, err := db.Exec(sql)

	if err != nil {
		logger.Warn(err)
		return -1
	}

	if id, err := affected.LastInsertId(); err == nil {
		return id
	}

	return -1
}

// helper functions below:

func StringToInt(str string) int {
	i, err := strconv.Atoi(str)
	if err != nil {
		fmt.Printf("%d of type %T", i, i)
		fmt.Printf("StringToInt err %v", err)
	}
	return i
}

func StringToInt64(str string) int64 {
	n, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		fmt.Printf("%d of type %T", n, n)
		fmt.Printf("StringToInt64 err %v", err)
	}
	return n
}

func Int64ToString(n int64) string {
	return strconv.FormatInt(n, 10)
}

func getCurrentShanghaiTime() time.Time {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	return time.Now().In(loc)
}

func getCurrentShanghaiTimeString() string {
	createdFormat := "2006-01-02 15:04:05"
	return getCurrentShanghaiTime().Format(createdFormat)
}

var hmacSampleSecret = []byte("M45IFQ7QhG") // for jwt

const tokenuidkey = "uid" // for jwt

func verifyJWTString(tokenString string) (jwt.MapClaims, bool) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return hmacSampleSecret, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// fmt.Println(claims[TOKENUIDKEY])
		return claims, true //fmt.Sprintf("%v", claims[TOKENUIDKEY]), true
	} else {
		fmt.Println(err)
		return nil, false
	}
}

func VerifyJWT(tokenString string) (string, bool) {
	if tokenString == "" {
		return "", false
	}
	claims, isValid := verifyJWTString(tokenString)
	if isValid {
		return fmt.Sprintf("%v", claims[tokenuidkey]), true
	}
	return "", false
}
