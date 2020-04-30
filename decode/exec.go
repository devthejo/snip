package decode

import (
	"fmt"
	"strings"
)

func Exec(execInterface interface{}) ([]string, error) {
	var execSlice []string
	switch execInterface.(type) {
	case string:
		execSlice = strings.Fields(execInterface.(string))
	case []string:
		execSlice = execInterface.([]string)
	case []interface{}:
		execSliceInterface := execInterface.([]interface{})
		l := len(execSliceInterface)
		if l > 0 {
			execSliceInterface := execInterface.([]interface{})
			execSlice = make([]string, l)
			for i, v := range execSliceInterface {
				execSlice[i] = fmt.Sprint(v)
			}
		}
	case nil:
	default:
		return execSlice, fmt.Errorf(`invalid exec type:"%T", value:"%v"`, execInterface, execInterface)
	}
	return execSlice, nil
}
