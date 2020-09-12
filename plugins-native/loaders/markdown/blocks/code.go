package blocks

import (
	"gitlab.com/youtopia.earth/ops/snip/plugin/processor"
)

type Code struct {
	Index      int
	Lang       string
	Content    string
	Processors []func(*processor.Config, *string) error
}
