package play

import (
	"strconv"
	"strings"
	"sync"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/sirupsen/logrus"
	"go.uber.org/multierr"

	"gitlab.com/youtopia.earth/ops/snip/errors"
)

type Play struct {
	App App

	RunCtx *RunCtx

	ParentLoopRow *LoopRow

	Index int
	Key   string
	Title string

	LoopRow []*LoopRow

	Vars map[string]*Var

	LoopSequential bool

	RegisterVars []string

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

	p := &Play{

		Index: cp.Index,
		Key:   cp.Key,
		Title: cp.Title,

		LoopSequential: loopSequential,
		CheckCommand:   cp.CheckCommand,

		// RegisterVars: cp.RegisterVars,
		// Dependencies: ,
		// PostInstall: ,

		Middlewares: cp.Middlewares,

		Depth:       cp.Depth,
		HasChildren: cp.HasChildren,
	}

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

	runLoopSeq := func(loop *LoopRow) {
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
			if err := pl.Run(); err != nil {
				errSlice = append(errSlice, err)
			}
		}
	}

	var wg sync.WaitGroup
	var runLoopRow func(loop *LoopRow)

	if p.LoopSequential {
		runLoopRow = runLoopSeq
	} else {
		runLoopRow = func(loop *LoopRow) {
			wg.Add(1)
			go func() {
				defer wg.Done()
				runLoopSeq(loop)
			}()
		}
	}

	for _, loop := range p.LoopRow {
		runLoopRow(loop)
	}
	wg.Wait()

	if len(errSlice) > 0 {
		return multierr.Combine(errSlice...)
	}
	return nil

}

func (p *Play) Start() error {
	logrus.Infof("🚀 running playbook")
	return p.Run()
}

func unexpectedTypePlay(m map[string]interface{}, key string) {
	errors.UnexpectedType(m, key, "playbook")
}
