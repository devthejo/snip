package play

import (
	"strings"

	"github.com/sirupsen/logrus"
	"gitlab.com/youtopia.earth/ops/snip/errors"
	snipplugin "gitlab.com/youtopia.earth/ops/snip/plugin"
	"gitlab.com/youtopia.earth/ops/snip/plugin/loader"
	"gitlab.com/youtopia.earth/ops/snip/plugin/middleware"
	"gitlab.com/youtopia.earth/ops/snip/plugin/runner"
	"gitlab.com/youtopia.earth/ops/snip/registry"
	"gitlab.com/youtopia.earth/ops/snip/variable"
)

type CfgCmd struct {
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

func CreateCfgCmd(cp *CfgPlay, c []string) *CfgCmd {
	originalCommand := make([]string, len(c))
	copy(originalCommand, c)

	ccmd := &CfgCmd{
		CfgPlay:         cp,
		OriginalCommand: c,
		Command:         c,
		Depth:           cp.Depth + 1,
		Dir:             cp.Dir,
		RequiredFiles:   make(map[string]string),
	}
	ccmd.Parse()
	return ccmd
}

func (ccmd *CfgCmd) Parse() {
	ccmd.ParseLoader()
	ccmd.ParseMiddlewares()
	ccmd.ParseRunner()

	ccmd.LoadLoader()
}

func (ccmd *CfgCmd) HandleDependencies() {
	ccmd.RequireDependencies()
	ccmd.RegisterInDependencies()
}

func (ccmd *CfgCmd) RequireDependencies() {
	cp := ccmd.CfgPlay
	ls := cp.BuildCtx.LoadedSnippetsUpstream
	for _, dep := range cp.Dependencies {
		if b, ok := ls[dep]; ok && b {
			continue
		}
		k := ccmd.OriginalCommand[0]
		logrus.Debugf(`◘ dependency required by "%s" autoloading "%s"`, k, dep)

		m := make(map[string]interface{})
		m["play"] = dep

		parent := cp.ParentCfgPlay

		buildCtx := CreateNextBuildCtx(cp.BuildCtx)
		buildCtx.LoadedSnippetsDownstreamParents = nil

		playSlice := parent.CfgPlay.([]*CfgPlay)
		playSlice = append(playSlice, CreateCfgPlay(cp.App, m, parent, buildCtx))
		parent.CfgPlay = playSlice
	}
}

func (ccmd *CfgCmd) RequirePostInstall() {
	cp := ccmd.CfgPlay
	buildCtx := cp.BuildCtx
	ls := buildCtx.LoadedSnippetsDownstream

	for _, dep := range cp.PostInstall {

		if b, ok := ls[dep]; ok && b {
			return
		}

		k := ccmd.OriginalCommand[0]
		logrus.Debugf(`◘ post-install required by "%s" autoloading "%s"`, k, dep)

		m := make(map[string]interface{})
		m["play"] = dep

		parent := cp.ParentCfgPlay

		playSlice := parent.CfgPlay.([]*CfgPlay)
		playSlice = append(playSlice, CreateCfgPlay(cp.App, m, parent, cp.BuildCtx))
		parent.CfgPlay = playSlice

	}
}

func (ccmd *CfgCmd) RegisterInDependencies() {
	cp := ccmd.CfgPlay
	buildCtx := cp.BuildCtx
	k := ccmd.OriginalCommand[0]
	buildCtx.RegisterLoadedSnippet(k)
}

func (ccmd *CfgCmd) GetLoaderVarsMap(useVars []string, mVar map[string]*variable.Var) map[string]string {
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

func (ccmd *CfgCmd) GetLoaderConfig(lr *loader.Loader) *loader.Config {
	app := ccmd.CfgPlay.App
	cfg := app.GetConfig()

	appConfig := &snipplugin.AppConfig{
		DeploymentName: cfg.DeploymentName,
		SnippetsDir:    cfg.SnippetsDir,
	}

	loaderVars := ccmd.GetLoaderVarsMap(lr.Plugin.UseVars, lr.Vars)

	command := make([]string, len(ccmd.Command))
	copy(command, ccmd.Command)
	registerVars := make(map[string]*registry.VarDef)
	for k, v := range ccmd.CfgPlay.RegisterVars {
		registerVars[k] = v
	}
	loaderCfg := &loader.Config{
		AppConfig:         appConfig,
		LoaderVars:        loaderVars,
		DefaultsPlayProps: make(map[string]interface{}),
		Command:           command,
		RequiredFiles:     ccmd.RequiredFiles,
		RegisterVars:      registerVars,
	}

	return loaderCfg

}

func (ccmd *CfgCmd) LoadLoader() {

	lr := ccmd.Loader

	loaderCfg := ccmd.GetLoaderConfig(lr)

	lr.Plugin.Load(loaderCfg)

	command := make([]string, len(loaderCfg.Command))
	copy(command, loaderCfg.Command)
	ccmd.Command = command
	ccmd.RequiredFiles = loaderCfg.RequiredFiles
	ccmd.CfgPlay.ParseMapAsDefault(loaderCfg.DefaultsPlayProps)
}

func (ccmd *CfgCmd) ParseLoader() {
	cp := ccmd.CfgPlay
	app := cp.App

	if cp.ForceLoader {
		ccmd.Loader = (*cp.Loaders)[0]
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
			loaderCfg := ccmd.GetLoaderConfig(lr)
			if loaderPlugin.Check(loaderCfg) {
				ccmd.Loader = lr
				break
			}
		}
		if ccmd.Loader == nil {
			logrus.Fatalf("no loader match with %v at depth %v", ccmd.Command, ccmd.Depth)
		}
		return
	}

	for _, v := range *cp.Loaders {
		loaderCfg := ccmd.GetLoaderConfig(v)
		if v.Plugin.Check(loaderCfg) {
			ccmd.Loader = v
			break
		}
	}

	if ccmd.Loader == nil {
		logrus.Fatalf("no loader match with %v at depth %v", ccmd.Command, ccmd.Depth)
		return
	}

}

func (ccmd *CfgCmd) ParseMiddlewares() {
	cp := ccmd.CfgPlay
	if cp.Middlewares == nil {
		return
	}
	for _, v := range *cp.Middlewares {
		ccmd.Middlewares = append(ccmd.Middlewares, v)
	}
}

func (ccmd *CfgCmd) ParseRunner() {
	cp := ccmd.CfgPlay
	var rr string
	if cp.Runner != nil {
		ccmd.Runner = cp.Runner
	} else {
		rr = cp.App.GetConfig().Runner
		ccmd.Runner = &runner.Runner{
			Name: rr,
		}
	}
}

func unexpectedTypeCmd(m map[string]interface{}, key string) {
	errors.UnexpectedType(m, key, "cmd")
}
