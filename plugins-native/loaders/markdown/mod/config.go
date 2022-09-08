package mod

import (
	"github.com/devthejo/snip/plugin/loader"
	"github.com/devthejo/snip/plugins-native/loaders/markdown/blocks"
)

type Config struct {
	Args         []string
	CodeBlock    *blocks.Code
	LoaderConfig *loader.Config
	LoopContinue bool
}
