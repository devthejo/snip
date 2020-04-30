package xjsonnet

import (
	"os"

	"gitlab.com/youtopia.earth/ops/snip/tools"
)

func Render(src string) (string, error) {
	envMap := tools.EnvToMap(os.Environ())
	return RenderEnv(src, envMap)
}
