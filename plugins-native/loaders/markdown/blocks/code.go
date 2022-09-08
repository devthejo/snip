package blocks

import (
	"github.com/devthejo/snip/plugin/processor"
)

type Code struct {
	Index      int
	Lang       string
	Content    string
	Processors []func(*processor.Config, *string) error
}
