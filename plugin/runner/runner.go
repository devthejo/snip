package runner

import (
	"gitlab.com/ytopia/ops/snip/variable"
)

type Runner struct {
	Name   string
	Vars   map[string]*variable.Var
	Plugin *Plugin
}
