package xjsonnet

import (
	jsonnet "github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

func NativeFunctionEnv(envMap map[string]string) *jsonnet.NativeFunction {
	var nativeFunctionEnv = &jsonnet.NativeFunction{
		Name:   "env",
		Params: ast.Identifiers{"key"},
		Func: func(arguments []interface{}) (interface{}, error) {
			key := arguments[0].(string)
			if value, hasKey := envMap[key]; hasKey {
				return value, nil
			}
			return "", nil
		},
	}
	return nativeFunctionEnv
}
