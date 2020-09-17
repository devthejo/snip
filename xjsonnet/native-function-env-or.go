package xjsonnet

import (
	jsonnet "github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

func NativeFunctionEnvOr(envMap map[string]string) *jsonnet.NativeFunction {
	var nativeFunctionEnvOr = &jsonnet.NativeFunction{
		Name:   "envOr",
		Params: ast.Identifiers{"key","defaultValue"},
		Func: func(arguments []interface{}) (interface{}, error) {
			key := arguments[0].(string)
			if value, hasKey := envMap[key]; hasKey {
				return value, nil
			}
			if len(arguments) > 1 {
				defaultValue := arguments[1].(string)
				return defaultValue, nil
			}
			return "", nil
		},
	}
	return nativeFunctionEnvOr
}
