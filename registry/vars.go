package registry

import "encoding/json"

type Vars struct {
	m map[string]string
}

func CreateVars() *Vars {
	v := &Vars{
		m: make(map[string]string),
	}
	return v
}

func (v *Vars) GetMap() map[string]string {
	return v.m
}
func (v *Vars) HasVar(k string) bool {
	_, ok := v.m[k]
	return ok
}
func (v *Vars) GetVar(k string) string {
	return v.m[k]
}
func (v *Vars) SetVar(k string, val string) {
	v.m[k] = val
}
func (v *Vars) HasVarBySlice(s []string) bool {
	b, _ := json.Marshal(s)
	return v.HasVar(string(b))
}
func (v *Vars) GetVarBySlice(s []string) string {
	b, _ := json.Marshal(s)
	return v.GetVar(string(b))
}
func (v *Vars) SetVarBySlice(s []string, val string) {
	b, _ := json.Marshal(s)
	v.SetVar(string(b), val)
}
