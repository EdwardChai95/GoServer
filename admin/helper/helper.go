package helper

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pborman/uuid"
)

var RECORDS_PER_PAGE = 50

var PASSWORDSEPERATOR = ":::"
var SESSION_USER = "User"
var SESSION_TEMPMESSAGE = "tempMessage"
var SESSION_TEMPFAILMESSAGE = "tempFailMessage"
var SESSION_ISLOGGEDIN = "isLoggedIn"
var SESSION_ISADMINLOGGEDIN = "isAdminLoggedIn"

var hmacSampleSecret = []byte("M45IFQ7QhG")
var tokenUidKey = "uid"
var tokenAdminUidKey = "adminuid"

func VerifyJWTString(tokenString string) (string, bool) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return hmacSampleSecret, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// fmt.Println(claims[tokenUidKey])
		return fmt.Sprintf("%v", claims[tokenUidKey]), true
	} else {
		fmt.Println(err)
		return "", false
	}
}

func NewJWT(uid string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		tokenUidKey:      uid,
		tokenAdminUidKey: uid,
		// "nbf": time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
	})
	tokenString, err := token.SignedString(hmacSampleSecret)

	if err != nil {
		fmt.Println(tokenString, err)
	}

	return tokenString
}

func StringToInt(str string) int {
	i, err := strconv.Atoi(str)
	if err != nil {
		fmt.Println(err)
	}
	return i
}

func StringToInt64(str string) int64 {
	n, err := strconv.ParseInt(str, 10, 64)
	if err == nil {
		fmt.Printf("%d of type %T", n, n)
	}
	return n
}

func Int64ToString(n int64) string {
	return strconv.FormatInt(n, 10)
}

func passwordHash(pwd, salt string) string {
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "%s%x%s", salt, pwd, salt)
	result1 := sha1.Sum(buf.Bytes())

	buf.Reset()

	fmt.Fprintf(buf, "%s%s%x%s%s", pwd, salt, result1, salt, pwd)
	result2 := sha1.Sum(buf.Bytes())

	buf.Reset()
	fmt.Fprintf(buf, "%x", result2)
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

// PasswordHash accept password and generate with uuid as salt
// FORMAT: sha1.Sum(pwd + salt + sha1.Sum(salt + pwd + salt) + salt + pwd)
func PasswordHash(pwd string) (hash, salt string) {
	salt = strings.Replace(uuid.New(), "-", "", -1)
	hash = passwordHash(pwd, salt)
	return hash, salt
}

func VerifyPassword(pwd, salt, hash string) bool {
	return passwordHash(pwd, salt) == hash
}

func GetCurrentShanghaiTime() time.Time {
	loc, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	// fmt.Println(time.Now().Add(time.Hour * time.Duration(8)).Unix())
	return time.Now().In(loc) //.Add(time.Hour * time.Duration(8)).Unix()
}

func GetCurrentShanghaiTimeString() string {
	createdFormat := "2006-01-02 15:04:05"
	return GetCurrentShanghaiTime().Format(createdFormat)
	// return strconv.Itoa(int(GetCurrentShanghaiTimeUnix()))
}

func SQLUpdateDataStr(data map[string][]string) string {
	i := 0
	str := ""
	for k, v := range data {
		if len(v) > 1 && v[1] == "toIgnore" {
			delete(data, k) // delete
		} else if len(v) > 1 && v[1] == "isPassword" {
			if v[0] != "" { // if empty dont update means remain same password
				hash, salt := PasswordHash(v[0])
				v[0] = hash + PASSWORDSEPERATOR + salt
			} else {
				delete(data, k) // delete
			}
		}
	}

	for k, v := range data {
		i++
		str += "`" + k + "`='" + v[0] + "'"

		if i != len(data) {
			str += ", "
		} else {
			str += " "
		}
	}
	return str
}

func MapStringToMapInterface(m []map[string]string) []map[string]interface{} {
	out := []map[string]interface{}{}
	for _, mss := range m {
		m2 := make(map[string]interface{}, len(mss))
		for k, v := range mss {
			m2[k] = v
		}
		out = append(out, m2)
	}
	return out
}
