package decode

import (
	"fmt"
)

// ToStrings converts the given interface{} to a []string, or returns an error.
// In the case where the argument is a string already, will wrap the arg in
// a slice.
func ToStrings(raw interface{}) ([]string, error) {
	if raw == nil {
		return nil, nil
	}
	switch t := raw.(type) {
	case string:
		return []string{t}, nil
	case []string:
		return t, nil
	case []interface{}:
		return interfaceToStringArray(t), nil
	default:
		return nil, fmt.Errorf("unexpected argument type: %T", t)
	}
}

func interfaceToString(raw interface{}) string {
	switch t := raw.(type) {
	case string:
		return t
	default:
		return fmt.Sprintf("%v", t)
	}
}

func interfaceToStringArray(rawArray []interface{}) []string {
	if rawArray == nil || len(rawArray) == 0 {
		return nil
	}
	var stringArray []string
	for _, raw := range rawArray {
		stringArray = append(stringArray, interfaceToString(raw))
	}
	return stringArray
}
