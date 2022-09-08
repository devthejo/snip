package play

import (
	"fmt"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	shellquote "github.com/kballard/go-shellquote"
	"github.com/mgutz/ansi"
	"github.com/sirupsen/logrus"

	"gitlab.com/ytopia/ops/snip/decode"
	"gitlab.com/ytopia/ops/snip/errors"
	"gitlab.com/ytopia/ops/snip/plugin/loader"
	"gitlab.com/ytopia/ops/snip/plugin/middleware"
	"gitlab.com/ytopia/ops/snip/plugin/runner"
	"gitlab.com/ytopia/ops/snip/registry"
	"gitlab.com/ytopia/ops/snip/variable"
)

type CfgPlay struct {
	App App

	BuildCtx *BuildCtx

	ParentCfgPlay *CfgPlay

	Index int
	Key   string
	Title string

	CfgPlay interface{}

	VarsClean *bool
	Vars      map[string]*variable.Var
	LoopOn    []*CfgLoopRow

	VarsSets       map[string]map[string]*variable.Var
	LoopSets       map[string][]map[string]*variable.Var
	LoopSequential *bool

	RegisterVars map[string]*registry.VarDef

	Quiet          *bool
	CheckQuiet     *bool
	PreCheckQuiet  *bool
	PostCheckQuiet *bool

	CheckRetry        interface{}
	PreCheckRetry     interface{}
	PostCheckRetry    interface{}
	CheckInterval     time.Duration
	PreCheckInterval  time.Duration
	PostCheckInterval time.Duration
	CheckTimeout      time.Duration
	PreCheckTimeout   time.Duration
	PostCheckTimeout  time.Duration

	Check     []string
	PreCheck  []string
	PostCheck []string

	CfgPreChk  *CfgChk
	CfgPostChk *CfgChk

	Retry *int

	Dependencies []string
	PostInstall  []string

	Depth       int
	HasChildren bool

	Dir string

	ExecTimeout *time.Duration

	ForceLoader bool

	ParentBuildFile string
	BuildFile       string

	Loaders     *[]*loader.Loader
	Middlewares *[]*middleware.Middleware
	Runner      *runner.Runner

	GlobalRunCtx *GlobalRunCtx

	Scope string

	Tmpdir *bool

	Use     map[string]string
	Persist map[string]string
}

func CreateCfgPlay(app App, m map[string]interface{}, parentCfgPlay *CfgPlay, buildCtx *BuildCtx) *CfgPlay {

	cp := &CfgPlay{}

	cp.Vars = make(map[string]*variable.Var)
	cp.VarsSets = make(map[string]map[string]*variable.Var)
	cp.LoopSets = make(map[string][]map[string]*variable.Var)

	cp.App = app

	cp.BuildCtx = buildCtx

	if buildCtx.DefaultRunner != nil {
		cp.Runner = buildCtx.DefaultRunner
	}

	if buildCtx.DefaultVars != nil {
		for k, v := range buildCtx.DefaultVars {
			cp.Vars[k] = v
		}
	}

	if parentCfgPlay != nil {
		if parentCfgPlay.BuildFile != "" {
			cp.ParentBuildFile = parentCfgPlay.BuildFile
		} else {
			cp.ParentBuildFile = parentCfgPlay.ParentBuildFile
		}
	}

	if parentCfgPlay != nil {
		cp.GlobalRunCtx = parentCfgPlay.GlobalRunCtx
	} else {
		cp.GlobalRunCtx = CreateGlobalRunCtx()

		cfg := app.GetConfig()
		for _, pkey := range cfg.PlayKey {
			cp.GlobalRunCtx.NoSkipTreeKeys[pkey] = true
		}

	}

	cp.SetParentCfgPlay(parentCfgPlay)

	cp.ParseMap(m)

	return cp
}

func (cp *CfgPlay) SetParentCfgPlay(parentCfgPlay *CfgPlay) {
	cp.ParentCfgPlay = parentCfgPlay
	if parentCfgPlay == nil {
		return
	}
	cp.Depth = parentCfgPlay.Depth + 1
}

func (cp *CfgPlay) ParseMapAsDefault(m map[string]interface{}) {
	cp.ParseMapRun(m, false)
}
func (cp *CfgPlay) ParseMap(m map[string]interface{}) {
	cp.ParseMapRun(m, true)
}

func (cp *CfgPlay) ParseMapRun(m map[string]interface{}, override bool) {
	cp.ParseKey(m, override)
	cp.ParseTitle(m, override)
	cp.ParseDir(m, override)
	cp.ParseExecTimeout(m, override)
	cp.ParseVarsClean(m, override)
	cp.ParseVarsSets(m, override)
	cp.ParseLoopSets(m, override)
	cp.ParseLoopOn(m, override)
	cp.ParseLoopSequential(m, override)
	cp.ParseRetry(m, override)
	cp.ParseVars(m, override)
	cp.ParseRegisterVars(m, override)
	cp.ParseQuiet(m, override)
	cp.ParseCheckQuiet(m, override)
	cp.ParsePreCheckQuiet(m, override)
	cp.ParsePostCheckQuiet(m, override)
	cp.ParseDependencies(m, override)
	cp.ParsePostInstall(m, override)
	cp.ParseLoader(m, override)
	cp.ParseMiddlewares(m, override)
	cp.ParseRunner(m, override)
	cp.ParseScope(m, override)

	cp.ParseCheckRetry(m, override)
	cp.ParsePreCheckRetry(m, override)
	cp.ParsePostCheckRetry(m, override)
	cp.ParseCheckTimeout(m, override)
	cp.ParsePreCheckTimeout(m, override)
	cp.ParsePostCheckTimeout(m, override)
	cp.ParseCheckInterval(m, override)
	cp.ParsePreCheckInterval(m, override)
	cp.ParsePostCheckInterval(m, override)

	cp.ParseCheck(m, override)
	cp.ParsePreCheck(m, override)
	cp.ParsePostCheck(m, override)

	cp.ParseTmpdir(m, override)

	cp.ParseUse(m, override)
	cp.ParsePersist(m, override)
	cp.ParseVolumes(m, override)

	cp.ParsePlay(m, override)
}

func (cp *CfgPlay) ParsePlay(m map[string]interface{}, override bool) {
	if !override && cp.CfgPlay != nil {
		return
	}
	switch v := m["play"].(type) {
	case []interface{}:
		cp.HasChildren = true
		cp.CfgPlay = make([]*CfgPlay, 0)
		prevBuildCtx := cp.BuildCtx

		for i, mCfgPlay := range v {
			switch p2 := mCfgPlay.(type) {
			case map[interface{}]interface{}:
				m = make(map[string]interface{}, len(p2))
				for k, v := range p2 {
					m[k.(string)] = v
				}
			case map[string]interface{}:
				m = make(map[string]interface{}, len(p2))
				for k, v := range p2 {
					m[k] = v
				}
			case string:
				m = make(map[string]interface{})
				m["play"] = p2
			default:
				unexpectedTypeCfgPlay(m, "play")
			}

			buildCtx := CreateNextBuildCtx(prevBuildCtx)
			prevBuildCtx = buildCtx

			pI := CreateCfgPlay(cp.App, m, cp, buildCtx)
			pI.Index = i

			playSlice := cp.CfgPlay.([]*CfgPlay)
			playSlice = append(playSlice, pI)
			cp.CfgPlay = playSlice
		}
		for _, child := range cp.CfgPlay.([]*CfgPlay) {
			switch c := child.CfgPlay.(type) {
			case *CfgCmd:
				c.LoadCfgPlaySubstitution()
				c.RequirePostInstall()
				c.registerRequiredByPostInstall()
			}
		}

	case string:
		c, err := shellquote.Split(v)
		errors.Check(err)
		ccmd := CreateCfgCmd(cp, c)
		cp.CfgPlay = ccmd
		ccmd.RequireDependencies()
		ccmd.registerRequiredByDependencies()
		ccmd.RegisterInDependencies()
	case nil:
	default:
		unexpectedTypeCfgPlay(m, "play")
	}
}

func (cp *CfgPlay) ParseKey(m map[string]interface{}, override bool) {
	if !override && cp.Key != "" {
		return
	}
	switch v := m["key"].(type) {
	case string:
		cp.Key = v
	case nil:
	default:
		unexpectedTypeCmd(m, "key")
	}
}

func (cp *CfgPlay) ParseTitle(m map[string]interface{}, override bool) {
	if !override && cp.Title != "" {
		return
	}
	switch v := m["title"].(type) {
	case string:
		cp.Title = v
	case nil:
	default:
		unexpectedTypeCmd(m, "title")
	}
}

func (cp *CfgPlay) ParseDir(m map[string]interface{}, override bool) {
	if !override && cp.Dir != "" {
		return
	}
	switch v := m["dir"].(type) {
	case string:
		if v == "." {
			cfg := cp.App.GetConfig()
			cp.Dir = path.Join(cfg.SnippetsDir, m["snippetDir"].(string))
		} else {
			cp.Dir = v
		}
		fmt.Printf("cp.Dir %v \n", cp.Dir)
	case nil:
	default:
		unexpectedTypeCmd(m, "dir")
	}
}

func (cp *CfgPlay) ParseExecTimeout(m map[string]interface{}, override bool) {
	if !override && cp.ExecTimeout != nil {
		return
	}
	t, ok := m["timeout"]
	if !ok {
		if cp.ParentCfgPlay != nil {
			cp.ExecTimeout = cp.ParentCfgPlay.ExecTimeout
		}
		return
	}
	timeout, err := decode.Duration(t)
	errors.Check(err)
	if timeout != 0 {
		cp.ExecTimeout = &timeout
	}
}

func (cp *CfgPlay) ParseVarsClean(m map[string]interface{}, override bool) {
	if !override && cp.VarsClean != nil {
		return
	}
	switch v := m["vars_clean"].(type) {
	case bool:
		cp.VarsClean = &v
	case nil:
	default:
		unexpectedTypeCfgPlay(m, "vars_clean")
	}
}

func (cp *CfgPlay) ParseVarsSets(m map[string]interface{}, override bool) {
	switch v := m["vars_sets"].(type) {
	case map[string]interface{}:
		vsets, err := decode.ToMap(v)
		errors.Check(err)
		for vsetKey, vsetVal := range vsets {
			switch vsetV := vsetVal.(type) {
			case map[string]interface{}:
				_, hk := cp.VarsSets[vsetKey]
				if !hk || override {
					cp.VarsSets[vsetKey] = variable.ParseVarsMap(vsetV, cp.Depth)
				}
			default:
				variable.UnexpectedTypeVarValue(vsetKey, vsetVal)
			}
		}
	case nil:
	default:
		unexpectedTypeCfgPlay(m, "vars_sets")
	}
	if cp.ParentCfgPlay != nil {
		for k, v := range cp.ParentCfgPlay.VarsSets {
			if _, hk := cp.VarsSets[k]; !hk {
				cp.VarsSets[k] = v
			}
		}
	}
}
func (cp *CfgPlay) ParseLoopSets(m map[string]interface{}, override bool) {
	switch v := m["loop_sets"].(type) {
	case map[string]interface{}:
		loops, err := decode.ToMap(v)
		errors.Check(err)
		for loopKey, loopVal := range loops {
			_, hk := cp.LoopSets[loopKey]
			if hk && !override {
				continue
			}
			switch loopV := loopVal.(type) {
			case []interface{}:
				cp.LoopSets[loopKey] = make([]map[string]*variable.Var, len(loopV))
				for i, rowVal := range loopV {
					switch rowV := rowVal.(type) {
					case map[string]interface{}:
						cp.LoopSets[loopKey][i] = variable.ParseVarsMap(rowV, cp.Depth)
					case string:
						if cp.VarsSets[rowV] == nil {
							logrus.Fatalf("undefined VarsSet %v called by loop_sets %v", rowV, loopKey)
						}
						cp.LoopSets[loopKey][i] = cp.VarsSets[rowV]
					default:
						variable.UnexpectedTypeVarValue(strconv.Itoa(i), rowVal)
					}
				}
			default:
				variable.UnexpectedTypeVarValue(loopKey, loopVal)
			}
		}
	case nil:
	default:
		unexpectedTypeCfgPlay(m, "loop_sets")
	}
	if cp.ParentCfgPlay != nil {
		for k, v := range cp.ParentCfgPlay.LoopSets {
			if _, hk := cp.LoopSets[k]; !hk {
				cp.LoopSets[k] = v
			}
		}
	}
}
func (cp *CfgPlay) ParseLoopOn(m map[string]interface{}, override bool) {
	if cp.LoopOn != nil && !override {
		return
	}
	switch v := m["loop_on"].(type) {
	case string:
		if cp.LoopSets[v] == nil {
			logrus.Fatalf("undefined LoopSet %v called by loop_on", v)
		}
		cp.LoopOn = make([]*CfgLoopRow, len(cp.LoopSets[v]))
		for loopI, loopV := range cp.LoopSets[v] {
			cfgLoopRow := CreateCfgLoopRow(loopI, v, loopV)
			cp.LoopOn[loopI] = cfgLoopRow
		}
	case []interface{}:
		cp.LoopOn = make([]*CfgLoopRow, len(v))
		for loopI, loopV := range v {
			var cfgLoopRow *CfgLoopRow
			switch loop := loopV.(type) {
			case string:
				loop = strings.ToLower(loop)
				if cp.VarsSets[loop] == nil {
					logrus.Fatalf("undefined LoopSet %v", loop)
				}
				cfgLoopRow = CreateCfgLoopRow(loopI, loop, cp.VarsSets[loop])
			case map[interface{}]interface{}, map[string]interface{}:
				l, err := decode.ToMap(loop)
				errors.Check(err)
				cfgLoopRow = CreateCfgLoopRow(loopI, "", variable.ParseVarsMap(l, cp.Depth))
			default:
				variable.UnexpectedTypeVarValue(strconv.Itoa(loopI), loopV)
			}
			cp.LoopOn[loopI] = cfgLoopRow
		}
	case nil:
	default:
		unexpectedTypeCfgPlay(m, "loop_on")
	}
}
func (cp *CfgPlay) ParseLoopSequential(m map[string]interface{}, override bool) {
	switch v := m["loop_sequential"].(type) {
	case bool:
		if cp.LoopSequential == nil || override {
			cp.LoopSequential = &v
		}
	case nil:
	default:
		unexpectedTypeCfgPlay(m, "loop_sequential")
	}
}

func (cp *CfgPlay) ParseVars(m map[string]interface{}, override bool) {
	switch v := m["vars"].(type) {
	case map[interface{}]interface{}, map[string]interface{}:
		m, err := decode.ToMap(v)
		errors.Check(err)
		for key, val := range variable.ParseVarsMap(m, cp.Depth) {
			_, hk := cp.Vars[key]
			if override || !hk {
				cp.Vars[key] = val
			}
		}
	case nil:
	default:
		unexpectedTypeCmd(m, "vars")
	}
}

func parseRegisterVarsItemMap(mI map[string]interface{}, defaultKey string) *registry.VarDef {
	m, err := decode.ToMap(mI)
	if err != nil {
		logrus.Fatalf("unexpected register_vars type %T value %v, %v", mI, mI, err)
	}

	switch val := m["key"].(type) {
	case string:
		defaultKey = val
	case nil:
	default:
		logrus.Fatalf("unexpected register_vars key type %T value %v, %v", val, val, err)
	}

	var enable bool
	switch val := m["enable"].(type) {
	case bool:
		enable = val
	case nil:
		enable = true
	default:
		logrus.Fatalf("unexpected register_vars enable type %T value %v, %v", val, val, err)
	}

	var persist bool
	switch val := m["persist"].(type) {
	case bool:
		persist = val
	case nil:
	default:
		logrus.Fatalf("unexpected register_vars persist type %T value %v, %v", val, val, err)
	}

	var sourceStdout bool
	switch val := m["stdout"].(type) {
	case bool:
		sourceStdout = val
	case nil:
	default:
		logrus.Fatalf("unexpected register_vars stdout type %T value %v, %v", val, val, err)
	}

	var to string
	switch val := m["to"].(type) {
	case string:
		to = val
	case nil:
		if defaultKey == "" {
			logrus.Fatalf("missing register_vars to %v", m)
		}
		to = defaultKey
	default:
		logrus.Fatalf("unexpected register_vars to type %T value %v, %v", val, val, err)
	}

	var from string
	switch val := m["from"].(type) {
	case string:
		from = val
	case nil:
		if defaultKey != "" {
			from = defaultKey
		} else {
			from = to
		}
	default:
		logrus.Fatalf("unexpected register_vars from type %T value %v, %v", val, val, err)
	}

	var source string
	switch val := m["source"].(type) {
	case string:
		source = val
	case nil:
	default:
		logrus.Fatalf("unexpected register_vars source type %T value %v, %v", val, val, err)
	}

	to = strings.ToUpper(to)
	from = strings.ToUpper(from)
	source = strings.ToUpper(source)

	if sourceStdout && source != "" {
		logrus.Fatalf("unexpected, register_vars source and source_stdout are mutually exclusive %v", m)
	}

	v := &registry.VarDef{
		To:           to,
		From:         from,
		Source:       source,
		SourceStdout: sourceStdout,
		Enable:       enable,
		Persist:      persist,
	}
	return v
}

func (cp *CfgPlay) ParseRegisterVars(m map[string]interface{}, override bool) {

	tmpV := make(map[string]*registry.VarDef)

	if cp.RegisterVars == nil {
		cp.RegisterVars = make(map[string]*registry.VarDef)
		if cp.ParentCfgPlay != nil {
			for _, v := range cp.ParentCfgPlay.RegisterVars {
				key := v.GetFrom()
				var source string
				if !v.SourceStdout {
					source = v.GetSource()
				}
				tmpV[key] = &registry.VarDef{
					To:           key,
					From:         key,
					Source:       source,
					SourceStdout: v.SourceStdout,
					Enable:       v.Enable,
					Persist:      v.Persist,
				}
			}
		}
	}

	switch rVars := m["register_vars"].(type) {
	case []interface{}:
		for _, varsItemI := range rVars {
			switch item := varsItemI.(type) {
			case string:
				item = strings.ToUpper(item)
				tmpV[item] = &registry.VarDef{
					To:     item,
					From:   item,
					Enable: true,
				}
			case map[interface{}]interface{}:
				m, err := decode.ToMap(item)
				if err != nil {
					logrus.Fatalf("unexpected register_vars item type %T value %v, %v", rVars, rVars, err)
				}
				v := parseRegisterVarsItemMap(m, "")
				tmpV[v.To] = v
			case map[string]interface{}:
				v := parseRegisterVarsItemMap(item, "")
				tmpV[v.To] = v
			default:
				logrus.Fatalf("unexpected register_vars type %T value %v", item, item)
			}
		}
	case map[interface{}]interface{}, map[string]interface{}:
		rvm, err := decode.ToMap(rVars)
		if err != nil {
			logrus.Fatalf("unexpected register_vars type %T value %v, %v", rVars, rVars, err)
		}
		for k, rVarI := range rvm {
			switch rVar := rVarI.(type) {
			case map[interface{}]interface{}:
				m, err := decode.ToMap(rVar)
				if err != nil {
					logrus.Fatalf("unexpected register_vars item type %T value %v, %v", rVars, rVars, err)
				}
				v := parseRegisterVarsItemMap(m, k)
				tmpV[v.To] = v
			case map[string]interface{}:
				v := parseRegisterVarsItemMap(rVar, k)
				tmpV[v.To] = v
			case bool:
				k := strings.ToUpper(k)
				tmpV[k] = &registry.VarDef{
					To:     k,
					From:   k,
					Enable: rVar,
				}
			case nil:
				k := strings.ToUpper(k)
				tmpV[k] = &registry.VarDef{
					To:     k,
					From:   k,
					Enable: true,
				}
			default:
				logrus.Fatalf("unexpected register_vars value type %T value %v, %v", rVar, rVar, err)
			}
		}
	case nil:
	default:
		unexpectedTypeCmd(m, "register_vars")
	}

	for _, v := range tmpV {
		if !v.Enable && !override {
			continue
		}
		k := v.To
		cp.RegisterVars[k] = v
		if v.Enable && cp.Vars[k] == nil {
			cp.Vars[k] = &variable.Var{
				Name:  k,
				Depth: cp.Depth,
			}
		}
	}
}

func (cp *CfgPlay) ParseQuiet(m map[string]interface{}, override bool) {
	switch v := m["quiet"].(type) {
	case bool:
		if cp.Quiet == nil || override {
			cp.Quiet = &v
		}
	case nil:
		if cp.ParentCfgPlay != nil && cp.ParentCfgPlay.Quiet != nil {
			cp.Quiet = cp.ParentCfgPlay.Quiet
		}
	default:
		unexpectedTypeCfgPlay(m, "quiet")
	}
}

func (cp *CfgPlay) ParseTmpdir(m map[string]interface{}, override bool) {
	switch v := m["tmpdir"].(type) {
	case bool:
		if cp.Tmpdir == nil || override {
			cp.Tmpdir = &v
		}
	case nil:
		if cp.ParentCfgPlay != nil && cp.ParentCfgPlay.Tmpdir != nil {
			cp.Tmpdir = cp.ParentCfgPlay.Tmpdir
		}
	default:
		unexpectedTypeCfgPlay(m, "tmpdir")
	}
}

func (cp *CfgPlay) ParseVolumes(m map[string]interface{}, override bool) {
	if cp.Persist == nil {
		cp.Persist = make(map[string]string, 0)
	}
	if cp.Use == nil {
		cp.Use = make(map[string]string, 0)
	}
	switch v := m["volumes"].(type) {
	case nil:
	case map[string]interface{}, map[interface{}]interface{}:
		itemMap, err := decode.ToMap(v)
		if err != nil {
			logrus.Fatalf("unexpected volumes type %T value %v, %v", v, v, err)
		}
		for target, sourceI := range itemMap {
			if _, ok := cp.Use[target]; !ok || override {
				cp.Use[target] = sourceI.(string)
				cp.Use[sourceI.(string)] = target
			}
		}
	case []interface{}, []string:
		var source string
		var target string
		for _, itemI := range v.([]interface{}) {
			switch item := itemI.(type) {
			case string:
				if strings.Contains(item, ":") {
					parts := strings.Split(item, ":")
					target = parts[0]
					source = parts[1]
				} else {
					target = item
					source = item
				}
			case map[string]interface{}, map[interface{}]interface{}:
				itemMap, err := decode.ToMap(item)
				if err != nil {
					logrus.Fatalf("unexpected use item type %T value %v, %v", item, item, err)
				}
				source = itemMap["source"].(string)
				target = itemMap["target"].(string)
			}
			if _, ok := cp.Use[target]; !ok || override {
				cp.Use[target] = source
				cp.Persist[source] = target
			}
		}
	default:
		unexpectedTypeCmd(m, "volumes")
	}
}

func (cp *CfgPlay) ParseUse(m map[string]interface{}, override bool) {
	if cp.Use == nil {
		cp.Use = make(map[string]string, 0)
	}
	switch v := m["use"].(type) {
	case nil:
	case map[string]interface{}, map[interface{}]interface{}:
		itemMap, err := decode.ToMap(v)
		if err != nil {
			logrus.Fatalf("unexpected use type %T value %v, %v", v, v, err)
		}
		for target, sourceI := range itemMap {
			if _, ok := cp.Use[target]; !ok || override {
				cp.Use[target] = sourceI.(string)
			}
		}
	case []interface{}, []string:
		var source string
		var target string
		for _, itemI := range v.([]interface{}) {
			switch item := itemI.(type) {
			case string:
				if strings.Contains(item, ":") {
					parts := strings.Split(item, ":")
					target = parts[0]
					source = parts[1]
				} else {
					target = item
					source = item
				}
			case map[string]interface{}, map[interface{}]interface{}:
				itemMap, err := decode.ToMap(item)
				if err != nil {
					logrus.Fatalf("unexpected use item type %T value %v, %v", item, item, err)
				}
				source = itemMap["source"].(string)
				target = itemMap["target"].(string)
			}
			if _, ok := cp.Use[target]; !ok || override {
				cp.Use[target] = source
			}
		}
	default:
		unexpectedTypeCmd(m, "use")
	}
}

func (cp *CfgPlay) ParsePersist(m map[string]interface{}, override bool) {
	if cp.Persist == nil {
		cp.Persist = make(map[string]string, 0)
	}
	switch v := m["persist"].(type) {
	case nil:
	case map[string]interface{}, map[interface{}]interface{}:
		itemMap, err := decode.ToMap(v)
		if err != nil {
			logrus.Fatalf("unexpected persist type %T value %v, %v", v, v, err)
		}
		for target, sourceI := range itemMap {
			if _, ok := cp.Persist[target]; !ok || override {
				cp.Persist[target] = sourceI.(string)
			}
		}
	case []interface{}, []string:
		var source string
		var target string
		for _, itemI := range v.([]interface{}) {
			switch item := itemI.(type) {
			case string:
				if strings.Contains(item, ":") {
					parts := strings.Split(item, ":")
					target = parts[0]
					source = parts[1]
				} else {
					target = item
					source = item
				}
			case map[string]interface{}, map[interface{}]interface{}:
				itemMap, err := decode.ToMap(item)
				if err != nil {
					logrus.Fatalf("unexpected persist item type %T value %v, %v", item, item, err)
				}
				source = itemMap["source"].(string)
				target = itemMap["target"].(string)
			}
			if _, ok := cp.Persist[target]; !ok || override {
				cp.Persist[target] = source
			}
		}
	default:
		unexpectedTypeCmd(m, "persist")
	}
}

func (cp *CfgPlay) ParseCheck(m map[string]interface{}, override bool) {
	if !override && cp.Check != nil {
		return
	}
	switch v := m["check"].(type) {
	case string:
		if strings.Contains(v, "\n") {
			cp.Check = []string{v}
		} else {
			s, err := shellquote.Split(v)
			errors.Check(err)
			cp.Check = s
		}
	case []interface{}:
		s, err := decode.ToStrings(v)
		errors.Check(err)
		cp.Check = s
	case nil:
		if _, ok := m["check"]; ok {
			cp.Check = make([]string, 0)
		}
	default:
		unexpectedTypeCmd(m, "check")
	}
}
func (cp *CfgPlay) ParsePreCheck(m map[string]interface{}, override bool) {
	if !override && cp.CfgPreChk != nil {
		return
	}
	switch v := m["pre_check"].(type) {
	case string:
		if strings.Contains(v, "\n") {
			cp.PreCheck = []string{v}
		} else {
			s, err := shellquote.Split(v)
			errors.Check(err)
			cp.PreCheck = s
		}
	case []interface{}:
		s, err := decode.ToStrings(v)
		errors.Check(err)
		cp.PreCheck = s
	case nil:
		if _, ok := m["pre_check"]; ok {
			cp.PreCheck = make([]string, 0)
		}
	default:
		unexpectedTypeCmd(m, "pre_check")
	}
	if len(cp.PreCheck) > 0 {
		cp.CfgPreChk = CreateCfgChk(cp, cp.PreCheck)
	} else if len(cp.Check) > 0 {
		cp.CfgPreChk = CreateCfgChk(cp, cp.Check)
	}
}
func (cp *CfgPlay) ParsePostCheck(m map[string]interface{}, override bool) {
	if !override && cp.CfgPostChk != nil {
		return
	}
	switch v := m["post_check"].(type) {
	case string:
		if strings.Contains(v, "\n") {
			cp.PostCheck = []string{v}
		} else {
			s, err := shellquote.Split(v)
			errors.Check(err)
			cp.PostCheck = s
		}
	case []interface{}:
		s, err := decode.ToStrings(v)
		errors.Check(err)
		cp.PostCheck = s
	case nil:
		if _, ok := m["post_check"]; ok {
			cp.PostCheck = make([]string, 0)
		}
	default:
		unexpectedTypeCmd(m, "post_check")
	}
	if len(cp.PostCheck) > 0 {
		cp.CfgPostChk = CreateCfgChk(cp, cp.PostCheck)
	} else if len(cp.Check) > 0 {
		cp.CfgPostChk = CreateCfgChk(cp, cp.Check)
	}
}

func (cp *CfgPlay) ParseCheckQuiet(m map[string]interface{}, override bool) {
	switch v := m["check_quiet"].(type) {
	case bool:
		if cp.CheckQuiet == nil || override {
			cp.CheckQuiet = &v
		}
	case nil:
		if cp.ParentCfgPlay != nil && cp.ParentCfgPlay.CheckQuiet != nil {
			cp.CheckQuiet = cp.ParentCfgPlay.CheckQuiet
		} else if cp.Quiet != nil {
			cp.CheckQuiet = cp.Quiet
		}
	default:
		unexpectedTypeCfgPlay(m, "check_quiet")
	}
}
func (cp *CfgPlay) ParsePreCheckQuiet(m map[string]interface{}, override bool) {
	if !override && cp.PreCheckQuiet != nil {
		return
	}
	switch v := m["pre_check_quiet"].(type) {
	case bool:
		cp.PreCheckQuiet = &v
	case nil:
		if cp.PreCheckQuiet == nil && cp.CheckQuiet != nil {
			v := *cp.CheckQuiet
			cp.PreCheckQuiet = &v
		}
	default:
		unexpectedTypeCmd(m, "pre_check_quiet")
	}
}
func (cp *CfgPlay) ParsePostCheckQuiet(m map[string]interface{}, override bool) {
	if !override && cp.PostCheckQuiet != nil {
		return
	}
	switch v := m["post_check_quiet"].(type) {
	case bool:
		cp.PostCheckQuiet = &v
	case nil:
		if cp.PostCheckQuiet == nil && cp.CheckQuiet != nil {
			v := *cp.CheckQuiet
			cp.PostCheckQuiet = &v
		}
	default:
		unexpectedTypeCmd(m, "post_check_quiet")
	}
}

func (cp *CfgPlay) ParseRetry(m map[string]interface{}, override bool) {
	if !override && cp.Retry != nil {
		return
	}
	switch v := m["retry"].(type) {
	case int64:
		r := int(v)
		cp.Retry = &r
	case int:
		cp.Retry = &v
	case string:
		r, err := strconv.Atoi(v)
		if err != nil {
			unexpectedTypeCfgPlay(m, "retry")
		}
		cp.Retry = &r
	case nil:
		if cp.ParentCfgPlay != nil {
			cp.Retry = cp.ParentCfgPlay.Retry
		}
	default:
		unexpectedTypeCfgPlay(m, "retry")
	}
}

func (cp *CfgPlay) ParseCheckRetry(m map[string]interface{}, override bool) {
	if !override && cp.CheckRetry != nil {
		return
	}
	switch v := m["check_retry"].(type) {
	case int64:
		r := int(v)
		cp.CheckRetry = &r
	case int:
		cp.CheckRetry = &v
	case string:
		if v == "false" {
			b := false
			cp.CheckRetry = &b
		} else if v == "true" {
			b := true
			cp.CheckRetry = &b
		}
		r, err := strconv.Atoi(v)
		if err != nil {
			unexpectedTypeCfgPlay(m, "check_retry")
		}
		cp.CheckRetry = &r
	case bool:
		cp.CheckRetry = &v
	case nil:
		if cp.ParentCfgPlay != nil {
			cp.CheckRetry = cp.ParentCfgPlay.CheckRetry
		}
	default:
		unexpectedTypeCfgPlay(m, "check_retry")
	}
}

func (cp *CfgPlay) ParsePreCheckRetry(m map[string]interface{}, override bool) {
	if !override && cp.PreCheckRetry != nil {
		return
	}
	switch v := m["pre_check_retry"].(type) {
	case int64:
		r := int(v)
		cp.PreCheckRetry = &r
	case int:
		cp.PreCheckRetry = &v
	case string:
		if v == "false" {
			b := false
			cp.PreCheckRetry = &b
		} else if v == "true" {
			b := true
			cp.PreCheckRetry = &b
		}
		r, err := strconv.Atoi(v)
		if err != nil {
			unexpectedTypeCfgPlay(m, "pre_check_retry")
		}
		cp.PreCheckRetry = &r
	case bool:
		cp.PreCheckRetry = &v
	case nil:
		cp.PreCheckRetry = cp.CheckRetry
	default:
		unexpectedTypeCfgPlay(m, "pre_check_retry")
	}
}
func (cp *CfgPlay) ParsePostCheckRetry(m map[string]interface{}, override bool) {
	if !override && cp.PostCheckRetry != nil {
		return
	}
	switch v := m["post_check_retry"].(type) {
	case int64:
		r := int(v)
		cp.PostCheckRetry = &r
	case int:
		cp.PostCheckRetry = &v
	case string:
		if v == "false" {
			b := false
			cp.PostCheckRetry = &b
		} else if v == "true" {
			b := true
			cp.PostCheckRetry = &b
		}
		r, err := strconv.Atoi(v)
		if err != nil {
			unexpectedTypeCfgPlay(m, "post_check_retry")
		}
		cp.PostCheckRetry = &r
	case bool:
		cp.PostCheckRetry = &v
	case nil:
		cp.PostCheckRetry = cp.CheckRetry
	default:
		unexpectedTypeCfgPlay(m, "post_check_retry")
	}
}
func (cp *CfgPlay) ParseCheckTimeout(m map[string]interface{}, override bool) {
	if !override && cp.CheckTimeout != 0 {
		return
	}
	timeoutDuration, err := decode.Duration(m["check_timeout"])
	if err != nil {
		logrus.Error(err)
		logrus.Fatalf(`Unexpected check_timeout format "%v", expected seconds (e.g: 60) or duration (e.g 1h15m30s)`, timeoutDuration)
	}
	cp.CheckTimeout = timeoutDuration
}
func (cp *CfgPlay) ParsePreCheckTimeout(m map[string]interface{}, override bool) {
	if !override && cp.PreCheckTimeout != 0 {
		return
	}
	timeoutDuration, err := decode.Duration(m["pre_check_timeout"])
	if err != nil {
		logrus.Error(err)
		logrus.Fatalf(`Unexpected pre_check_timeout format "%v", expected seconds (e.g: 60) or duration (e.g 1h15m30s)`, timeoutDuration)
	}
	if timeoutDuration == 0 {
		timeoutDuration = cp.CheckTimeout
	}
	cp.PreCheckTimeout = timeoutDuration
}
func (cp *CfgPlay) ParsePostCheckTimeout(m map[string]interface{}, override bool) {
	if !override && cp.PostCheckTimeout != 0 {
		return
	}
	timeoutDuration, err := decode.Duration(m["post_check_timeout"])
	if err != nil {
		logrus.Error(err)
		logrus.Fatalf(`Unexpected post_check_timeout format "%v", expected seconds (e.g: 60) or duration (e.g 1h15m30s)`, timeoutDuration)
	}
	if timeoutDuration == 0 {
		timeoutDuration = cp.CheckTimeout
	}
	cp.PostCheckTimeout = timeoutDuration
}

func (cp *CfgPlay) ParseCheckInterval(m map[string]interface{}, override bool) {
	if !override && cp.CheckInterval != 0 {
		return
	}
	if checkInterval, ok := m["check_interval"]; ok {
		intervalDuration, err := decode.Duration(checkInterval)
		if err != nil {
			logrus.Error(err)
			logrus.Fatalf(`Unexpected check_interval format "%v", expected seconds (e.g: 60) or duration (e.g 1h15m30s)`, intervalDuration)
		}
		cp.CheckInterval = intervalDuration
	} else {
		cp.CheckInterval = time.Second * 2
	}
}
func (cp *CfgPlay) ParsePreCheckInterval(m map[string]interface{}, override bool) {
	if !override && cp.PreCheckInterval != 0 {
		return
	}
	intervalDuration, err := decode.Duration(m["pre_check_interval"])
	if err != nil {
		logrus.Error(err)
		logrus.Fatalf(`Unexpected pre_check_interval format "%v", expected seconds (e.g: 60) or duration (e.g 1h15m30s)`, intervalDuration)
	}
	if intervalDuration == 0 {
		intervalDuration = cp.CheckInterval
	}
	cp.PreCheckInterval = intervalDuration
}
func (cp *CfgPlay) ParsePostCheckInterval(m map[string]interface{}, override bool) {
	if !override && cp.PostCheckInterval != 0 {
		return
	}
	intervalDuration, err := decode.Duration(m["post_check_timeout"])
	if err != nil {
		logrus.Error(err)
		logrus.Fatalf(`Unexpected post_check_timeout format "%v", expected seconds (e.g: 60) or duration (e.g 1h15m30s)`, intervalDuration)
	}
	if intervalDuration == 0 {
		intervalDuration = cp.CheckInterval
	}
	cp.PostCheckInterval = intervalDuration
}

func (cp *CfgPlay) ParseDependencies(m map[string]interface{}, override bool) {
	if !override && cp.Dependencies != nil {
		return
	}
	switch m["dependencies"].(type) {
	case []interface{}:
		dependencies, err := decode.ToStrings(m["dependencies"])
		errors.Check(err)
		cp.Dependencies = dependencies
	case nil:
		if _, ok := m["dependencies"]; ok {
			cp.Dependencies = make([]string, 0)
		}
	default:
		unexpectedTypeCmd(m, "dependencies")
	}
}

func (cp *CfgPlay) ParsePostInstall(m map[string]interface{}, override bool) {
	if !override && cp.PostInstall != nil {
		return
	}
	switch m["post_install"].(type) {
	case []interface{}:
		post_install, err := decode.ToStrings(m["post_install"])
		errors.Check(err)
		cp.PostInstall = post_install
	case nil:
		if _, ok := m["post_install"]; ok {
			cp.PostInstall = make([]string, 0)
		}
	default:
		unexpectedTypeCmd(m, "post_install")
	}
}

func (cp *CfgPlay) ParseLoader(m map[string]interface{}, override bool) {
	if !override && cp.Loaders != nil {
		return
	}

	app := cp.App

	var loadersI []interface{}

	switch v := m["loader"].(type) {
	case []interface{}:
		loadersI = v
	case string, map[string]interface{}, map[interface{}]interface{}:
		cp.ForceLoader = true
		loadersI = append(loadersI, v)
	case nil:
		if cp.ParentCfgPlay != nil {
			cp.Loaders = cp.ParentCfgPlay.Loaders
		}
		return
	default:
		unexpectedTypeCmd(m, "loader")
		return
	}

	loaders := make([]*loader.Loader, len(loadersI))
	for i, loaderI := range loadersI {
		switch loaderV := loaderI.(type) {
		case string:
			loaders[i] = &loader.Loader{
				Name:   loaderV,
				Plugin: app.GetLoader(loaderV),
			}
		case map[interface{}]interface{}, map[string]interface{}:
			loaderMap, err := decode.ToMap(loaderV)
			if err != nil {
				logrus.Fatalf("unexpected loader type %T value %v, %v", loaderV, loaderV, err)
			}
			name := loaderMap["name"].(string)
			mr := &loader.Loader{
				Name:   name,
				Plugin: app.GetLoader(name),
			}
			if loaderMap["vars"] != nil {
				varsI, err := decode.ToMap(loaderMap["vars"])
				if err != nil {
					logrus.Fatalf("unexpected loader vars type %T value %v, %v", loaderMap["vars"], loaderMap["vars"], err)
				}
				mr.Vars = variable.ParseVarsMap(varsI, cp.Depth+1)
			}
			loaders[i] = mr
		default:
			logrus.Fatalf("unexpected loader type %T value %v", loaderI, loaderI)
		}
	}
	cp.Loaders = &loaders

}
func (cp *CfgPlay) ParseRunner(m map[string]interface{}, override bool) {
	if !override && cp.Runner != nil {
		return
	}

	switch runnerV := m["runner"].(type) {
	case string:
		cp.Runner = &runner.Runner{
			Name: runnerV,
		}
	case map[interface{}]interface{}, map[string]interface{}:
		runnerMap, err := decode.ToMap(runnerV)
		if err != nil {
			logrus.Fatalf("unexpected runner type %T value %v, %v", runnerV, runnerV, err)
		}
		name := runnerMap["name"].(string)
		rr := &runner.Runner{
			Name: name,
		}
		if runnerMap["vars"] != nil {
			varsI, err := decode.ToMap(runnerMap["vars"])
			if err != nil {
				logrus.Fatalf("unexpected runner vars type %T value %v, %v", runnerMap["vars"], runnerMap["vars"], err)
			}
			rr.Vars = variable.ParseVarsMap(varsI, cp.Depth+1)
		}
		cp.Runner = rr
	case nil:
		if cp.Runner == nil && cp.ParentCfgPlay != nil {
			cp.Runner = cp.ParentCfgPlay.Runner
		}
	default:
		logrus.Fatalf("unexpected runner type %T value %v", runnerV, runnerV)
	}
}

func (cp *CfgPlay) ParseScope(m map[string]interface{}, override bool) {
	if !override && cp.Scope != "" {
		return
	}

	var scopeSlice []string

	switch scopeNs := m["scope_namespace"].(type) {
	case string:
		scopeSlice = append(scopeSlice, scopeNs)
	case nil:
	default:
		logrus.Fatalf("unexpected scope_namespace type %T value %v", scopeNs, scopeNs)
	}

	switch scope := m["scope"].(type) {
	case string:
		scopeSlice = append(scopeSlice, scope)
	case nil:
		switch runner := m["runner"].(type) {
		case string:
			scopeSlice = append(scopeSlice, runner)
		default:
			if cp.Scope == "" && cp.ParentCfgPlay != nil {
				scopeSlice = append(scopeSlice, cp.ParentCfgPlay.Scope)
			}
		}
		switch loopOn := m["loop_on"].(type) {
		case string:
			scopeSlice = append(scopeSlice, loopOn)
		}
	default:
		logrus.Fatalf("unexpected scope type %T value %v", scope, scope)
	}

	cp.Scope = strings.Join(scopeSlice, ":")
}

func (cp *CfgPlay) ParseMiddlewares(m map[string]interface{}, override bool) {
	if !override && cp.Middlewares != nil {
		return
	}

	app := cp.App

	switch v := m["middlewares"].(type) {
	case []interface{}:
		middlewares := make([]*middleware.Middleware, len(v))
		for i, middlewareI := range v {
			switch middlewareV := middlewareI.(type) {
			case string:
				middlewares[i] = &middleware.Middleware{
					Name:   middlewareV,
					Plugin: app.GetMiddleware(middlewareV),
				}
			case map[interface{}]interface{}, map[string]interface{}:
				middlewareMap, err := decode.ToMap(middlewareV)
				if err != nil {
					logrus.Fatalf("unexpected middleware type %T value %v, %v", middlewareV, middlewareV, err)
				}
				var name string
				switch v := middlewareMap["name"].(type) {
				case string:
					name = v
				case nil:
					logrus.Fatalf("missing middleware name in %v", middlewareMap)
				default:
					logrus.Fatalf("unexpected middleware name type %T value %v", v, v)
				}
				mr := &middleware.Middleware{
					Name:   name,
					Plugin: app.GetMiddleware(name),
				}
				if middlewareMap["vars"] != nil {
					varsI, err := decode.ToMap(middlewareMap["vars"])
					if err != nil {
						logrus.Fatalf("unexpected middleware vars type %T value %v, %v", middlewareMap["vars"], middlewareMap["vars"], err)
					}
					mr.Vars = variable.ParseVarsMap(varsI, cp.Depth+1)
				}
				middlewares[i] = mr
			default:
				logrus.Fatalf("unexpected middleware type %T value %v", middlewareI, middlewareI)
			}
		}
		cp.Middlewares = &middlewares
	case nil:
		if cp.ParentCfgPlay != nil {
			cp.Middlewares = cp.ParentCfgPlay.Middlewares
		}
	default:
		unexpectedTypeCmd(m, "middlewares")
	}
}

func (cp *CfgPlay) GetTitle() string {
	title := cp.Title
	if title == "" {
		title = cp.GetKey()
	}
	return title
}

func (cp *CfgPlay) GetKey() string {
	key := cp.Key
	if key == "" {
		key = strconv.Itoa(cp.Index)
	}
	return key
}

func (cp *CfgPlay) PromptPluginVars() {
	if cp.Loaders != nil {
		for _, lr := range *cp.Loaders {
			o := &sync.Once{}
			for _, v := range lr.Vars {
				v.OnPromptMessageOnce("🠶 loader "+lr.Name, o)
				v.PromptOnEmptyDefault()
				v.PromptOnEmptyValue()
				v.HandleRequired(nil, nil)
			}
		}
	}
	if cp.Middlewares != nil {
		for _, mr := range *cp.Middlewares {
			o := &sync.Once{}
			for _, v := range mr.Vars {
				v.OnPromptMessageOnce("🠶 middleware "+mr.Name, o)
				v.PromptOnEmptyDefault()
				v.PromptOnEmptyValue()
				v.HandleRequired(nil, nil)
			}
		}
	}
	if cp.Runner != nil {
		o := &sync.Once{}
		for _, v := range cp.Runner.Vars {
			v.OnPromptMessageOnce("🠶 runner "+cp.Runner.Name, o)
			v.PromptOnEmptyDefault()
			v.PromptOnEmptyValue()
			v.HandleRequired(nil, nil)
		}
	}
}

func (cp *CfgPlay) BuildRoot() *Play {
	ctx := CreateRunVars()
	rootPlay := CreatePlay(cp, ctx, nil)
	if rootPlay == nil {
		logrus.Fatal("no root play config found in current working directory")
	}

	nstk := rootPlay.GlobalRunCtx.NoSkipTreeKeys
	nstkl := len(nstk)
	for {
		rootPlay.LoadSkip()
		if l := len(nstk); nstkl != l {
			nstkl = l
		} else {
			break
		}
	}

	logrus.Infof(ansi.Color("≡ ", "green") + "collecting variables")
	rootPlay.LoadVars()

	return rootPlay
}

func unexpectedTypeCfgPlay(m map[string]interface{}, key string) {
	errors.UnexpectedType(m, key, "playbook")
}
