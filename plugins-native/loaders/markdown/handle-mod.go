package mainNative

import (
	"github.com/devthejo/snip/plugin/loader"
	"github.com/devthejo/snip/plugins-native/loaders/markdown/blocks"
	loaderMarkdownMod "github.com/devthejo/snip/plugins-native/loaders/markdown/mod"
)

func handleMod(mod string, args []string, codeBlock *blocks.Code, cfg *loader.Config) bool {
	// logrus.Debugf("mod: %v, args: %v", mod, args)
	modPluginI := getPlugin(mod)
	modCfg := &loaderMarkdownMod.Config{
		Args:         args,
		CodeBlock:    codeBlock,
		LoaderConfig: cfg,
	}
	modPlugin := modPluginI.(*loaderMarkdownMod.Plugin)
	modPlugin.Mod(modCfg)
	return modCfg.LoopContinue
}
