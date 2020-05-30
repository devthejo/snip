package play

import (
	"strconv"
	"strings"
	"sync"

	shellquote "github.com/kballard/go-shellquote"
	"github.com/mgutz/ansi"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/sirupsen/logrus"

	"gitlab.com/youtopia.earth/ops/snip/decode"
	"gitlab.com/youtopia.earth/ops/snip/errors"
)

type Play struct {
	App App

	ParentPlay *Play

	Title string

	Play interface{}

	Vars   map[string]*Var
	LoopOn []*Loop

	LoopSets       map[string]map[string]*Var
	LoopSequential *bool

	RegisterVars []string

	CheckCommand []string

	Dependencies []string
	PostInstall  []string

	Sudo *bool
	SSH  *bool

	State PlayStateType

	Depth       int
	HasChildren bool
}

type PlayStateType int

const (
	PlayStateReady PlayStateType = iota
	PlayStateUpdated
	PlayStateFailed
)

func CreatePlay(app App, m map[string]interface{}, parentPlay *Play) *Play {

	p := &Play{}

	p.Vars = make(map[string]*Var)
	p.LoopSets = make(map[string]map[string]*Var)

	p.App = app

	p.SetParentPlay(parentPlay)

	p.ParseMap(m)

	return p
}

func (p *Play) SetParentPlay(parentPlay *Play) {
	p.ParentPlay = parentPlay
	if parentPlay == nil {
		return
	}
	p.Depth = parentPlay.Depth + 1
}

func (p *Play) ParseMapAsDefault(m map[string]interface{}) {
	p.ParseTitle(m, false)
	p.ParseLoopSets(m, false)
	p.ParseLoopOn(m, false)
	p.ParseLoopSequential(m, false)
	p.ParseVars(m, false)
	p.ParseRegisterVars(m, false)
	p.ParseCheckCommand(m, false)
	p.ParseDependencies(m, false)
	p.ParsePostInstall(m, false)
	p.ParseSudo(m, false)
	p.ParseSSH(m, false)
	p.ParsePlay(m, false)
}

func (p *Play) ParseMap(m map[string]interface{}) {
	p.ParseTitle(m, true)
	p.ParseLoopSets(m, true)
	p.ParseLoopOn(m, true)
	p.ParseLoopSequential(m, true)
	p.ParseVars(m, true)
	p.ParseRegisterVars(m, true)
	p.ParseCheckCommand(m, true)
	p.ParseDependencies(m, true)
	p.ParsePostInstall(m, true)
	p.ParseSudo(m, true)
	p.ParseSSH(m, true)
	p.ParsePlay(m, true)
}

func (p *Play) ParsePlay(m map[string]interface{}, override bool) {
	if !override && p.Play != nil {
		return
	}
	switch v := m["play"].(type) {
	case []interface{}:
		playSlice := make([]*Play, len(v))
		for i, mPlay := range v {
			switch p2 := mPlay.(type) {
			case map[interface{}]interface{}:
				m = make(map[string]interface{}, len(p2))
				for k, v := range p2 {
					m[k.(string)] = v
				}
			case string:
				m = make(map[string]interface{})
				m["play"] = p2
			default:
				unexpectedTypePlay(m, "play")
			}

			playSlice[i] = CreatePlay(p.App, m, p)
		}
		p.Play = playSlice
		p.HasChildren = true

	case string:
		c, err := shellquote.Split(v)
		errors.Check(err)
		cmd := &Cmd{}
		cmd.Play = p
		cmd.Parse(c)
		p.Play = cmd
	case nil:
	default:
		unexpectedTypePlay(m, "play")
	}
}

func (p *Play) ParseTitle(m map[string]interface{}, override bool) {
	if !override && p.Title != "" {
		return
	}
	switch v := m["title"].(type) {
	case string:
		p.Title = v
	case nil:
	default:
		unexpectedTypeCmd(m, "title")
	}
}

func (p *Play) ParseLoopSets(m map[string]interface{}, override bool) {
	if p.ParentPlay != nil {
		for k, v := range p.ParentPlay.LoopSets {
			p.LoopSets[k] = v
		}
	}
	switch v := m["loop_sets"].(type) {
	case map[string]interface{}:
		loops, err := decode.ToMap(v)
		errors.Check(err)
		for loopKey, loopVal := range loops {
			switch loopV := loopVal.(type) {
			case map[string]interface{}:
				_, hk := p.LoopSets[loopKey]
				if !hk || override {
					p.LoopSets[loopKey] = ParsesVarsMap(loopV, p.Depth)
				}
			default:
				unexpectedTypeVarValue(loopKey, loopVal)
			}
		}
	case nil:
	default:
		unexpectedTypePlay(m, "loop_sets")
	}
}
func (p *Play) ParseLoopOn(m map[string]interface{}, override bool) {
	if p.LoopOn != nil && !override {
		return
	}
	switch v := m["loop_on"].(type) {
	case []interface{}:
		p.LoopOn = make([]*Loop, len(v))
		for loopI, loopV := range v {
			switch loop := loopV.(type) {
			case string:
				loop = strings.ToLower(loop)
				if p.LoopSets[loop] == nil {
					logrus.Fatalf("undefined LoopSet %v", loop)
				}
				p.LoopOn[loopI] = &Loop{
					Name: loop,
					Vars: p.LoopSets[loop],
				}
			case map[interface{}]interface{}:
				l, err := decode.ToMap(loop)
				errors.Check(err)
				p.LoopOn[loopI] = &Loop{
					Name: "item: " + strconv.Itoa(loopI),
					Vars: ParsesVarsMap(l, p.Depth),
				}
			case map[string]interface{}:
				p.LoopOn[loopI] = &Loop{
					Name: "item: " + strconv.Itoa(loopI),
					Vars: ParsesVarsMap(loop, p.Depth),
				}
			default:
				unexpectedTypeVarValue(strconv.Itoa(loopI), loopV)
			}
		}
	case nil:
	default:
		unexpectedTypePlay(m, "loop_on")
	}
}
func (p *Play) ParseLoopSequential(m map[string]interface{}, override bool) {
	switch v := m["loop_sequential"].(type) {
	case bool:
		if p.LoopSequential == nil || override {
			p.LoopSequential = &v
		}
	case nil:
	default:
		unexpectedTypePlay(m, "loop_sequential")
	}
}

func (p *Play) ParseVars(m map[string]interface{}, override bool) {
	switch v := m["vars"].(type) {
	case map[string]interface{}:
		m, err := decode.ToMap(v)
		errors.Check(err)
		for key, val := range ParsesVarsMap(m, p.Depth) {
			_, hk := p.Vars[key]
			logrus.Warnf("%v: %v -> %v", key, p.Vars[key], val)
			if override || !hk {
				p.Vars[key] = val
			}
		}
	case map[interface{}]interface{}:
		m, err := decode.ToMap(v)
		errors.Check(err)
		for key, val := range ParsesVarsMap(m, p.Depth) {
			_, hk := p.Vars[key]
			// logrus.Warnf("override %v", override)
			// logrus.Warnf("%v: %v -> %v", key, p.Vars[key], val)
			if override || !hk {
				p.Vars[key] = val
			}
		}
	case nil:
	default:
		unexpectedTypeCmd(m, "vars")
	}
}

func (p *Play) ParseRegisterVars(m map[string]interface{}, override bool) {
	if !override && p.RegisterVars == nil {
		return
	}
	switch v := m["register_vars"].(type) {
	case []interface{}:
		s, err := decode.ToStrings(v)
		errors.Check(err)
		p.RegisterVars = s
		for _, v := range p.RegisterVars {
			key := strings.ToLower(v)
			if p.Vars[key] == nil {
				p.Vars[key] = &Var{
					Name:  key,
					Depth: p.Depth,
				}
			}
		}
	case nil:
	default:
		unexpectedTypeCmd(m, "register_vars")
	}
}
func (p *Play) ParseCheckCommand(m map[string]interface{}, override bool) {
	if !override && p.CheckCommand != nil {
		return
	}
	switch v := m["check_command"].(type) {
	case string:
		s, err := shellquote.Split(v)
		errors.Check(err)
		p.CheckCommand = s
	case []interface{}:
		s, err := decode.ToStrings(v)
		errors.Check(err)
		p.CheckCommand = s
	case nil:
	default:
		unexpectedTypeCmd(m, "check_command")
	}
}
func (p *Play) ParseDependencies(m map[string]interface{}, override bool) {
	if !override && p.Dependencies == nil {
		return
	}
	switch m["dependencies"].(type) {
	case []interface{}:
		dependencies, err := decode.ToStrings(m["dependencies"])
		errors.Check(err)
		p.Dependencies = dependencies
	case nil:
	default:
		unexpectedTypeCmd(m, "dependencies")
	}
}

func (p *Play) ParsePostInstall(m map[string]interface{}, override bool) {
	if !override && p.PostInstall == nil {
		return
	}
	switch m["post_install"].(type) {
	case []interface{}:
		post_install, err := decode.ToStrings(m["post_install"])
		errors.Check(err)
		p.PostInstall = post_install
	case nil:
	default:
		unexpectedTypeCmd(m, "post_install")
	}
}

func (p *Play) ParseSudo(m map[string]interface{}, override bool) {
	if !override && p.Sudo != nil {
		return
	}
	switch s := m["sudo"].(type) {
	case bool:
		p.Sudo = &s
	case string:
		var b bool
		if s == "true" || s == "1" {
			b = true
		} else if s == "false" || s == "0" || s == "" {
			b = false
		} else {
			unexpectedTypeCmd(m, "sudo")
		}
		p.Sudo = &b
	case nil:
	default:
		unexpectedTypeCmd(m, "sudo")
	}
}

func (p *Play) ParseSSH(m map[string]interface{}, override bool) {
	if !override && p.SSH != nil {
		return
	}
	switch s := m["ssh"].(type) {
	case bool:
		p.SSH = &s
	case string:
		var b bool
		if s == "true" || s == "1" {
			b = true
		} else if s == "false" || s == "0" || s == "" {
			b = false
		} else {
			unexpectedTypeCmd(m, "ssh")
		}
		p.SSH = &b
	case nil:
	default:
		unexpectedTypeCmd(m, "ssh")
	}
}

func (p *Play) Run(ctx *RunCtx) {

	var icon string
	if p.ParentPlay == nil {
		icon = `🠞`
	} else if !p.HasChildren {
		icon = `⯈`
	} else {
		icon = `⤷`
	}

	logrus.Info(strings.Repeat("  ", p.Depth+1) + icon + " " + p.Title)

	runLoopSeq := func(loop *Loop) {
		vars := cmap.New()
		varsDefault := cmap.New()

		for k, v := range ctx.Vars.Items() {
			vars.Set(k, v)
		}
		for _, v := range p.Vars {
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
		for _, v := range p.Vars {
			v.RegisterDefaultTo(varsDefault)
			v.HandleRequired(varsDefault, vars)
		}

		runCtx := &RunCtx{
			Vars:        vars,
			VarsDefault: varsDefault,
			ReadyToRun:  ctx.ReadyToRun,
		}

		switch pl := p.Play.(type) {
		case []*Play:
			for _, child := range pl {
				child.Run(runCtx)
			}
		case *Cmd:
			pl.Run(runCtx)
		}
	}

	var wg sync.WaitGroup
	var runLoop func(loop *Loop)

	var loopSequential bool
	if p.LoopSequential != nil {
		loopSequential = *p.LoopSequential
	}

	if loopSequential || !ctx.ReadyToRun {
		runLoop = runLoopSeq
	} else {
		runLoop = func(loop *Loop) {
			wg.Add(1)
			go func() {
				defer wg.Done()
				runLoopSeq(loop)
			}()
		}
	}

	var loops []*Loop
	if len(p.LoopOn) == 0 {
		loops = append(loops, &Loop{
			Name: "",
			Vars: make(map[string]*Var),
		})
	} else {
		loops = p.LoopOn
	}

	for _, loop := range loops {
		runLoop(loop)
	}
	wg.Wait()

}

func (p *Play) Start() {
	ctx := CreateRunCtx()

	logrus.Infof(ansi.Color("≡ ", "green") + "collecting variables")
	p.Run(ctx)

	ctx.ReadyToRun = true
	logrus.Infof("🚀 running playbook")
	p.Run(ctx)

}

func unexpectedTypePlay(m map[string]interface{}, key string) {
	errors.UnexpectedType(m, key, "playbook")
}
