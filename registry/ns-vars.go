package registry

import (
	"encoding/json"

	cmap "github.com/orcaman/concurrent-map"
)

type NsVars struct {
	m cmap.ConcurrentMap
}

func CreateNsVars() *NsVars {
	v := &NsVars{
		m: cmap.New(),
	}
	return v
}

func (v *NsVars) GetNsMap(ns string) cmap.ConcurrentMap {
	return v.m
}
func (v *NsVars) GetMap(ns string) map[string]string {
	i, ok := v.m.Get(ns)
	m := make(map[string]string)
	if ok {
		cm := i.(cmap.ConcurrentMap)
		for k, v := range cm.Items() {
			m[k] = v.(string)
		}
	}
	return m
}
func (v *NsVars) HasVar(ns string, k string) bool {
	i, ok := v.m.Get(ns)
	if !ok {
		return false
	}
	cm := i.(cmap.ConcurrentMap)
	if _, ok := cm.Get(k); !ok {
		return false
	}
	return true
}
func (v *NsVars) GetVar(ns string, k string) string {
	i, ok := v.m.Get(ns)
	if !ok {
		return ""
	}
	cm := i.(cmap.ConcurrentMap)
	if val, ok := cm.Get(k); ok {
		return val.(string)
	}
	return ""
}
func (v *NsVars) SetVar(ns string, k string, val string) {
	i, ok := v.m.Get(ns)
	var cm cmap.ConcurrentMap
	if !ok {
		cm = cmap.New()
		v.m.Set(ns, cm)
	} else {
		cm = i.(cmap.ConcurrentMap)
	}
	cm.Set(k, val)
}
func (v *NsVars) GetMapBySlice(s []string) map[string]string {
	b, _ := json.Marshal(s)
	return v.GetMap(string(b))
}
func (v *NsVars) HasVarBySlice(s []string, k string) bool {
	b, _ := json.Marshal(s)
	return v.HasVar(string(b), k)
}
func (v *NsVars) GetVarBySlice(s []string, k string) string {
	b, _ := json.Marshal(s)
	return v.GetVar(string(b), k)
}
func (v *NsVars) SetVarBySlice(s []string, k string, val string) {
	b, _ := json.Marshal(s)
	v.SetVar(string(b), k, val)
}
