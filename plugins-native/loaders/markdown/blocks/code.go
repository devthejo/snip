package blocks

import (
	"gitlab.com/youtopia.earth/ops/snip/plugin/runner"
)

type Code struct {
	Index      int
	Lang       string
	Content    string
	Processors []func(*runner.Config, *string) (func(), error)
}
