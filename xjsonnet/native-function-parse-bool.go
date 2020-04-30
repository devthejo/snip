package xjsonnet

import (
	"fmt"

	jsonnet "github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

func NativeFunctionParseBool() *jsonnet.NativeFunction {
	var nativeFunctionParseBool = &jsonnet.NativeFunction{
		Name:   "parseBool",
		Params: ast.Identifiers{"x"},
		Func: func(x []interface{}) (interface{}, error) {
			s := x[0].(string)
			if s == "true" || s == "1" {
				return true, nil
			} else if s == "false" || s == "0" || s == "" {
				return false, nil
			} else {
				return nil, fmt.Errorf("invalid boolean: %v", s)
			}
		},
	}
	return nativeFunctionParseBool
}
