package mainNative

import (
	"strings"

	"gitlab.com/youtopia.earth/ops/snip/plugin/loader"
)

var (
	Loader = loader.Plugin{
		Check: func(cfg *loader.Config) bool {
			return strings.HasSuffix(cfg.Command[0], ".md")
		},
		Load: func(cfg *loader.Config) error {
			cfg.DefaultsPlayProps = ParseMarkdownMetas(cfg)
			return nil
		},
		PostLoad: func(cfg *loader.Config) error {
			return BuildScripts(cfg)
		},
	}
)
