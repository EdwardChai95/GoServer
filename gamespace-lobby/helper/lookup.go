package helper

import (
	"encoding/csv"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"gitlab.com/wolfplus/gamespace-lobby/define"
)

var (
	sensitiveList = [][]string{}
	ipv4_list     = [][]string{}
	ipv6_list     = [][]string{}
)

func IsVerifiedIP(httpRequest *http.Request) bool {

	if len(ipv4_list) == 0 {
		ipv4_list = readCsv("lookup/ip2lite_v4.csv")
	}

	if len(ipv6_list) == 0 {
		ipv6_list = readCsv("lookup/ip2lite_v6.csv")
	}

	ip, _, err := net.SplitHostPort(httpRequest.RemoteAddr)
	if err != nil {
		log.Println(err)
	}

	userIP := net.ParseIP(ip)

	if userIP != nil && strings.Contains(ip, ":") {
		// ipv6
		s := strings.Split(ip, ":")
		ipnumber := (65536^7)*hexaNumberToInteger(s[0]) + (65536^6)*hexaNumberToInteger(s[1]) + (65536^5)*hexaNumberToInteger(s[2]) +
			(65536^4)*hexaNumberToInteger(s[3]) + (65536^3)*hexaNumberToInteger(s[4]) + (65536^2)*hexaNumberToInteger(s[5]) +
			65536*hexaNumberToInteger(s[6]) + hexaNumberToInteger(s[7])

		for _, line := range ipv6_list {
			if StringToInt64(line[0]) <= ipnumber && StringToInt64(line[1]) >= ipnumber {
				// log.Println(line[2])
				// check code whether need to block
				if ArrayContainsString(define.BLOCKED_IPCODES, line[2]) {
					return false
				}
				break
			}
		}
	} else {
		s := strings.Split(ip, ".")
		ipnumber := 16777216*StringToInt64(s[0]) + 65536*StringToInt64(s[1]) + 256*StringToInt64(s[2]) + StringToInt64(s[3])

		// log.Println(ipnumber)

		for _, line := range ipv4_list {
			if StringToInt64(line[0]) <= ipnumber && StringToInt64(line[1]) >= ipnumber {
				// log.Println(line[2])
				// check code whether need to block
				if ArrayContainsString(define.BLOCKED_IPCODES, line[2]) {
					return false
				}
				break
			}
		}
	}

	return true
}

func CheckForSensitiveWords(s string) string {
	if len(sensitiveList) == 0 {
		sensitiveList = readCsv("lookup/sensitive_words.csv")
	}

	for _, line := range sensitiveList {
		s = strings.ReplaceAll(s, line[0], "***")
	}
	return s
}

func readCsv(filename string) [][]string {
	// Open CSV file
	f, err := os.Open(filename)
	if err != nil {
		return [][]string{}
	}
	defer f.Close()
	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return [][]string{}
	}
	return lines
}

