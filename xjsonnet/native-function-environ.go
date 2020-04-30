package xjsonnet

import (
	jsonnet "github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

func NativeFunctionEnviron(envMap map[string]string) *jsonnet.NativeFunction {
	var NativeFunctionEnviron = &jsonnet.NativeFunction{
		Name:   "environ",
		Params: ast.Identifiers{},
		Func: func(arguments []interface{}) (interface{}, error) {
			envMapI := make(map[string]interface{}, len(envMap))
			for k, v := range envMap {
				envMapI[k] = v
			}
			return envMapI, nil
		},
	}
	return NativeFunctionEnviron
}
