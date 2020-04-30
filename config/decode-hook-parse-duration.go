package config

import (
	"reflect"
	"strconv"
	"time"

	"github.com/mitchellh/mapstructure"
)

func DecodeHookParseDuration() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(time.Duration(5)) {
			return data, nil
		}

		raw := data.(string)
		if _, err := strconv.Atoi(raw); err == nil {
			raw = raw + "s"
		}

		return time.ParseDuration(raw)
	}
}
