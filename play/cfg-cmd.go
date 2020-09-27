package play

import (
	"strings"

	"github.com/sirupsen/logrus"
	"gitlab.com/ytopia/ops/snip/errors"
	snipplugin "gitlab.com/ytopia/ops/snip/plugin"
	"gitlab.com/ytopia/ops/snip/plugin/loader"
	"gitlab.com/ytopia/ops/snip/plugin/middleware"
	"gitlab.com/ytopia/ops/snip/plugin/processor"
	"gitlab.com/ytopia/ops/snip/plugin/runner"
	"gitlab.com/ytopia/ops/snip/registry"
	"gitlab.com/ytopia/ops/snip/variable"
)

type CfgCmd struct {
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

	CfgPlaySubstitutionMap map[string]interface{}
	CfgPlaySubstitution    *CfgPlay
}

func CreateCfgCmd(cp *CfgPlay, c []string) *CfgCmd {
	originalCommand := make([]string, len(c))
	copy(originalCommand, c)
	ccmd := &CfgCmd{
		CfgPlay:                    cp,
		OriginalCommand:            c,
		Command:                    c,
		Depth:                      cp.Depth + 1,
		Dir:                        cp.Dir,
		RequiredFiles:              make(map[string]string),
		RequiredFilesSrcProcessors: make(map[string][]func(*processor.Config, *string) error),
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

func (ccmd *CfgCmd) RequireDependencies() {
	cp := ccmd.CfgPlay
	buildCtx := cp.BuildCtx
	ls := buildCtx.LoadedSnippetsUpstream
	for _, dep := range cp.Dependencies {
		loadedDepKey := buildCtx.LoadedSnippetKey(cp.Scope, dep)
		if _, ok := ls[loadedDepKey]; ok {
			continue
		}
		k := ccmd.OriginalCommand[0]
		logrus.Debugf(`◘ dependency required by "%s" autoloading "%s"`, k, dep)

		m := make(map[string]interface{})
		m["play"] = dep

		parent := cp.ParentCfgPlay

		nextBuildCtx := CreateNextBuildCtx(buildCtx)
		nextBuildCtx.LoadedSnippetsDownstreamParents = nil

		// default runner from dependency caller
		nextBuildCtx.DefaultRunner = ccmd.Runner
		nextBuildCtx.DefaultVars = cp.Vars
		depPlay := CreateCfgPlay(cp.App, m, parent, nextBuildCtx)

		playSlice := parent.CfgPlay.([]*CfgPlay)
		playSlice = append(playSlice, depPlay)
		parent.CfgPlay = playSlice
	}
}

func (ccmd *CfgCmd) RequirePostInstall() {
	cp := ccmd.CfgPlay
	buildCtx := cp.BuildCtx
	ls := buildCtx.LoadedSnippetsDownstream
	scope := cp.Scope

	for _, dep := range cp.PostInstall {
		loadedDepKey := buildCtx.LoadedSnippetKey(scope, dep)
		if _, ok := ls[loadedDepKey]; ok {
			return
		}

		k := ccmd.OriginalCommand[0]
		logrus.Debugf(`◘ post-install required by "%s" autoloading "%s"`, k, dep)

		m := make(map[string]interface{})
		m["play"] = dep

		parent := cp.ParentCfgPlay

		playSlice := parent.CfgPlay.([]*CfgPlay)
		playSlice = append(playSlice, CreateCfgPlay(cp.App, m, parent, buildCtx))
		parent.CfgPlay = playSlice
	}
}

func (ccmd *CfgCmd) registerRequiredByDependencies() {
	cp := ccmd.CfgPlay
	buildCtx := cp.BuildCtx
	ls := buildCtx.LoadedSnippets
	scope := cp.Scope
	by := ccmd.OriginalCommand[0]

	for _, dep := range cp.Dependencies {
		loadedDepKey := buildCtx.LoadedSnippetKey(scope, dep)
		if _, hasKey := ls[loadedDepKey]; !hasKey {
			ls[loadedDepKey] = CreateLoadedSnippet()
		}
		loadedDep := ls[loadedDepKey]
		loadedDep.requiredByDependencies[by] = true
	}
}
func (ccmd *CfgCmd) registerRequiredByPostInstall() {
	cp := ccmd.CfgPlay
	buildCtx := cp.BuildCtx
	ls := buildCtx.LoadedSnippets
	scope := cp.Scope
	by := ccmd.OriginalCommand[0]

	for _, dep := range cp.PostInstall {
		loadedDepKey := buildCtx.LoadedSnippetKey(scope, dep)
		if _, hasKey := ls[loadedDepKey]; !hasKey {
			ls[loadedDepKey] = CreateLoadedSnippet()
		}
		loadedDep := ls[loadedDepKey]
		loadedDep.requiredByDependencies[by] = true
	}
}

func (ccmd *CfgCmd) RegisterInDependencies() {
	cp := ccmd.CfgPlay
	buildCtx := cp.BuildCtx
	k := ccmd.OriginalCommand[0]
	buildCtx.RegisterLoadedSnippet(cp.Scope, k)
}

func (ccmd *CfgCmd) LoadCfgPlaySubstitution() {
	if ccmd.CfgPlaySubstitutionMap == nil {
		return
	}

	cp := ccmd.CfgPlay
	buildCtx := cp.BuildCtx

	m := ccmd.CfgPlaySubstitutionMap

	ccmd.CfgPlaySubstitution = CreateCfgPlay(cp.App, m, cp, buildCtx)
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

func (ccmd *CfgCmd) GetLoaderConfig(lr *loader.Loader, defaultCfg *loader.Config) *loader.Config {
	app := ccmd.CfgPlay.App
	cfg := app.GetConfig()

	loaderVars := ccmd.GetLoaderVarsMap(lr.Plugin.UseVars, lr.Vars)

	command := make([]string, len(ccmd.Command))
	copy(command, ccmd.Command)
	registerVars := make(map[string]*registry.VarDef)
	for k, v := range ccmd.CfgPlay.RegisterVars {
		registerVars[k] = v
	}

	var loaderCfg *loader.Config
	if defaultCfg == nil {

		loaderCfg = &loader.Config{}
		loaderCfg.DefaultsPlayProps = make(map[string]interface{})
		loaderCfg.AppConfig = &snipplugin.AppConfig{
			DeploymentName: cfg.DeploymentName,
			SnippetsDir:    cfg.SnippetsDir,
		}
		loaderCfg.ParentBuildFile = ccmd.CfgPlay.ParentBuildFile

	} else {
		loaderCfgCopy := *defaultCfg
		loaderCfg = &loaderCfgCopy
	}

	loaderCfg.LoaderVars = loaderVars
	loaderCfg.Command = command
	loaderCfg.RequiredFiles = ccmd.RequiredFiles
	loaderCfg.RequiredFilesSrcProcessors = ccmd.RequiredFilesSrcProcessors
	loaderCfg.RegisterVars = registerVars

	return loaderCfg

}

func (ccmd *CfgCmd) LoadLoader() {

	lr := ccmd.Loader

	if lr == nil {
		return
	}

	loaderCfg := ccmd.GetLoaderConfig(lr, nil)
	lr.Plugin.Load(loaderCfg)

	ccmd.CfgPlay.ParseMapAsDefault(loaderCfg.DefaultsPlayProps)

	loaderCfg = ccmd.GetLoaderConfig(lr, loaderCfg)
	if lr.Plugin.PostLoad != nil {
		lr.Plugin.PostLoad(loaderCfg)
	}

	command := make([]string, len(loaderCfg.Command))
	copy(command, loaderCfg.Command)
	ccmd.Command = command
	ccmd.RequiredFiles = loaderCfg.RequiredFiles
	ccmd.RequiredFilesSrcProcessors = loaderCfg.RequiredFilesSrcProcessors
	ccmd.CfgPlaySubstitutionMap = loaderCfg.CfgPlaySubstitutionMap
	ccmd.CfgPlay.BuildFile = loaderCfg.BuildFile

	// re-inject props from cfg-play after ParseMapAsDefault
	ccmd.ParseMiddlewares()
	ccmd.ParseRunner()
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
			loaderCfg := ccmd.GetLoaderConfig(lr, nil)
			if loaderPlugin.Check(loaderCfg) {
				ccmd.Loader = lr
				break
			}
		}
		// if ccmd.Loader == nil {
		// 	logrus.Fatalf("no loader match with %v at depth %v", ccmd.Command, ccmd.Depth)
		// }
		return
	}

	for _, v := range *cp.Loaders {
		loaderCfg := ccmd.GetLoaderConfig(v, nil)
		if v.Plugin.Check(loaderCfg) {
			ccmd.Loader = v
			break
		}
	}

	// if ccmd.Loader == nil {
	// 	logrus.Fatalf("no loader match with %v at depth %v", ccmd.Command, ccmd.Depth)
	// 	return
	// }

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
