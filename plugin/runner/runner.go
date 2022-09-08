package runner

import (
	"github.com/devthejo/snip/variable"
)

type Runner struct {
	Name   string
	Vars   map[string]*variable.Var
	Plugin *Plugin
}
