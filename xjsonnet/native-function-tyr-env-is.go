package xjsonnet

import (
	"strings"

	jsonnet "github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
	"gitlab.com/youtopia.earth/ops/snip/tools"
)

func NativeFunctionK1YEnvIs(envMap map[string]string) *jsonnet.NativeFunction {
	var nativeFunctionEnv = &jsonnet.NativeFunction{
		Name:   "snipEnvIs",
		Params: ast.Identifiers{"key"},
		Func: func(arguments []interface{}) (interface{}, error) {
			envName := arguments[0].(string)
			envKey := "K1Y_ENV"
			envVal := envMap[envKey]
			envList := strings.Split(envVal, ",")
			if len(envList) == 0 {
				return false, nil
			}
			is := tools.SliceContainsString(envList, envName)
			return is, nil
		},
	}
	return nativeFunctionEnv
}
