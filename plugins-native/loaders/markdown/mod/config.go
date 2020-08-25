package mod

import (
	"gitlab.com/youtopia.earth/ops/snip/plugin/loader"
	"gitlab.com/youtopia.earth/ops/snip/plugins-native/loaders/markdown/blocks"
)

type Config struct {
	Args         []string
	CodeBlock    *blocks.Code
	LoaderConfig *loader.Config
	LoopContinue bool
}
