package xjsonnet

import (
	"os"

	"github.com/devthejo/snip/tools"
)

func Render(src string) (string, error) {
	envMap := tools.EnvToMap(os.Environ())
	return RenderEnv(src, envMap)
}
