package variable

import (
	cmap "github.com/orcaman/concurrent-map"
)

type RunVars interface {
	GetValues() cmap.ConcurrentMap
	GetDefaults() cmap.ConcurrentMap

	SetValueString(k string, v string)
	GetAll() map[string]string
	GetPluginVars(string, string, []string, map[string]*Var) map[string]string
	Get(k string) string
}
