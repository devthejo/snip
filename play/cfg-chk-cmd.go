package play

import (
	"strings"

	snipplugin "gitlab.com/youtopia.earth/ops/snip/plugin"
	"gitlab.com/youtopia.earth/ops/snip/plugin/loader"
	"gitlab.com/youtopia.earth/ops/snip/plugin/middleware"
	"gitlab.com/youtopia.earth/ops/snip/plugin/runner"
	"gitlab.com/youtopia.earth/ops/snip/registry"
	"gitlab.com/youtopia.earth/ops/snip/variable"
)

type CfgChkCmd struct {
	CfgPlay *CfgPlay

	OriginalCommand []string
	Command         []string

	Loader      *loader.Loader
	Middlewares []*middleware.Middleware
	Runner      *runner.Runner

	Dir string

	RequiredFiles map[string]string

	Depth int
}

func CreateCfgChkCmd(cp *CfgPlay, c []string) *CfgChkCmd {
	originalCommand := make([]string, len(c))
	copy(originalCommand, c)

	chk := &CfgChkCmd{
		CfgPlay:         cp,
		OriginalCommand: c,
		Command:         c,
		Depth:           cp.Depth + 1,
		Dir:             cp.Dir,
		RequiredFiles:   make(map[string]string),
	}
	chk.Parse()
	return chk
}

func (chk *CfgChkCmd) Parse() {
	chk.ParseLoader()
	chk.ParseMiddlewares()
	chk.ParseRunner()

	chk.LoadLoader()
}

func (chk *CfgChkCmd) GetLoaderVarsMap(useVars []string, mVar map[string]*variable.Var) map[string]string {
	pVars := make(map[string]string)
	for _, useV := range useVars {
		key := strings.ToUpper(useV)
		v := mVar[key]
		var val string
		if v != nil && v.Default != "" {
			val = v.Default
		}
		if v != nil && v.Value != "" {
			val = v.Value
		}
		pVars[strings.ToLower(key)] = val
	}
	return pVars
}

func (chk *CfgChkCmd) GetLoaderConfig(lr *loader.Loader) *loader.Config {
	app := chk.CfgPlay.App
	cfg := app.GetConfig()

	appConfig := &snipplugin.AppConfig{
		DeploymentName: cfg.DeploymentName,
		SnippetsDir:    cfg.SnippetsDir,
	}

	loaderVars := chk.GetLoaderVarsMap(lr.Plugin.UseVars, lr.Vars)

	command := make([]string, len(chk.Command))
	copy(command, chk.Command)
	registerVars := make(map[string]*registry.VarDef)
	for k, v := range chk.CfgPlay.RegisterVars {
		registerVars[k] = v
	}
	loaderCfg := &loader.Config{
		AppConfig:         appConfig,
		LoaderVars:        loaderVars,
		DefaultsPlayProps: make(map[string]interface{}),
		Command:           command,
		RequiredFiles:     chk.RequiredFiles,
		RegisterVars:      registerVars,
	}

	return loaderCfg

}

func (chk *CfgChkCmd) LoadLoader() {

	lr := chk.Loader

	if lr == nil {
		return
	}

	loaderCfg := chk.GetLoaderConfig(lr)

	lr.Plugin.Load(loaderCfg)

	command := make([]string, len(loaderCfg.Command))
	copy(command, loaderCfg.Command)
	chk.Command = command
	chk.RequiredFiles = loaderCfg.RequiredFiles
	chk.CfgPlay.ParseMapAsDefault(loaderCfg.DefaultsPlayProps)
}

func (chk *CfgChkCmd) ParseLoader() {
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

func (chk *CfgChkCmd) ParseMiddlewares() {
	cp := chk.CfgPlay
	if cp.Middlewares == nil {
		return
	}
	for _, v := range *cp.Middlewares {
		chk.Middlewares = append(chk.Middlewares, v)
	}
}

func (chk *CfgChkCmd) ParseRunner() {
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
