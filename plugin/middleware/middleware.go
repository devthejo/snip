package middleware

import (
	"github.com/devthejo/snip/variable"
)

type Middleware struct {
	Name   string
	Vars   map[string]*variable.Var
	Plugin *Plugin
}
