package play

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"go.uber.org/multierr"

	"gitlab.com/ytopia/ops/snip/config"
	"gitlab.com/ytopia/ops/snip/errors"
	"gitlab.com/ytopia/ops/snip/registry"
	"gitlab.com/ytopia/ops/snip/tools"
	"gitlab.com/ytopia/ops/snip/variable"
)

type Play struct {
	App App

	RunCtx *RunCtx

	ParentLoopRow *LoopRow

	Index int
	Key   string
	Title string

	TreeKeyParts []string
	TreeKey      string

	LoopRow []*LoopRow

	LoopSequential bool

	Retry int

	ExecTimeout *time.Duration

	RegisterVars map[string]*registry.VarDef

	Logger      *logrus.Entry
	Depth       int
	HasChildren bool

	State StateType

	Skip           bool
	NoSkipChildren bool
	NoSkip         bool

	GlobalRunCtx *GlobalRunCtx
	CfgPlay      *CfgPlay

	VarsClean bool
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

		ExecTimeout: cp.ExecTimeout,

		RegisterVars: registerVars,

		Depth:       cp.Depth,
		HasChildren: cp.HasChildren,

		GlobalRunCtx: cp.GlobalRunCtx,
		CfgPlay:      cp,
	}

	if cp.VarsClean != nil {
		p.VarsClean = *cp.VarsClean
	}

	if cp.Retry != nil {
		p.Retry = *cp.Retry
	}

	p.RunCtx = ctx
	p.ParentLoopRow = parentLoopRow

	p.TreeKeyParts = GetTreeKeyParts(p)
	p.TreeKey = strings.Join(p.TreeKeyParts, "|")
	logger := logrus.WithFields(logrus.Fields{
		"tree": p.TreeKey,
	})
	loggerCtx := context.WithValue(context.Background(), config.LogContextKey("indentation"), p.Depth+1)
	logger = logger.WithContext(loggerCtx)
	p.Logger = logger

	cfg := cp.App.GetConfig()
	if parentLoopRow != nil {
		p.NoSkipChildren = parentLoopRow.ParentPlay.NoSkipChildren
	}
	if len(cfg.PlayKey) != 0 && !p.NoSkipChildren {
		match := p.KeysMatch(cfg.PlayKey)
		p.handlePlayKey(match)
	}

	gRunCtx := p.GlobalRunCtx
	if cfg.PlayKeyStart != "" && gRunCtx.SkippingState == nil {
		b := true
		gRunCtx.SkippingState = &b
	}

	if cfg.PlayKeyStart != "" && p.KeyMatch(cfg.PlayKeyStart) {
		b := false
		gRunCtx.SkippingState = &b
	}
	if gRunCtx.SkippingState != nil && !p.NoSkipChildren {
		p.handlePlayKey(!*gRunCtx.SkippingState)
	}
	if cfg.PlayKeyEnd != "" && p.KeyMatch(cfg.PlayKeyEnd) {
		b := true
		gRunCtx.SkippingState = &b
	}

	if p.Skip {
		logger.Debug("  " + p.GetTitleMsg())
		logger.Debug("  skipping...")
		return nil
	}

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

		runCtx := CreateRunCtx()

		loop := &LoopRow{
			Name:          cfgLoopRow.Name,
			Key:           cfgLoopRow.Key,
			Index:         cfgLoopRow.Index,
			Vars:          cfgLoopRow.Vars,
			IsLoopRowItem: cfgLoopRow.IsLoopRowItem,
			ParentPlay:    p,
			RunCtx:        runCtx,
		}

		p.LoopRow[i] = loop

		if cp.CfgChk != nil {
			loop.HasChk = true
			loop.PreChk = CreateChk(cp.CfgChk, loop, true)
			loop.PostChk = CreateChk(cp.CfgChk, loop, false)
			p.NoSkip = true
		}

		switch pl := cp.CfgPlay.(type) {
		case []*CfgPlay:
			var sp []*Play
			for _, child := range pl {
				np := CreatePlay(child, runCtx, loop)
				if np != nil {
					sp = append(sp, np)
					if np.NoSkip {
						p.NoSkip = true
					}
				}
			}
			loop.Play = sp
		case *CfgCmd:
			if pl.CfgPlaySubstitution != nil {
				var sp []*Play
				np := CreatePlay(pl.CfgPlaySubstitution, runCtx, loop)
				if np != nil {
					sp = []*Play{np}
					if np.NoSkip {
						p.NoSkip = true
					}
				}
				loop.Play = sp
			} else {
				loop.Play = CreateCmd(pl, loop)
				p.NoSkip = true
			}
		}

	}

	return p
}

func (p *Play) LoadVars() {
	logger := p.Logger

	if !p.NoSkip {
		logger.Debug("  " + p.GetTitleMsg())
		logger.Debug("  skipping...")
		return
	}

	logger.Info("  " + p.GetTitleMsg())

	cp := p.CfgPlay

	parentCtx := p.RunCtx

	for _, loop := range p.LoopRow {
		ctx := loop.RunCtx
		vars := ctx.Vars
		varsDefault := ctx.VarsDefault

		if loop.IsLoopRowItem {
			logger.Info(strings.Repeat("  ", 2) + "⦿ " + loop.Name)
		}
		if !p.VarsClean {
			for k, v := range parentCtx.Vars.Items() {
				vars.Set(k, v)
			}
		}
		for _, v := range cp.Vars {
			v.RegisterValueTo(vars)
		}
		for _, v := range loop.Vars {
			v.RegisterValueTo(vars)
		}

		if !p.VarsClean {
			for k, v := range parentCtx.VarsDefault.Items() {
				varsDefault.Set(k, v)
			}
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

		if loop.HasChk {
			loop.PreChk.LoadVars()
			loop.PostChk.LoadVars()
		}

		switch pl := loop.Play.(type) {
		case []*Play:
			for _, p := range pl {
				p.LoadVars()
			}
		case *Cmd:
			pl.LoadVars()
		}
	}
}

func (p *Play) handlePlayKey(match bool) {
	app := p.App
	cfg := app.GetConfig()
	cp := p.CfgPlay
	if p.HasChildren {
		if match {
			p.NoSkipChildren = true
		}
	} else {
		if !match {
			skip := true
			buildCtx := cp.BuildCtx
			loadedSnippetKey := buildCtx.LoadedSnippetKey(cp.Scope, p.Key)

			for _, pkey := range cfg.PlayKey {
				if loadedSnippet, hasKey := buildCtx.LoadedSnippets[loadedSnippetKey]; hasKey {
					if !cfg.PlayKeyNoDeps {
						if b, ok := loadedSnippet.requiredByDependencies[pkey]; b && ok {
							skip = false
							break
						}
					}
					if !cfg.PlayKeyNoPost {
						if b, ok := loadedSnippet.requiredByPostInstall[pkey]; b && ok {
							skip = false
							break
						}
					}
				}
			}

			p.Skip = skip
		}
	}
}

func (p *Play) KeysMatch(keys []string) bool {
	return tools.SliceContainsString(keys, p.Key) ||
		tools.SliceContainsString(keys, p.TreeKey)
}
func (p *Play) KeyMatch(key string) bool {
	return p.Key == key || p.TreeKey == key
}

func (p *Play) GetTitleMsg() string {
	cp := p.CfgPlay
	var icon string
	if cp.ParentCfgPlay == nil {
		icon = `🠞`
	} else if !cp.HasChildren {
		icon = `⯈`
	} else {
		icon = `⤷`
	}
	return icon + " " + p.GetTitle()
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
		key := vr.GetFrom()
		value := varsRegistry.GetVarBySlice(kp, key)
		if value == "" && vr.Persist {
			for i := len(kp); i >= 0; i-- {
				dp2 := kp[0:i]
				value = varsRegistry.PersistGetVarBySlice(dp2, key)
				if value != "" {
					break
				}
			}
		}
		if value != "" {
			varsRegistry.SetVarBySlice(dp, vr.To, value)
			if vr.Persist {
				varsRegistry.PersistSetVarBySlice(dp, key, value)
			}
		}
	}
}

func (p *Play) Run() error {

	app := p.App
	if app.IsExiting() {
		return nil
	}

	if !p.NoSkip {
		return nil
	}

	var icon string
	if p.ParentLoopRow == nil {
		icon = `🠞`
	} else if !p.HasChildren {
		icon = `⯈`
	} else {
		icon = `⤷`
	}

	logger := p.Logger

	logger.Info(icon + " " + p.GetTitle())

	gRunCtx := p.GlobalRunCtx
	runReport := gRunCtx.RunReport

	var errSlice []error
	runLoopSeq := func(loop *LoopRow) error {

		if loop.IsLoopRowItem {
			logger.Info("  ⦿ " + loop.Name)
		}

		switch pl := loop.Play.(type) {
		case *Cmd:
			if err := pl.PreflightRun(); err != nil {
				return err
			}
			if loop.HasChk {
				for k, v := range pl.Vars {
					loop.PreChk.Vars[k] = v
					loop.PostChk.Vars[k] = v
				}
			}
		}

		if loop.HasChk {
			runReport.Total++
			if ok, _ := loop.PreChk.Run(); ok {
				return nil
			}
		} else {
			if !p.HasChildren {
				runReport.Total++
				runReport.OK++
			}
		}

		var localErrSlice []error

		localErrSlice = make([]error, 0)

		switch pl := loop.Play.(type) {
		case []*Play:
			for tries := p.Retry + 1; tries > 0; tries-- {

				for _, child := range pl {
					if err := child.Run(); err != nil {
						localErrSlice = append(localErrSlice, err)
						break
					}

					if app.IsExiting() {
						break
					}
				}

				if len(localErrSlice) == 0 && loop.HasChk {
					if ok, err := loop.PostChk.Run(); !ok {
						localErrSlice = append(localErrSlice, err)
					} else {
						break
					}
				}

				if app.IsExiting() {
					break
				}
			}
		case *Cmd:
			for tries := p.Retry + 1; tries > 0; tries-- {
				pl.RegisterVarsLoad()
				if err := pl.Run(); err != nil {
					localErrSlice = append(localErrSlice, err)
				}

				if len(localErrSlice) == 0 && loop.HasChk {
					if ok, err := loop.PostChk.Run(); !ok {
						localErrSlice = append(localErrSlice, err)
					} else {
						break
					}
				}

				if app.IsExiting() {
					break
				}
			}
		}

		if len(localErrSlice) > 0 {
			errSlice = append(errSlice, localErrSlice...)
			return multierr.Combine(localErrSlice...)
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
	logrus.Info("🚀 running playbook")
	return p.Run()
}

func unexpectedTypePlay(m map[string]interface{}, key string) {
	errors.UnexpectedType(m, key, "playbook")
}
