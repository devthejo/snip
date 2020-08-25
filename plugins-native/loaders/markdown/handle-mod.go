package mainNative

import (
	"gitlab.com/youtopia.earth/ops/snip/plugin/loader"
	"gitlab.com/youtopia.earth/ops/snip/plugins-native/loaders/markdown/blocks"
	loaderMardownMod "gitlab.com/youtopia.earth/ops/snip/plugins-native/loaders/markdown/mod"
)

func handleMod(mod string, args []string, codeBlock *blocks.Code, cfg *loader.Config) bool {
	// logrus.Debugf("mod: %v, args: %v", mod, args)
	modPluginI := getPlugin(mod)
	modCfg := &loaderMardownMod.Config{
		Args:         args,
		CodeBlock:    codeBlock,
		LoaderConfig: cfg,
	}
	modPlugin := modPluginI.(*loaderMardownMod.Plugin)
	modPlugin.Mod(modCfg)
	return modCfg.LoopContinue
}
