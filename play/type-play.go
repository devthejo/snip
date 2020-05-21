package play

import (
	"strconv"

	shellquote "github.com/kballard/go-shellquote"
	"gitlab.com/youtopia.earth/ops/snip/decode"
	"gitlab.com/youtopia.earth/ops/snip/errors"
)

type Play struct {
	App App

	Scope *Scope

	Title string

	Play interface{}

	LoopSets       map[string]map[string]*Var
	LoopOn         []interface{}
	LoopSequential bool

	Vars         map[string]*Var
	RegisterVars []string

	CheckCommand []string

	Dependencies []string
	PostInstall  []string

	Sudo bool
	SSH  bool

	State PlayStateType
}

type PlayStateType int

const (
	PlayStateReady PlayStateType = iota
	PlayStateUpdated
	PlayStateFailed
)

func ParsePlay(app App, m map[string]interface{}, scope *Scope) *Play {
	p := &Play{}

	p.App = app

	if scope == nil {
		scope = &Scope{
			Plays: []*Play{p},
		}
	}
	p.Scope = scope

	p.ParseMap(m)

	return p
}

func (p *Play) ParseMap(m map[string]interface{}) {
	p.ParsePlay(m)
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
}

func (p *Play) ParsePlay(m map[string]interface{}) {
	switch v := m["play"].(type) {
	case []interface{}:
		playSlice := make([]*Play, len(v))
		for i, mPlay := range v {
			m, err := decode.ToMap(mPlay)
			errors.Check(err)
			scope := CreateScope(p)
			playSlice[i] = ParsePlay(p.App, m, scope)
		}
		p.Play = playSlice

	case string:
		c, err := shellquote.Split(v)
		errors.Check(err)
		p.ParsePlayCmd(c)
	case nil:
	default:
		unexpectedTypePlay(m, "play")
	}
}

func (p *Play) ParsePlayCmd(c []string) {
	cmd := &Cmd{}
	cmd.Play = p
	cmd.Parse(c)
	p.Play = cmd
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
	p.LoopSets = make(map[string]map[string]*Var)
	switch v := m["loop_sets"].(type) {
	case map[string]interface{}:
		loops, err := decode.ToMap(v)
		errors.Check(err)
		for loopKey, loopVal := range loops {
			switch loopV := loopVal.(type) {
			case map[string]interface{}:
				p.LoopSets[loopKey] = ParsesVarsMap(loopV)
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
		p.LoopOn = make([]interface{}, len(v))
		for loopI, loopV := range v {
			switch loop := loopV.(type) {
			case string:
				p.LoopOn[loopI] = p.Scope.LoopSets[loop]
			case map[interface{}]interface{}:
				l, err := decode.ToMap(loop)
				errors.Check(err)
				p.LoopOn[loopI] = ParsesVarsMap(l)
			case map[string]interface{}:
				p.LoopOn[loopI] = ParsesVarsMap(loop)
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
		p.Vars = ParsesVarsMap(m)
	case map[interface{}]interface{}:
		m, err := decode.ToMap(v)
		errors.Check(err)
		p.Vars = ParsesVarsMap(m)
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
		p.RegisterVars = s
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

func (p *Play) PromptVars(varsMap map[string]string) {

	if varsMap == nil {
		varsMap = make(map[string]string)
	}

	vars := p.Vars
	for _, v := range vars {
		var currentVal string
		if v.Required {
			if vars[v.Name] != nil {
				currentVal = vars[v.Name].Default
			} else if v.Default != "" {
				currentVal = v.Default
			}
		}
		if v.ForcePrompt || (v.Required && currentVal == "" && v.DefaultFromVar == "") {
			if varsMap[v.Name] != "" {
				v.PromptAnswer = varsMap[v.Name]
				continue
			}
			PromptVar(v)
		}

		varsMap[v.Name] = v.PromptAnswer

	}

	switch pSlice := p.Play.(type) {
	case []*Play:
		for _, p2 := range pSlice {
			p2.PromptVars(varsMap)
		}
	}

}

func unexpectedTypePlay(m map[string]interface{}, key string) {
	errors.UnexpectedType(m, key, "playbook")
}
