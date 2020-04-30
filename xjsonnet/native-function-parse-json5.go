package xjsonnet

import (
	jsonnet "github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
	"github.com/yosuke-furukawa/json5/encoding/json5"
)

func NativeFunctionParseJson5() *jsonnet.NativeFunction {
	var nativeFunctionParseJson5 = &jsonnet.NativeFunction{
		Name:   "parseJson5",
		Params: ast.Identifiers{"x"},
		Func: func(x []interface{}) (interface{}, error) {

			str := x[0].(string)

			var json interface{}
			if err := json5.Unmarshal([]byte(str), &json); err != nil {
				return nil, err
			}
			return json, nil
		},
	}
	return nativeFunctionParseJson5
}
