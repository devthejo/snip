package play

import (
	"strconv"
	"strings"
	"time"

	shellquote "github.com/kballard/go-shellquote"
	"github.com/mgutz/ansi"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/sirupsen/logrus"

	"gitlab.com/youtopia.earth/ops/snip/decode"
	"gitlab.com/youtopia.earth/ops/snip/errors"
)

type CfgPlay struct {
	App App

	ParentCfgPlay *CfgPlay

	Index int
	Key   string
	Title string

	CfgPlay interface{}

	Vars   map[string]*Var
	LoopOn []*CfgLoopRow

	LoopSets       map[string]map[string]*Var
	LoopSequential *bool

	RegisterVars []string

	CheckCommand []string

	Dependencies []string
	PostInstall  []string

	Depth       int
	HasChildren bool

	Dir string

	ExecTimeout *time.Duration

	Loader      interface{}
	Middlewares *[]string
	Runner      string
}

func CreateCfgPlay(app App, m map[string]interface{}, parentCfgPlay *CfgPlay) *CfgPlay {

	cp := &CfgPlay{}

	cp.Vars = make(map[string]*Var)
	cp.LoopSets = make(map[string]map[string]*Var)

	cp.App = app

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
	cp.ParseLoopSets(m, override)
	cp.ParseLoopOn(m, override)
	cp.ParseLoopSequential(m, override)
	cp.ParseVars(m, override)
	cp.ParseRegisterVars(m, override)
	cp.ParseCheckCommand(m, override)
	cp.ParseDependencies(m, override)
	cp.ParsePostInstall(m, override)
	cp.ParseLoader(m, override)
	cp.ParseMiddlewares(m, override)
	cp.ParseRunner(m, override)
	cp.ParsePlay(m, override)
}

func (cp *CfgPlay) ParsePlay(m map[string]interface{}, override bool) {
	if !override && cp.CfgPlay != nil {
		return
	}
	switch v := m["play"].(type) {
	case []interface{}:
		playSlice := make([]*CfgPlay, len(v))
		for i, mCfgPlay := range v {
			switch p2 := mCfgPlay.(type) {
			case map[interface{}]interface{}:
				m = make(map[string]interface{}, len(p2))
				for k, v := range p2 {
					m[k.(string)] = v
				}
			case string:
				m = make(map[string]interface{})
				m["play"] = p2
			default:
				unexpectedTypeCfgPlay(m, "play")
			}

			pI := CreateCfgPlay(cp.App, m, cp)
			pI.Index = i
			playSlice[i] = pI
		}
		cp.CfgPlay = playSlice
		cp.HasChildren = true

	case string:
		c, err := shellquote.Split(v)
		errors.Check(err)
		cp.CfgPlay = CreateCfgCmd(cp, c)
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
		cp.Dir = v
	case nil:
	default:
		unexpectedTypeCmd(m, "dir")
	}
}

func (cp *CfgPlay) ParseExecTimeout(m map[string]interface{}, override bool) {
	if !override && cp.ExecTimeout != nil {
		return
	}
	timeout, err := decode.Duration(m["timeout"])
	errors.Check(err)
	if timeout != 0 {
		cp.ExecTimeout = &timeout
	}
}

func (cp *CfgPlay) ParseLoopSets(m map[string]interface{}, override bool) {
	if cp.ParentCfgPlay != nil {
		for k, v := range cp.ParentCfgPlay.LoopSets {
			cp.LoopSets[k] = v
		}
	}
	switch v := m["loop_sets"].(type) {
	case map[string]interface{}:
		loops, err := decode.ToMap(v)
		errors.Check(err)
		for loopKey, loopVal := range loops {
			switch loopV := loopVal.(type) {
			case map[string]interface{}:
				_, hk := cp.LoopSets[loopKey]
				if !hk || override {
					cp.LoopSets[loopKey] = ParsesVarsMap(loopV, cp.Depth)
				}
			default:
				unexpectedTypeVarValue(loopKey, loopVal)
			}
		}
	case nil:
	default:
		unexpectedTypeCfgPlay(m, "loop_sets")
	}
}
func (cp *CfgPlay) ParseLoopOn(m map[string]interface{}, override bool) {
	if cp.LoopOn != nil && !override {
		return
	}
	switch v := m["loop_on"].(type) {
	case []interface{}:
		cp.LoopOn = make([]*CfgLoopRow, len(v))
		for loopI, loopV := range v {
			var cfgLoopRow *CfgLoopRow
			switch loop := loopV.(type) {
			case string:
				loop = strings.ToLower(loop)
				if cp.LoopSets[loop] == nil {
					logrus.Fatalf("undefined LoopSet %v", loop)
				}
				cfgLoopRow = CreateCfgLoopRow(loopI, loop, cp.LoopSets[loop])
			case map[interface{}]interface{}:
				l, err := decode.ToMap(loop)
				errors.Check(err)
				cfgLoopRow = CreateCfgLoopRow(loopI, "", ParsesVarsMap(l, cp.Depth))
			case map[string]interface{}:
				cfgLoopRow = CreateCfgLoopRow(loopI, "", ParsesVarsMap(loop, cp.Depth))
			default:
				unexpectedTypeVarValue(strconv.Itoa(loopI), loopV)
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
	case map[string]interface{}:
		m, err := decode.ToMap(v)
		errors.Check(err)
		for key, val := range ParsesVarsMap(m, cp.Depth) {
			_, hk := cp.Vars[key]
			if override || !hk {
				cp.Vars[key] = val
			}
		}
	case map[interface{}]interface{}:
		m, err := decode.ToMap(v)
		errors.Check(err)
		for key, val := range ParsesVarsMap(m, cp.Depth) {
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

func (cp *CfgPlay) ParseRegisterVars(m map[string]interface{}, override bool) {
	if !override && cp.RegisterVars != nil {
		return
	}
	switch v := m["register_vars"].(type) {
	case []interface{}:
		s, err := decode.ToStrings(v)
		errors.Check(err)
		cp.RegisterVars = s
		for _, v := range cp.RegisterVars {
			key := strings.ToLower(v)
			if cp.Vars[key] == nil {
				cp.Vars[key] = &Var{
					Name:  key,
					Depth: cp.Depth,
				}
			}
		}
	case nil:
	default:
		unexpectedTypeCmd(m, "register_vars")
	}
}
func (cp *CfgPlay) ParseCheckCommand(m map[string]interface{}, override bool) {
	if !override && cp.CheckCommand != nil {
		return
	}
	switch v := m["check_command"].(type) {
	case string:
		s, err := shellquote.Split(v)
		errors.Check(err)
		cp.CheckCommand = s
	case []interface{}:
		s, err := decode.ToStrings(v)
		errors.Check(err)
		cp.CheckCommand = s
	case nil:
	default:
		unexpectedTypeCmd(m, "check_command")
	}
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
	default:
		unexpectedTypeCmd(m, "post_install")
	}
}

func (cp *CfgPlay) ParseLoader(m map[string]interface{}, override bool) {
	if !override && cp.Loader != nil {
		return
	}
	switch v := m["loader"].(type) {
	case string:
		cp.Loader = v
	case []string:
		s, err := decode.ToStrings(v)
		errors.Check(err)
		cp.Loader = s
	case nil:
	default:
		unexpectedTypeCmd(m, "loader")
	}
}
func (cp *CfgPlay) ParseRunner(m map[string]interface{}, override bool) {
	if !override && cp.Runner != "" {
		return
	}
	switch v := m["runner"].(type) {
	case string:
		cp.Runner = v
	case nil:
		if cp.ParentCfgPlay != nil {
			cp.Runner = cp.ParentCfgPlay.Runner
		}
	default:
		unexpectedTypeCmd(m, "runner")
	}
}
func (cp *CfgPlay) ParseMiddlewares(m map[string]interface{}, override bool) {
	if !override && cp.Middlewares != nil {
		return
	}
	switch m["middlewares"].(type) {
	case []interface{}:
		middlewares, err := decode.ToStrings(m["middlewares"])
		errors.Check(err)
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

func (cp *CfgPlay) BuildRoot() *Play {
	logrus.Infof(ansi.Color("≡ ", "green") + "collecting variables")
	ctx := &RunCtx{
		Vars:        cmap.New(),
		VarsDefault: cmap.New(),
	}
	return CreatePlay(cp, ctx, nil)
}

func unexpectedTypeCfgPlay(m map[string]interface{}, key string) {
	errors.UnexpectedType(m, key, "playbook")
}
