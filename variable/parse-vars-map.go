package variable

import (
	"strconv"
	"strings"

	"github.com/devthejo/snip/decode"
	"github.com/devthejo/snip/errors"
)

func ParseVarsMap(varsI map[string]interface{}, depth int) map[string]*Var {
	vars := make(map[string]*Var)
	for key, val := range varsI {
		var value map[string]interface{}
		switch v := val.(type) {
		case map[string]interface{}:
			value = v
		case map[interface{}]interface{}:
			var err error
			value, err = decode.ToMap(v)
			errors.Check(err)
		case string:
			value = make(map[string]interface{})
			value["value"] = v
		case int:
			value = make(map[string]interface{})
			value["value"] = strconv.Itoa(v)
		case bool:
			value = make(map[string]interface{})
			if v {
				value["value"] = "true"
			} else {
				value["value"] = "false"
			}
		case nil:
			value = make(map[string]interface{})
			value["value"] = ""
		default:
			UnexpectedTypeVarValue(key, v)
		}
		vr := &Var{
			Depth: depth,
		}
		key = strings.ToUpper(key)
		vr.Parse(key, value)
		vars[key] = vr
	}
	return vars
}
