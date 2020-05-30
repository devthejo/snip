package play

import (
	cmap "github.com/orcaman/concurrent-map"
)

func CreateRunCtx() *RunCtx {
	ctx := &RunCtx{
		Vars:        cmap.New(),
		VarsDefault: cmap.New(),
	}
	return ctx
}

type RunCtx struct {
	Vars        cmap.ConcurrentMap
	VarsDefault cmap.ConcurrentMap

	ReadyToRun bool
}
