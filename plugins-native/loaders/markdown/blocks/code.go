package blocks

import (
	"gitlab.com/ytopia/ops/snip/plugin/processor"
)

type Code struct {
	Index      int
	Lang       string
	Content    string
	Processors []func(*processor.Config, *string) error
}
