package play

import (
	"strings"

	snipplugin "gitlab.com/ytopia/ops/snip/plugin"
	"gitlab.com/ytopia/ops/snip/plugin/loader"
	"gitlab.com/ytopia/ops/snip/plugin/middleware"
	"gitlab.com/ytopia/ops/snip/plugin/processor"
	"gitlab.com/ytopia/ops/snip/plugin/runner"
	"gitlab.com/ytopia/ops/snip/variable"
)

type CfgChk struct {
	CfgPlay *CfgPlay

	OriginalCommand []string
	Command         []string

	Loader      *loader.Loader
	Middlewares []*middleware.Middleware
	Runner      *runner.Runner

	Dir string

	RequiredFiles              map[string]string
	RequiredFilesSrcProcessors map[string][]func(*processor.Config, *string) error

	Depth int
}

func CreateCfgChk(cp *CfgPlay, c []string) *CfgChk {
	originalCommand := make([]string, len(c))
	copy(originalCommand, c)

	chk := &CfgChk{
		CfgPlay:                    cp,
		OriginalCommand:            c,
		Command:                    c,
		Depth:                      cp.Depth + 1,
		Dir:                        cp.Dir,
		RequiredFiles:              make(map[string]string),
		RequiredFilesSrcProcessors: make(map[string][]func(*processor.Config, *string) error),
	}

	chk.Parse()
	return chk
}

func (chk *CfgChk) Parse() {
	chk.ParseLoader()
	chk.ParseMiddlewares()
	chk.ParseRunner()

	chk.LoadLoader()
}

func (chk *CfgChk) GetLoaderVarsMap(useVars []string, mVar map[string]*variable.Var) map[string]string {
	pVars := make(map[string]string)
	for _, useV := range useVars {
		key := strings.ToUpper(useV)
		v := mVar[key]
		var val string
		if v != nil && v.GetDefault() != "" {
			val = v.GetDefault()
		}
		if v != nil && v.GetValue() != "" {
			val = v.GetValue()
		}
		pVars[strings.ToLower(key)] = val
	}
	return pVars
}

func (chk *CfgChk) GetLoaderConfig(lr *loader.Loader) *loader.Config {
	app := chk.CfgPlay.App
	cfg := app.GetConfig()

	appConfig := &snipplugin.AppConfig{
		DeploymentName: cfg.DeploymentName,
		SnippetsDir:    cfg.SnippetsDir,
	}

	loaderVars := chk.GetLoaderVarsMap(lr.Plugin.UseVars, lr.Vars)

	command := make([]string, len(chk.Command))
	copy(command, chk.Command)
	loaderCfg := &loader.Config{
		AppConfig:                  appConfig,
		LoaderVars:                 loaderVars,
		DefaultsPlayProps:          make(map[string]interface{}),
		Command:                    command,
		RequiredFiles:              chk.RequiredFiles,
		RequiredFilesSrcProcessors: chk.RequiredFilesSrcProcessors,
		ParentBuildFile:            chk.CfgPlay.ParentBuildFile,
	}

	return loaderCfg

}

func (chk *CfgChk) LoadLoader() {

	lr := chk.Loader

	if lr == nil {
		return
	}

	loaderCfg := chk.GetLoaderConfig(lr)
	lr.Plugin.Load(loaderCfg)

	chk.CfgPlay.ParseMapAsDefault(loaderCfg.DefaultsPlayProps)

	loaderCfg = chk.GetLoaderConfig(lr)
	if lr.Plugin.PostLoad != nil {
		lr.Plugin.PostLoad(loaderCfg)
	}

	command := make([]string, len(loaderCfg.Command))
	copy(command, loaderCfg.Command)
	chk.Command = command
	chk.RequiredFiles = loaderCfg.RequiredFiles
	chk.RequiredFilesSrcProcessors = loaderCfg.RequiredFilesSrcProcessors

	// re-inject props from cfg-play after ParseMapAsDefault
	chk.ParseMiddlewares()
	chk.ParseRunner()
}

func (chk *CfgChk) ParseLoader() {
	cp := chk.CfgPlay
	app := cp.App

	if cp.ForceLoader {
		chk.Loader = (*cp.Loaders)[0]
		return
	}

	if cp.Loaders == nil {
		cfg := app.GetConfig()
		for _, v := range cfg.Loaders {
			loaderPlugin := app.GetLoader(v)
			lr := &loader.Loader{
				Name:   v,
				Plugin: loaderPlugin,
			}
			loaderCfg := chk.GetLoaderConfig(lr)
			if loaderPlugin.Check(loaderCfg) {
				chk.Loader = lr
				break
			}
		}
		return
	}

	for _, v := range *cp.Loaders {
		loaderCfg := chk.GetLoaderConfig(v)
		if v.Plugin.Check(loaderCfg) {
			chk.Loader = v
			break
		}
	}

}

func (chk *CfgChk) ParseMiddlewares() {
	cp := chk.CfgPlay
	if cp.Middlewares == nil {
		return
	}
	for _, v := range *cp.Middlewares {
		chk.Middlewares = append(chk.Middlewares, v)
	}
}

func (chk *CfgChk) ParseRunner() {
	cp := chk.CfgPlay
	var rr string
	if cp.Runner != nil {
		chk.Runner = cp.Runner
	} else {
		rr = cp.App.GetConfig().Runner
		chk.Runner = &runner.Runner{
			Name: rr,
		}
	}
}
