package play

import (
	"strings"

	cmap "github.com/orcaman/concurrent-map"
	"gitlab.com/ytopia/ops/snip/variable"
)

type RunVars struct {
	Parent   *RunVars
	Values   cmap.ConcurrentMap
	Defaults cmap.ConcurrentMap
}

func CreateRunVars() *RunVars {
	ctx := &RunVars{
		Values:   cmap.New(),
		Defaults: cmap.New(),
	}
	return ctx
}

func (ctx *RunVars) NewChild() *RunVars {
	r := CreateRunVars()
	r.Parent = ctx
	return r
}

func (ctx *RunVars) SetValueString(k string, v string) {
	r := variable.CreateRunVar()
	r.Param = v
	ctx.Values.Set(k, r)
}

func (ctx *RunVars) GetValues() cmap.ConcurrentMap {
	return ctx.Values
}
func (ctx *RunVars) GetDefaults() cmap.ConcurrentMap {
	return ctx.Defaults
}

func (ctx *RunVars) GetAll() map[string]string {
	vars := make(map[string]string)
	for k, v := range ctx.Defaults.Items() {
		runVar := v.(*variable.RunVar)
		vars[k] = runVar.GetValue(ctx, ctx.Parent)
	}
	for k, v := range ctx.Values.Items() {
		runVar := v.(*variable.RunVar)
		value := runVar.GetValue(ctx, ctx.Parent)
		if value != "" {
			vars[k] = value
		}
	}
	// for k, v := range vars {
	// 	vars[k], _ = goenv.Expand(v, vars)
	// }
	return vars
}

func (ctx *RunVars) GetPluginVars(pluginType string, pluginName string, useVars []string, mVar map[string]*variable.Var) map[string]string {
	pVars := make(map[string]string)
	for _, useV := range useVars {

		var val string

		key := strings.ToUpper(useV)

		v := mVar[key]
		if v != nil && v.GetDefault() != "" {
			val = v.GetDefault()
		}

		k1 := strings.ToUpper("@" + key)
		if cv := ctx.Get(k1); cv != "" {
			val = cv
		}
		k2 := strings.ToUpper("@" + pluginName + "_" + key)
		if cv := ctx.Get(k2); cv != "" {
			val = cv
		}
		k3 := strings.ToUpper("@" + pluginType + "_" + pluginName + "_" + key)
		if cv := ctx.Get(k3); cv != "" {
			val = cv
		}

		if v != nil && v.GetValue() != "" {
			val = v.GetValue()
		}

		pVars[strings.ToLower(key)] = val
	}
	return pVars
}

func (ctx *RunVars) Get(k string) string {
	var val string
	if v, ok := ctx.Values.Get(k); ok {
		r := v.(*variable.RunVar)
		val = r.GetValue(ctx, ctx.Parent)
	}
	if val == "" {
		if v, ok := ctx.Defaults.Get(k); ok {
			r := v.(*variable.RunVar)
			val = r.GetValue(ctx, ctx.Parent)
		}
	}
	// val, _ = goenv.Expand(val)
	return val
}
