package mainNative

import (
	"strconv"
	"strings"

	cmap "github.com/orcaman/concurrent-map"

	"gitlab.com/youtopia.earth/ops/snip/plugin/loader"

	pluginSCP "gitlab.com/youtopia.earth/ops/snip/plugins-native/loaders/markdown/plugins/scp"
)

var (
	Plugins = cmap.New()
	Loader  = loader.Plugin{
		Check: func(cfg *loader.Config) bool {
			return strings.HasSuffix(cfg.Command[0], ".md")
		},
		Load: func(cfg *loader.Config) error {
			cfg.DefaultsPlayProps = ParseMarkdownMetas(cfg)

			checksums := ParseMarkdownBlocksChecksum(cfg)
			var varsMap map[interface{}]interface{}
			if v, ok := cfg.DefaultsPlayProps["vars"]; ok {
				varsMap = v.(map[interface{}]interface{})
			} else {
				varsMap = make(map[interface{}]interface{})
			}
			for i, sum := range checksums {
				key := "SNIP_SHA256_" + strconv.Itoa(i+1)
				varsMap[key] = sum
			}
			cfg.DefaultsPlayProps["vars"] = varsMap

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
}
