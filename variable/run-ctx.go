package variable

import (
	cmap "github.com/orcaman/concurrent-map"
)

type RunCtx interface {
	GetVars()        cmap.ConcurrentMap
	GetVarsDefault() cmap.ConcurrentMap
}