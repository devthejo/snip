package play

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	shellquote "github.com/kballard/go-shellquote"
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
	LoopSequential bool

	RegisterVars []string

	CheckCommand []string

	Dependencies []string
	PostInstall  []string

	Sudo bool
	SSH  bool

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
	for k, v := range parentPlay.LoopSets {
		p.LoopSets[k] = v
	}
}

func (p *Play) ParseMap(m map[string]interface{}) {
	p.ParsePlayCmd(m)
	p.ParseTitle(m)
	p.ParseLoopSets(m)
	p.ParseLoopOn(m)
	p.ParseLoopSequential(m)
	p.ParseVars(m)
	p.ParseRegisterVars(m)
	p.ParseCheckCommand(m)
	p.ParseDependencies(m)
	p.ParsePostInstall(m)
	p.ParseSudo(m)
	p.ParseSSH(m)
	p.ParsePlayChildren(m)
}

func (p *Play) ParsePlayCmd(m map[string]interface{}) {
	switch m["play"].(type) {
	case string:
		p.ParsePlay(m)
	}
}
func (p *Play) ParsePlayChildren(m map[string]interface{}) {
	switch m["play"].(type) {
	case []interface{}:
		p.ParsePlay(m)
	}
}
func (p *Play) ParsePlay(m map[string]interface{}) {
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

func (p *Play) ParseTitle(m map[string]interface{}) {
	switch v := m["title"].(type) {
	case string:
		p.Title = v
	case nil:
	default:
		unexpectedTypeCmd(m, "title")
	}
}

func (p *Play) ParseLoopSets(m map[string]interface{}) {
	switch v := m["loop_sets"].(type) {
	case map[string]interface{}:
		loops, err := decode.ToMap(v)
		errors.Check(err)
		for loopKey, loopVal := range loops {
			switch loopV := loopVal.(type) {
			case map[string]interface{}:
				p.LoopSets[loopKey] = ParsesVarsMap(loopV, p.Depth)
			default:
				unexpectedTypeVarValue(loopKey, loopVal)
			}
		}
	case nil:
	default:
		unexpectedTypePlay(m, "loop_sets")
	}
}
func (p *Play) ParseLoopOn(m map[string]interface{}) {
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
func (p *Play) ParseLoopSequential(m map[string]interface{}) {
	switch v := m["loop_sequential"].(type) {
	case bool:
		p.LoopSequential = v
	case nil:
	default:
		unexpectedTypePlay(m, "loop_sequential")
	}
}

func (p *Play) ParseVars(m map[string]interface{}) {
	switch v := m["vars"].(type) {
	case map[string]interface{}:
		m, err := decode.ToMap(v)
		errors.Check(err)
		for key, val := range ParsesVarsMap(m, p.Depth) {
			p.Vars[key] = val
		}
	case map[interface{}]interface{}:
		m, err := decode.ToMap(v)
		errors.Check(err)
		for key, val := range ParsesVarsMap(m, p.Depth) {
			p.Vars[key] = val
		}
	case nil:
	default:
		unexpectedTypeCmd(m, "vars")
	}
}

func (p *Play) ParseRegisterVars(m map[string]interface{}) {
	switch v := m["register_vars"].(type) {
	case []interface{}:
		s, err := decode.ToStrings(v)
		errors.Check(err)
		p.RegisterVars = append(p.RegisterVars, s...)
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
func (p *Play) ParseCheckCommand(m map[string]interface{}) {
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
func (p *Play) ParseDependencies(m map[string]interface{}) {
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

func (p *Play) ParsePostInstall(m map[string]interface{}) {
	switch m["postInstall"].(type) {
	case []interface{}:
		postInstall, err := decode.ToStrings(m["postInstall"])
		errors.Check(err)
		p.PostInstall = postInstall
	case nil:
	default:
		unexpectedTypeCmd(m, "postInstall")
	}
}

func (p *Play) ParseSudo(m map[string]interface{}) {
	switch s := m["sudo"].(type) {
	case bool:
		p.Sudo = s
	case string:
		if s == "true" || s == "1" {
			p.Sudo = true
		} else if s == "false" || s == "0" || s == "" {
			p.Sudo = false
		} else {
			unexpectedTypeCmd(m, "sudo")
		}
	case nil:
	default:
		unexpectedTypeCmd(m, "sudo")
	}
}

func (p *Play) ParseSSH(m map[string]interface{}) {
	switch s := m["ssh"].(type) {
	case bool:
		p.SSH = s
	case string:
		if s == "true" || s == "1" {
			p.SSH = true
		} else if s == "false" || s == "0" || s == "" {
			p.SSH = false
		} else {
			unexpectedTypeCmd(m, "ssh")
		}
	case nil:
	default:
		unexpectedTypeCmd(m, "ssh")
	}
}

func (p *Play) Run(parentVars cmap.ConcurrentMap, parentVarsDefault cmap.ConcurrentMap) {
	if parentVars == nil {
		parentVars = cmap.New()
	}
	if parentVarsDefault == nil {
		parentVarsDefault = cmap.New()
	}

	var icon string
	if p.ParentPlay == nil {
		icon = `🠞`
	} else if !p.HasChildren {
		icon = `⯈`
	} else {
		icon = `⤷`
	}
	fmt.Println(strings.Repeat("  ", p.Depth)+icon, p.Title)

	runLoopSeq := func(loop *Loop) {
		vars := cmap.New()
		varsDefault := cmap.New()

		for k, v := range parentVars.Items() {
			vars.Set(k, v)
		}
		for _, v := range p.Vars {
			v.RegisterValueTo(vars)
		}
		for _, v := range loop.Vars {
			v.RegisterValueTo(vars)
		}

		for k, v := range parentVarsDefault.Items() {
			varsDefault.Set(k, v)
		}
		for _, v := range p.Vars {
			v.RegisterDefaultTo(varsDefault)
			v.HandleRequired(varsDefault, vars)
		}
		for _, v := range loop.Vars {
			v.RegisterDefaultTo(varsDefault)
			v.HandleRequired(varsDefault, vars)
		}

		switch pl := p.Play.(type) {
		case []*Play:
			for _, child := range pl {
				child.Run(vars, varsDefault)
			}
		case *Cmd:
			pl.Run(vars, varsDefault)
		}
	}

	var wg sync.WaitGroup
	var runLoop func(loop *Loop)
	if p.LoopSequential {
		runLoop = runLoopSeq
	} else {
		runLoop = runLoopSeq
		// runLoop = func(loop *Loop) {
		// 	wg.Add(1)
		// 	go func() {
		// 		defer wg.Done()
		// 		runLoopSeq(loop)
		// 	}()
		// }
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

func unexpectedTypePlay(m map[string]interface{}, key string) {
	errors.UnexpectedType(m, key, "playbook")
}
