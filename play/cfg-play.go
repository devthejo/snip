package play

import (
	"strconv"
	"strings"

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

	Sudo *bool
	SSH  *bool

	Depth       int
	HasChildren bool
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
	cp.ParseLoopSets(m, override)
	cp.ParseLoopOn(m, override)
	cp.ParseLoopSequential(m, override)
	cp.ParseVars(m, override)
	cp.ParseRegisterVars(m, override)
	cp.ParseCheckCommand(m, override)
	cp.ParseDependencies(m, override)
	cp.ParsePostInstall(m, override)
	cp.ParseSudo(m, override)
	cp.ParseSSH(m, override)
	cp.ParseCfgPlay(m, override)
}

func (cp *CfgPlay) ParseCfgPlay(m map[string]interface{}, override bool) {
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
		cmd := &CfgCmd{}
		cmd.CfgPlay = cp
		cmd.Depth = cp.Depth + 1
		cmd.Parse(c)
		cp.CfgPlay = cmd
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
		cp.Key = strings.ReplaceAll(v, "|", "-")
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
	if !override && cp.RegisterVars == nil {
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
	if !override && cp.Dependencies == nil {
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
	if !override && cp.PostInstall == nil {
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

func (cp *CfgPlay) ParseSudo(m map[string]interface{}, override bool) {
	if !override && cp.Sudo != nil {
		return
	}
	switch s := m["sudo"].(type) {
	case bool:
		cp.Sudo = &s
	case string:
		var b bool
		if s == "true" || s == "1" {
			b = true
		} else if s == "false" || s == "0" || s == "" {
			b = false
		} else {
			unexpectedTypeCmd(m, "sudo")
		}
		cp.Sudo = &b
	case nil:
	default:
		unexpectedTypeCmd(m, "sudo")
	}
}

func (cp *CfgPlay) ParseSSH(m map[string]interface{}, override bool) {
	if !override && cp.SSH != nil {
		return
	}
	switch s := m["ssh"].(type) {
	case bool:
		cp.SSH = &s
	case string:
		var b bool
		if s == "true" || s == "1" {
			b = true
		} else if s == "false" || s == "0" || s == "" {
			b = false
		} else {
			unexpectedTypeCmd(m, "ssh")
		}
		cp.SSH = &b
	case nil:
	default:
		unexpectedTypeCmd(m, "ssh")
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

func CreatePlayFromCfgPlay(cp *CfgPlay) *Play {
	var loopSequential bool
	if cp.LoopSequential != nil {
		loopSequential = *cp.LoopSequential
	}

	var sudo bool
	if cp.Sudo != nil {
		sudo = *cp.Sudo
	}

	var ssh bool
	if cp.SSH != nil {
		ssh = *cp.SSH
	}

	p := &Play{

		Index: cp.Index,
		Key:   cp.Key,
		Title: cp.Title,

		LoopSequential: loopSequential,
		CheckCommand:   cp.CheckCommand,

		// RegisterVars: cp.RegisterVars,
		// Dependencies: ,
		// PostInstall: ,

		Sudo: sudo,
		SSH:  ssh,

		Depth:       cp.Depth,
		HasChildren: cp.HasChildren,
	}
	return p
}

func (cp *CfgPlay) Build(ctx *RunCtx, parentLoopRow *LoopRow) *Play {
	p := CreatePlayFromCfgPlay(cp)
	p.RunCtx = ctx
	p.ParentLoopRow = parentLoopRow

	var icon string
	if cp.ParentCfgPlay == nil {
		icon = `🠞`
	} else if !cp.HasChildren {
		icon = `⯈`
	} else {
		icon = `⤷`
	}

	logrus.Info(strings.Repeat("  ", cp.Depth+1) + icon + " " + cp.GetTitle())

	var loops []*CfgLoopRow
	if len(cp.LoopOn) == 0 {
		loops = append(loops, &CfgLoopRow{
			Name:          "",
			Key:           "",
			Index:         0,
			Vars:          make(map[string]*Var),
			IsLoopRowItem: false,
		})
	} else {
		loops = cp.LoopOn
	}

	p.LoopRow = make([]*LoopRow, len(loops))
	for i, cfgLoopRow := range loops {
		loop := &LoopRow{
			Name:          cfgLoopRow.Name,
			Key:           cfgLoopRow.Key,
			Index:         cfgLoopRow.Index,
			Vars:          cfgLoopRow.Vars,
			IsLoopRowItem: cfgLoopRow.IsLoopRowItem,
			ParentPlay:    p,
		}
		p.LoopRow[i] = loop

		if loop.IsLoopRowItem {
			logrus.Info(strings.Repeat("  ", cp.Depth+2) + "⦿ " + loop.Name)
		}

		vars := cmap.New()
		varsDefault := cmap.New()

		for k, v := range ctx.Vars.Items() {
			vars.Set(k, v)
		}
		for _, v := range cp.Vars {
			v.RegisterValueTo(vars)
		}
		for _, v := range loop.Vars {
			v.RegisterValueTo(vars)
		}

		for k, v := range ctx.VarsDefault.Items() {
			varsDefault.Set(k, v)
		}
		for _, v := range loop.Vars {
			v.RegisterDefaultTo(varsDefault)
			v.HandleRequired(varsDefault, vars)
		}
		for _, v := range cp.Vars {
			v.RegisterDefaultTo(varsDefault)
			v.HandleRequired(varsDefault, vars)
		}

		runCtx := &RunCtx{
			Vars:        vars,
			VarsDefault: varsDefault,
		}

		switch pl := cp.CfgPlay.(type) {
		case []*CfgPlay:
			sp := make([]*Play, len(pl))
			for i, child := range pl {
				sp[i] = child.Build(runCtx, loop)
			}
			loop.Play = sp
		case *CfgCmd:
			loop.Play = pl.Build(runCtx, loop)
		}
	}

	return p

}

func (cp *CfgPlay) BuildRoot() *Play {
	logrus.Infof(ansi.Color("≡ ", "green") + "collecting variables")
	ctx := &RunCtx{
		Vars:        cmap.New(),
		VarsDefault: cmap.New(),
	}
	return cp.Build(ctx, nil)
}

func unexpectedTypeCfgPlay(m map[string]interface{}, key string) {
	errors.UnexpectedType(m, key, "playbook")
}
