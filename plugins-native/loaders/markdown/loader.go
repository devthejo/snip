package mainNative

import (
	"strings"

	"gitlab.com/youtopia.earth/ops/snip/plugin/loader"
)

var (
	Loader = loader.Plugin{
		Check: func(command []string) bool {
			return strings.HasSuffix(command[0], ".md")
		},
		Load: func(cfg *loader.Config) error {
			return BuildBash(cfg)
		},
	}
)
