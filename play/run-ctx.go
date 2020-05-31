package play

import (
	cmap "github.com/orcaman/concurrent-map"
)

type RunCtx struct {
	Vars        cmap.ConcurrentMap
	VarsDefault cmap.ConcurrentMap
}
