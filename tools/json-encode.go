package tools

import (
	"log"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func JsonEncode(value interface{}) string {
	bs, err := json.Marshal(&value)
	if err != nil {
		log.Fatalf(`unable to json marshal "%v"`, value)
	}
	valueStr := string(bs)
	return valueStr
}
