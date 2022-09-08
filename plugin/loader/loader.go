package loader

import (
	"github.com/devthejo/snip/variable"
)

type Loader struct {
	Name   string
	Vars   map[string]*variable.Var
	Plugin *Plugin
}
