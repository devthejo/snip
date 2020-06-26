package registry

import "encoding/json"

type NsVars struct {
	m map[string]map[string]string
}

func CreateNsVars() *NsVars {
	v := &NsVars{
		m: make(map[string]map[string]string),
	}
	return v
}

func (v *NsVars) GetNsMap(ns string) map[string]map[string]string {
	return v.m
}
func (v *NsVars) GetMap(ns string) map[string]string {
	return v.m[ns]
}
func (v *NsVars) HasVar(ns string, k string) bool {
	if _, ok := v.m[ns]; !ok {
		return false
	}
	if _, ok := v.m[ns][k]; !ok {
		return false
	}
	return true
}
func (v *NsVars) GetVar(ns string, k string) string {
	if _, ok := v.m[ns]; !ok {
		return ""
	}
	return v.m[ns][k]
}
func (v *NsVars) SetVar(ns string, k string, val string) {
	if _, ok := v.m[ns]; !ok {
		v.m[ns] = make(map[string]string)
	}
	v.m[ns][k] = val
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
