package mainNative

import (
	"strings"

	cmap "github.com/orcaman/concurrent-map"

	"gitlab.com/youtopia.earth/ops/snip/plugin/loader"

	pluginSCP "gitlab.com/youtopia.earth/ops/snip/plugins-native/loaders/markdown/plugins/scp"
	pluginTMPL "gitlab.com/youtopia.earth/ops/snip/plugins-native/loaders/markdown/plugins/tmpl"
)

var (
	Plugins = cmap.New()
	Loader  = loader.Plugin{
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

func init() {
	loadNativePlugins()
}

func loadNativePlugins() {
	Plugins.Set("scp", &pluginSCP.Mod)
	Plugins.Set("tmpl", &pluginTMPL.Mod)
}
