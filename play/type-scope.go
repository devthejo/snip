package play

type Scope struct {
	Plays []*Play

	LoopSets     map[string]map[string]*Var
	Vars         map[string]*Var
	RegisterVars []string
}

func CreateScope(p *Play) *Scope {
	pscope := p.Scope

	loopSets := make(map[string]map[string]*Var)
	for k, v := range pscope.LoopSets {
		loopSets[k] = make(map[string]*Var)
		for k2, v2 := range v {
			loopSets[k][k2] = v2
		}
	}
	for k, v := range p.LoopSets {
		loopSets[k] = make(map[string]*Var)
		for k2, v2 := range v {
			loopSets[k][k2] = v2
		}
	}

	vars := make(map[string]*Var)
	for k, v := range pscope.Vars {
		vars[k] = v
	}
	for k, v := range p.Vars {
		vars[k] = v
	}

	var registerVars []string
	for _, v := range pscope.RegisterVars {
		registerVars = append(registerVars, v)
	}
	for _, v := range p.RegisterVars {
		registerVars = append(registerVars, v)
	}

	scope := &Scope{
		Plays:        append(p.Scope.Plays, p),
		LoopSets:     loopSets,
		Vars:         vars,
		RegisterVars: registerVars,
	}
	return scope
}
