package mainNative

import (
	cmap "github.com/orcaman/concurrent-map"
	loaderMardownMod "gitlab.com/youtopia.earth/ops/snip/plugins-native/loaders/markdown/mod"
)

func handleMod(mod string, args []string, codeBlock *CodeBlock, plugins cmap.ConcurrentMap) {
	// logrus.Debugf("mod: %v, args: %v", mod, args)
	modPluginI := getPlugin(plugins, mod)
	modCfg := &loaderMardownMod.Config{}
	modPlugin := modPluginI.(*loaderMardownMod.Plugin)
	modPlugin.Mod(modCfg)
}
