package play

import (
	cmap "github.com/orcaman/concurrent-map"
)

type RunCtx struct {
	Vars        cmap.ConcurrentMap
	VarsDefault cmap.ConcurrentMap
}

func CreateRunCtx() *RunCtx {
	ctx := &RunCtx{
		Vars:        cmap.New(),
		VarsDefault: cmap.New(),
	}
	return ctx
}

func (ctx *RunCtx) GetVars() cmap.ConcurrentMap {
	return ctx.Vars
}
func (ctx *RunCtx) GetVarsDefault() cmap.ConcurrentMap {
	return ctx.VarsDefault
}