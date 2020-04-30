package decode

import (
	"fmt"
	"strconv"
	"time"
)

func Duration(durationInterface interface{}) (time.Duration, error) {
	var durationString string
	var durationTimeDuration time.Duration
	var err error
	switch durationInterface.(type) {
	case float64:
		durationString = fmt.Sprintf("%f", durationInterface.(float64)) + "s"
	case int:
		durationString = strconv.Itoa(durationInterface.(int)) + "s"
	case string:
		durationString = durationInterface.(string)
		if _, err := strconv.Atoi(durationString); err == nil {
			durationString = durationString + "s"
		}
	case nil:
		return durationTimeDuration, nil
	default:
		return durationTimeDuration, fmt.Errorf(`invalid duration type:"%T", value:"%v"`+"\n", durationInterface, durationInterface)
	}
	durationTimeDuration, err = time.ParseDuration(durationString)
	return durationTimeDuration, err
}
