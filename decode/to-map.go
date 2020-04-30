package decode

import (
	"fmt"
)

// ToMap converts an interface{} to a map[string] of interfaces{}
func ToMap(raw interface{}) (map[string]interface{}, error) {
	if raw == nil {
		return nil, nil
	}
	var mapString map[string]interface{}
	switch raw.(type) {
	case map[string]interface{}:
		mapString = raw.(map[string]interface{})
	case map[interface{}]interface{}:
		mapIface := raw.(map[interface{}]interface{})
		mapString = make(map[string]interface{}, len(mapIface))
		for k, v := range mapIface {
			mapString[k.(string)] = v
		}
	default:
		return nil, fmt.Errorf("unexpected argument type: %T", raw)
	}
	return mapString, nil
}
