package helper

import (
	"encoding/json"
	"log"
)

func JsonObjToBool(objmap map[string]json.RawMessage, key string) bool {
	var boolean1 bool
	err := json.Unmarshal(objmap[key], &boolean1)
	if err != nil {
		log.Print(err)
	}
	return boolean1
}

func JsonObjToInt(objmap map[string]json.RawMessage, key string) int64 {
	var str int64
	err := json.Unmarshal(objmap[key], &str)
	if err != nil {
		log.Print(err)
	}
	return str
}

func JsonObjToStr(objmap map[string]json.RawMessage, key string) string {
	var str string
	err := json.Unmarshal(objmap[key], &str)
	if err != nil {
		// log.Print(err)
		return ""
	}
	return str
}

func JsonObjToArr(objmap map[string]json.RawMessage, key string) []map[string]json.RawMessage {
	var arr []map[string]json.RawMessage
	err := json.Unmarshal(objmap[key], &arr)
	if err != nil {
		log.Print(err)
	}
	return arr
}

func JsonAttributeToMap(objmap map[string]json.RawMessage, key string) map[string]json.RawMessage {
	var m1 map[string]json.RawMessage
	err := json.Unmarshal(objmap[key], &m1)
	if err != nil {
		// log.Print(err)
		return nil
	}
	return m1
}

func JsonStrToMap(s string) map[string]json.RawMessage {
	data := []byte(s)
	var objmap map[string]json.RawMessage
	err := json.Unmarshal(data, &objmap)
	if err != nil {
		// log.Print(err)
		return nil
	}
	return objmap
}

func IsJSONString(s string) bool { // not working
	var js string
	return json.Unmarshal([]byte(s), &js) == nil
}
