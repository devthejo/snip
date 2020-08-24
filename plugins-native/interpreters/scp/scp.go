package mainNative

import (
	"gitlab.com/youtopia.earth/ops/snip/plugins-native/loaders/markdown/interpreter"
)

var (
	Interpreter = interpreter.Plugin{
		Interpret: func(cfg *interpreter.Config) (bool, error) {

			return true, nil
		},
	}
)
