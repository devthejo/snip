package registry

import (
	"strings"

	cmap "github.com/orcaman/concurrent-map"
	diskv "github.com/peterbourgon/diskv/v3"
)

type NsVars struct {
	m  cmap.ConcurrentMap
	pM *diskv.Diskv
}
type NsVarsOptions struct {
	BasePath string
}

func advancedTransform(key string) *diskv.PathKey {
	path := strings.Split(key, "/")
	last := len(path) - 1
	return &diskv.PathKey{
		Path:     path[:last],
		FileName: path[last],
	}
}
func inverseTransform(pathKey *diskv.PathKey) (key string) {
	return strings.Join(pathKey.Path, "/") + pathKey.FileName
}

func CreateNsVars(o *NsVarsOptions) *NsVars {
	v := &NsVars{
		m: cmap.New(),
	}
	d := diskv.New(diskv.Options{
		BasePath:          o.BasePath,
		CacheSizeMax:      1024 * 1024,
		AdvancedTransform: advancedTransform,
		InverseTransform:  inverseTransform,
	})
	v.pM = d

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
	return v.GetMap(serializeKey(s))
}
func (v *NsVars) HasVarBySlice(s []string, k string) bool {
	return v.HasVar(serializeKey(s), k)
}
func (v *NsVars) GetVarBySlice(s []string, k string) string {
	return v.GetVar(serializeKey(s), k)
}
func (v *NsVars) SetVarBySlice(s []string, k string, val string) {
	v.SetVar(serializeKey(s), k, val)
}

func (v *NsVars) PersistHasVar(ns string, k string) bool {
	return v.pM.Has(ns + "/" + k)
}
func (v *NsVars) PersistGetVar(ns string, k string) string {
	val, err := v.pM.Read(ns + "/" + k)
	if err != nil {
		return ""
	}
	return string(val)
}
func (v *NsVars) PersistSetVar(ns string, k string, val string) {
	v.pM.Write(ns+"/"+k, []byte(val))
}
func (v *NsVars) PersistHasVarBySlice(s []string, k string) bool {
	return v.PersistHasVar(serializeKey(s), k)
}
func (v *NsVars) PersistGetVarBySlice(s []string, k string) string {
	return v.PersistGetVar(serializeKey(s), k)
}
func (v *NsVars) PersistSetVarBySlice(s []string, k string, val string) {
	v.PersistSetVar(serializeKey(s), k, val)
}

func serializeKey(s []string) string {
	// b, _ := json.Marshal(s)
	// return string(b)
	return strings.Join(s, "/")
}
