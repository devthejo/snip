package mod

import (
	"gitlab.com/ytopia/ops/snip/plugin/loader"
	"gitlab.com/ytopia/ops/snip/plugins-native/loaders/markdown/blocks"
)

type Config struct {
	Args         []string
	CodeBlock    *blocks.Code
	LoaderConfig *loader.Config
	LoopContinue bool
}
