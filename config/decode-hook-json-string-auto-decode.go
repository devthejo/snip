package config

import (
	"reflect"

	"github.com/yosuke-furukawa/json5/encoding/json5"
)

func DecodeHookJsonStringAutoDecode(m interface{}) func(rf reflect.Kind, rt reflect.Kind, data interface{}) (interface{}, error) {
	return func(rf reflect.Kind, rt reflect.Kind, data interface{}) (interface{}, error) {
		if rf != reflect.String ||
			rt == reflect.String ||
			rt == reflect.Int ||
			rt == reflect.Int8 ||
			rt == reflect.Int16 ||
			rt == reflect.Int32 ||
			rt == reflect.Int64 ||
			rt == reflect.Uint ||
			rt == reflect.Uint8 ||
			rt == reflect.Uint16 ||
			rt == reflect.Uint32 ||
			rt == reflect.Uint64 ||
			rt == reflect.Uintptr ||
			rt == reflect.Float32 ||
			rt == reflect.Float64 ||
			rt == reflect.Complex64 ||
			rt == reflect.Complex128 {
			return data, nil
		}

		raw := data.(string)
		if raw != "" && (raw[0:1] == "{" || raw[0:1] == "[") {
			err := json5.Unmarshal([]byte(raw), &m)
			return m, err
		}

		return data, nil
	}
}
