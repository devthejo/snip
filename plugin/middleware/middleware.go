package middleware

import (
	"gitlab.com/youtopia.earth/ops/snip/variable"
)

type Middleware struct {
	Name   string
	Vars   map[string]*variable.Var
	Plugin *Plugin
}
