package play

import (
	"strconv"
	"strings"
	"sync"
	"time"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/sirupsen/logrus"
	"go.uber.org/multierr"

	"gitlab.com/youtopia.earth/ops/snip/errors"
	"gitlab.com/youtopia.earth/ops/snip/registry"
	"gitlab.com/youtopia.earth/ops/snip/variable"
)

type Play struct {
	App App

	RunCtx *RunCtx

	ParentLoopRow *LoopRow

	Index int
	Key   string
	Title string

	TreeKeyParts []string

	LoopRow []*LoopRow

	// Vars map[string]*variable.Var

	LoopSequential bool

	ExecTimeout *time.Duration

	RegisterVars map[string]*registry.VarDef

	CheckCommand []string

	Dependencies []string
	PostInstall  []string

	Middlewares []string

	Depth       int
	HasChildren bool

	State StateType
}

func CreatePlay(cp *CfgPlay, ctx *RunCtx, parentLoopRow *LoopRow) *Play {
	var loopSequential bool
	if cp.LoopSequential != nil {
		loopSequential = *cp.LoopSequential
	}

	registerVars := make(map[string]*registry.VarDef)
	for k, v := range cp.RegisterVars {
		registerVars[k] = v
	}

	p := &Play{
		App: cp.App,

		Index: cp.Index,
		Key:   cp.Key,
		Title: cp.Title,

		LoopSequential: loopSequential,
		CheckCommand:   cp.CheckCommand,

		ExecTimeout: cp.ExecTimeout,

		RegisterVars: registerVars,
		// Dependencies: ,
		// PostInstall: ,

		Depth:       cp.Depth,
		HasChildren: cp.HasChildren,
	}

	p.RunCtx = ctx
	p.ParentLoopRow = parentLoopRow

	p.TreeKeyParts = GetTreeKeyParts(p)

	var icon string
	if cp.ParentCfgPlay == nil {
		icon = `🠞`
	} else if !cp.HasChildren {
		icon = `⯈`
	} else {
		icon = `⤷`
	}

	logrus.Info(strings.Repeat("  ", cp.Depth+1) + icon + " " + cp.GetTitle())

	var loopRows []*CfgLoopRow
	if len(cp.LoopOn) == 0 {
		loopRows = append(loopRows, &CfgLoopRow{
			Name:          "",
			Key:           "",
			Index:         0,
			Vars:          make(map[string]*variable.Var),
			IsLoopRowItem: false,
		})
	} else {
		loopRows = cp.LoopOn
	}

	p.LoopRow = make([]*LoopRow, len(loopRows))
	for i, cfgLoopRow := range loopRows {

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

		cp.PromptPluginVars()

		runCtx := &RunCtx{
			Vars:        vars,
			VarsDefault: varsDefault,
		}

		switch pl := cp.CfgPlay.(type) {
		case []*CfgPlay:
			sp := make([]*Play, len(pl))
			for i, child := range pl {
				sp[i] = CreatePlay(child, runCtx, loop)
			}
			loop.Play = sp
		case *CfgCmd:
			loop.Play = CreateCmd(pl, runCtx, loop)
		}
	}

	return p

}

func (p *Play) GetTitle() string {
	title := p.Title
	if title == "" {
		title = p.GetKey()
	}
	return title
}

func (p *Play) GetKey() string {
	key := p.Key
	if key == "" {
		key = strconv.Itoa(p.Index)
	}
	return key
}

func (p *Play) RegisterVarsSaveUpAndPersist() {
	kp := p.TreeKeyParts
	if len(p.RegisterVars) <= 0 || len(kp) < 3 {
		return
	}
	varsRegistry := p.App.GetVarsRegistry()
	dp := kp[0 : len(kp)-2]
	for _, vr := range p.RegisterVars {
		if !vr.Enable {
			continue
		}
		value := varsRegistry.GetVarBySlice(kp, vr.Key)
		if value == "" && vr.Persist {
			for i := len(kp); i >= 0; i-- {
				dp2 := kp[0:i]
				value = varsRegistry.PersistGetVarBySlice(dp2, vr.Key)
				if value != "" {
					break
				}
			}
		}
		if value != "" {
			varsRegistry.SetVarBySlice(dp, vr.Key, value)
			if vr.Persist {
				varsRegistry.PersistSetVarBySlice(dp, vr.Key, value)
			}
		}
	}
}

func (p *Play) Run() error {

	var icon string
	if p.ParentLoopRow == nil {
		icon = `🠞`
	} else if !p.HasChildren {
		icon = `⯈`
	} else {
		icon = `⤷`
	}

	logrus.Info(strings.Repeat("  ", p.Depth+1) + icon + " " + p.GetTitle())

	var errSlice []error

	runLoopSeq := func(loop *LoopRow) error {
		if loop.IsLoopRowItem {
			logrus.Info(strings.Repeat("  ", p.Depth+2) + "⦿ " + loop.Name)
		}

		switch pl := loop.Play.(type) {
		case []*Play:
			for _, child := range pl {
				if err := child.Run(); err != nil {
					errSlice = append(errSlice, err)
					break
				}
			}
		case *Cmd:
			pl.RegisterVarsLoad()

			if err := pl.Run(); err != nil {
				errSlice = append(errSlice, err)
			}
		}

		if len(errSlice) > 0 {
			return multierr.Combine(errSlice...)
		}
		return nil
	}

	var wg sync.WaitGroup
	var runLoopRow func(loop *LoopRow) error

	if p.LoopSequential {
		runLoopRow = runLoopSeq
	} else {
		runLoopRow = func(loop *LoopRow) error {
			wg.Add(1)
			go func() {
				defer wg.Done()
				runLoopSeq(loop)
			}()
			return nil
		}
	}

	for _, loop := range p.LoopRow {
		if err := runLoopRow(loop); err != nil {
			break
		}
	}
	wg.Wait()

	if len(errSlice) > 0 {
		return multierr.Combine(errSlice...)
	}

	p.RegisterVarsSaveUpAndPersist()

	return nil

}

func (p *Play) Start() error {
	logrus.Infof("🚀 running playbook")
	return p.Run()
}

func unexpectedTypePlay(m map[string]interface{}, key string) {
	errors.UnexpectedType(m, key, "playbook")
}
