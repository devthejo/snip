package play

import (
	"github.com/sirupsen/logrus"
	"gitlab.com/youtopia.earth/ops/snip/errors"
	snipplugin "gitlab.com/youtopia.earth/ops/snip/plugin"
	"gitlab.com/youtopia.earth/ops/snip/plugin/loader"
	"gitlab.com/youtopia.earth/ops/snip/plugin/middleware"
	"gitlab.com/youtopia.earth/ops/snip/plugin/runner"
	"gitlab.com/youtopia.earth/ops/snip/variable"
)

type CfgCmd struct {
	CfgPlay *CfgPlay

	Command []string

	Loader      *loader.Loader
	Middlewares []*middleware.Middleware
	Runner      *runner.Runner

	Dir string

	RequiredFiles map[string]string

	Depth int
}

func CreateCfgCmd(cp *CfgPlay, c []string) *CfgCmd {
	ccmd := &CfgCmd{
		CfgPlay:       cp,
		Command:       c,
		Depth:         cp.Depth + 1,
		Dir:           cp.Dir,
		RequiredFiles: make(map[string]string),
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

func (ccmd *CfgCmd) GetLoaderVarsMap(mVar map[string]*variable.Var) map[string]string {
	middlewareVars := make(map[string]string)
	for _, v := range mVar {
		var val string
		if v.Default != "" {
			val = v.Default
		}
		if v.Value != "" {
			val = v.Value
		}
		middlewareVars[v.Name] = val
	}
	return middlewareVars
}

func (ccmd *CfgCmd) LoadLoader() {
	app := ccmd.CfgPlay.App
	cfg := app.GetConfig()

	appConfig := &snipplugin.AppConfig{
		DeploymentName: cfg.DeploymentName,
		BuildDir:       cfg.BuildDir,
		SnippetsDir:    cfg.SnippetsDir,
	}

	lr := ccmd.Loader

	loaderVars := ccmd.GetLoaderVarsMap(lr.Vars)

	loadCfg := &loader.Config{
		AppConfig:         appConfig,
		LoaderVars:        loaderVars,
		DefaultsPlayProps: make(map[string]interface{}),
		Command:           ccmd.Command,
		RequiredFiles:     ccmd.RequiredFiles,
	}

	lr.Plugin.Load(loadCfg)

	ccmd.Command = loadCfg.Command
	ccmd.RequiredFiles = loadCfg.RequiredFiles
	ccmd.CfgPlay.ParseMapAsDefault(loadCfg.DefaultsPlayProps)
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
			if loaderPlugin.Check(ccmd.Command) {
				ccmd.Loader = &loader.Loader{
					Name:   v,
					Plugin: loaderPlugin,
				}
				break
			}
		}
		if ccmd.Loader == nil {
			logrus.Fatalf("no loader match with %v at depth %v", ccmd.Command, ccmd.Depth)
		}
		return
	}

	for _, v := range *cp.Loaders {
		if v.Plugin.Check(ccmd.Command) {
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
