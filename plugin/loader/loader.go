package loader

import (
	"gitlab.com/youtopia.earth/ops/snip/variable"
)

type Loader struct {
	Name   string
	Vars   map[string]*variable.Var
	Plugin *Plugin
}
