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
	snipplugin "gitlab.com/ytopia/ops/snip/plugin"
	"gitlab.com/ytopia/ops/snip/plugin/runner"
	"gitlab.com/ytopia/ops/snip/registry"
	"gitlab.com/ytopia/ops/snip/tools"
	"gitlab.com/ytopia/ops/snip/variable"
)

type Play struct {
	App       App
	AppConfig *snipplugin.AppConfig

	RunVars *RunVars

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

	NoSkip         bool
	NoSkipChildren bool

	GlobalRunCtx *GlobalRunCtx
	CfgPlay      *CfgPlay

	VarsClean bool

	Use        map[string]string
	Persist    map[string]string
	Runner     *runner.Runner
	Dir        string
	Tmpdir     bool
	TmpdirName string
}

func CreatePlay(cp *CfgPlay, ctx *RunVars, parentLoopRow *LoopRow) *Play {
	var loopSequential bool
	if cp.LoopSequential != nil {
		loopSequential = *cp.LoopSequential
	}

	registerVars := make(map[string]*registry.VarDef)
	for k, v := range cp.RegisterVars {
		registerVars[k] = v
	}

	app := cp.App
	cfg := app.GetConfig()

	p := &Play{
		App: cp.App,
		AppConfig: &snipplugin.AppConfig{
			DeploymentName: cfg.DeploymentName,
			SnippetsDir:    cfg.SnippetsDir,
			Runner:         cfg.Runner,
		},
		Dir: cp.Dir,

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

		Use:     cp.Use,
		Persist: cp.Persist,
		Runner:  cp.Runner,
	}

	if cp.VarsClean != nil {
		p.VarsClean = *cp.VarsClean
	}

	if cp.Retry != nil {
		p.Retry = *cp.Retry
	}

	if cp.Tmpdir == nil {
		p.Tmpdir = true
	} else {
		p.Tmpdir = (*cp.Tmpdir)
	}
	if p.Tmpdir {
		tempKey, _ := tools.GenerateRandomString(16)
		p.TmpdirName = "snip-" + tempKey
	}

	p.RunVars = ctx
	p.ParentLoopRow = parentLoopRow

	p.TreeKeyParts = GetTreeKeyParts(p)
	p.TreeKey = strings.Join(p.TreeKeyParts, "|")
	logger := logrus.WithFields(logrus.Fields{
		"tree": p.TreeKey,
	})
	loggerCtx := context.WithValue(context.Background(), config.LogContextKey("indentation"), p.Depth+1)
	logger = logger.WithContext(loggerCtx)
	p.Logger = logger

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

		runCtx := p.RunVars.NewChild()

		loop := &LoopRow{
			Name:          cfgLoopRow.Name,
			Key:           cfgLoopRow.Key,
			Index:         cfgLoopRow.Index,
			Vars:          cfgLoopRow.Vars,
			IsLoopRowItem: cfgLoopRow.IsLoopRowItem,
			ParentPlay:    p,
			RunVars:       runCtx,
		}

		p.LoopRow[i] = loop

		if cp.CfgPreChk != nil {
			loop.PreChk = CreateChk(cp.CfgPreChk, loop, true)
		}
		if cp.CfgPostChk != nil {
			loop.PostChk = CreateChk(cp.CfgPostChk, loop, false)
		}

		switch pl := cp.CfgPlay.(type) {
		case []*CfgPlay:
			var sp []*Play
			for _, child := range pl {
				np := CreatePlay(child, runCtx, loop)
				if np != nil {
					sp = append(sp, np)
				}
			}
			loop.Play = sp
		case *CfgCmd:
			if pl.CfgPlaySubstitution != nil {
				var sp []*Play
				np := CreatePlay(pl.CfgPlaySubstitution, runCtx, loop)
				if np != nil {
					sp = []*Play{np}
				}
				loop.Play = sp
			} else {
				loop.Play = CreateCmd(pl, loop)
			}
		}

	}

	return p
}

func (p *Play) LoadSkip() {

	cp := p.CfgPlay

	cfg := cp.App.GetConfig()
	if p.ParentLoopRow != nil && p.ParentLoopRow.ParentPlay.NoSkipChildren {
		p.NoSkipChildren = true
		p.NoSkip = true
	}

	if len(cfg.PlayKey) != 0 {
		if p.KeysMatchSlice(cfg.PlayKey) {
			p.NoSkipChildren = true
		}
		match := p.KeysMatch(p.GlobalRunCtx.NoSkipTreeKeys)
		p.handlePlayKey(match)
	} else {
		p.NoSkip = true
	}

	gRunVars := p.GlobalRunCtx
	if cfg.PlayKeyStart != "" && gRunVars.SkippingState == nil {
		b := true
		gRunVars.SkippingState = &b
	}

	if cfg.PlayKeyStart != "" && p.KeyMatch(cfg.PlayKeyStart) {
		b := false
		gRunVars.SkippingState = &b
	}
	if gRunVars.SkippingState != nil {
		p.handlePlayKey(!*gRunVars.SkippingState)
	}
	if cfg.PlayKeyEnd != "" && p.KeyMatch(cfg.PlayKeyEnd) {
		b := true
		gRunVars.SkippingState = &b
	}

	for _, loop := range p.LoopRow {
		switch pl := loop.Play.(type) {
		case *Cmd:
			if p.NoSkip {
				p.GlobalRunCtx.NoSkipTreeKeys[p.Key] = true
				p.GlobalRunCtx.NoSkipTreeKeys[p.TreeKey] = true
			}
		case []*Play:
			if cp.CfgPreChk != nil || cp.CfgPostChk != nil {
				p.NoSkip = true
			}
			for _, child := range pl {
				child.LoadSkip()
				if child.NoSkip {
					p.NoSkip = true
				}
			}
		}
	}

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

	parentCtx := p.RunVars

	for _, loop := range p.LoopRow {
		ctx := loop.RunVars
		values := ctx.Values
		defaults := ctx.Defaults

		if loop.IsLoopRowItem {
			logger.Info(strings.Repeat("  ", 2) + "⦿ " + loop.Name)
		}
		if !p.VarsClean {
			for k, v := range parentCtx.Values.Items() {
				values.Set(k, v)
			}
		}
		for _, v := range cp.Vars {
			v.RegisterValueTo(values)
		}
		for _, v := range loop.Vars {
			v.RegisterValueTo(values)
		}

		if !p.VarsClean {
			for k, v := range parentCtx.Defaults.Items() {
				defaults.Set(k, v)
			}
		}
		for _, v := range loop.Vars {
			v.RegisterDefaultTo(defaults)
			v.HandleRequired(defaults, values)
		}
		for _, v := range cp.Vars {
			v.RegisterDefaultTo(defaults)
			v.HandleRequired(defaults, values)
		}

		cp.PromptPluginVars()

		switch pl := loop.Play.(type) {
		case []*Play:
			for _, p := range pl {
				p.LoadVars()
			}
		}
	}
}

func (p *Play) handlePlayKey(match bool) {
	app := p.App
	cfg := app.GetConfig()
	cp := p.CfgPlay
	if match {
		p.NoSkip = true
	} else {
		buildCtx := cp.BuildCtx
		loadedSnippetKey := buildCtx.LoadedSnippetKey(cp.Scope, p.Key)
		for pkey := range p.GlobalRunCtx.NoSkipTreeKeys {
			if loadedSnippet, hasKey := buildCtx.LoadedSnippets[loadedSnippetKey]; hasKey {
				if cfg.PlayKeyDeps {
					if b, ok := loadedSnippet.requiredByDependencies[pkey]; b && ok {
						p.NoSkip = true
						break
					}
				}
				if cfg.PlayKeyPost {
					if b, ok := loadedSnippet.requiredByPostInstall[pkey]; b && ok {
						p.NoSkip = true
						break
					}
				}
			}
		}

	}
}

func (p *Play) KeysMatchSlice(keys []string) bool {
	return tools.SliceContainsString(keys, p.Key) ||
		tools.SliceContainsString(keys, p.TreeKey)
}
func (p *Play) KeysMatch(keys map[string]bool) bool {
	if b, ok := keys[p.Key]; b && ok {
		return true
	}
	if b, ok := keys[p.TreeKey]; b && ok {
		return true
	}
	return false
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

func (p *Play) UpUseBeforeRun() error {
	r := p.Runner

	if r.Plugin == nil {
		r.Plugin = p.App.GetRunner(r.Name)
	}
	runnerVars := p.ParentLoopRow.RunVars.GetPluginVars("runner", r.Name, r.Plugin.UseVars, r.Vars)
	runCfg := &runner.Config{
		AppConfig:    p.AppConfig,
		RunnerVars:   runnerVars,
		Logger:       p.Logger,
		Cache:        p.App.GetCache(),
		VarsRegistry: p.App.GetVarsRegistry(),
		TreeKeyParts: p.TreeKeyParts,
		Dir:          p.Dir,
		// Quiet:         false,
		Use:        p.Use,
		TmpdirName: p.TmpdirName,
	}

	return r.Plugin.UpUse(runCfg)
}

func (p *Play) DownPersistAfterRun() error {

	r := p.Runner

	if r.Plugin == nil {
		r.Plugin = p.App.GetRunner(r.Name)
	}
	runnerVars := p.ParentLoopRow.RunVars.GetPluginVars("runner", r.Name, r.Plugin.UseVars, r.Vars)
	runCfg := &runner.Config{
		AppConfig:    p.AppConfig,
		RunnerVars:   runnerVars,
		Logger:       p.Logger,
		Cache:        p.App.GetCache(),
		VarsRegistry: p.App.GetVarsRegistry(),
		TreeKeyParts: p.TreeKeyParts,
		Dir:          p.Dir,
		// Quiet:         false,
		Persist:    p.Persist,
		TmpdirName: p.TmpdirName,
	}

	return r.Plugin.DownPersist(runCfg)
}

func (p *Play) Run() error {

	app := p.App
	if app.IsExiting() {
		return nil
	}

	if !p.NoSkip {
		return nil
	}

	p.GlobalRunCtx.CurrentTreeKey = p.TreeKey

	logger := p.Logger
	logger.Info(p.GetTitleMsg())

	gRunVars := p.GlobalRunCtx
	runReport := gRunVars.RunReport

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
		}

		if loop.PreChk != nil {
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

			upOK := true
			for _, child := range pl {
				if err := child.UpUseBeforeRun(); err != nil {
					localErrSlice = append(localErrSlice, err)
					upOK = false
					break
				}
				if app.IsExiting() {
					upOK = false
					break
				}
			}

			var runOK bool
			if upOK {
				runOK = true
				for tries := p.Retry + 1; tries > 0; tries-- {

					for _, child := range pl {
						if err := child.Run(); err != nil {
							localErrSlice = append(localErrSlice, err)
							runOK = false
							break
						}

						if app.IsExiting() {
							runOK = false
							break
						}
					}

					if len(localErrSlice) == 0 && loop.PostChk != nil {
						if ok, err := loop.PostChk.Run(); !ok {
							runOK = false
							localErrSlice = append(localErrSlice, err)
						} else {
							break
						}
					}

					if app.IsExiting() {
						runOK = false
						break
					}
				}
			}

			if runOK {
				for _, child := range pl {
					if err := child.DownPersistAfterRun(); err != nil {
						localErrSlice = append(localErrSlice, err)
						upOK = false
						break
					}
					if app.IsExiting() {
						upOK = false
						break
					}
				}
			}

		case *Cmd:
			for tries := p.Retry + 1; tries > 0; tries-- {
				if err := pl.Run(); err != nil {
					localErrSlice = append(localErrSlice, err)
				}

				if len(localErrSlice) == 0 && loop.PostChk != nil {
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
