package play

import (
	"github.com/sirupsen/logrus"
	"gitlab.com/youtopia.earth/ops/snip/decode"
	"gitlab.com/youtopia.earth/ops/snip/errors"
	snipplugin "gitlab.com/youtopia.earth/ops/snip/plugin"
	"gitlab.com/youtopia.earth/ops/snip/plugin/loader"
	"gitlab.com/youtopia.earth/ops/snip/plugin/middleware"
	"gitlab.com/youtopia.earth/ops/snip/plugin/runner"
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

func (ccmd *CfgCmd) LoadLoader() {
	app := ccmd.CfgPlay.App
	cfg := app.GetConfig()

	loadCfg := &loader.Config{
		AppConfig: &snipplugin.AppConfig{
			DeploymentName: cfg.DeploymentName,
			BuildDir:       cfg.BuildDir,
			SnippetsDir:    cfg.SnippetsDir,
		},
		DefaultsPlayProps: make(map[string]interface{}),
		Command:           ccmd.Command,
		RequiredFiles:     ccmd.RequiredFiles,
	}

	ccmd.Loader.Load(loadCfg)

	ccmd.Command = loadCfg.Command
	ccmd.RequiredFiles = loadCfg.RequiredFiles
	ccmd.CfgPlay.ParseMapAsDefault(loadCfg.DefaultsPlayProps)
}

func (ccmd *CfgCmd) ParseLoader() {
	cp := ccmd.CfgPlay
	app := cp.App
	switch v := cp.Loader.(type) {
	case string:
		ccmd.Loader = app.GetLoader(v)
	case []string:
		s, err := decode.ToStrings(v)
		errors.Check(err)
		for _, v := range s {
			loader := app.GetLoader(v)
			if loader.Check(ccmd.Command) {
				ccmd.Loader = loader
				break
			}
		}
		if ccmd.Loader == nil {
			logrus.Fatalf("no loader match with %v at depth %v", ccmd.Command, ccmd.Depth)
		}
	case nil:
		cfg := app.GetConfig()
		for _, v := range cfg.Loaders {
			loader := app.GetLoader(v)
			if loader.Check(ccmd.Command) {
				ccmd.Loader = loader
				break
			}
		}
		if ccmd.Loader == nil {
			logrus.Fatalf("no loader match with %v at depth %v", ccmd.Command, ccmd.Depth)
		}
	}
}

func (ccmd *CfgCmd) ParseMiddlewares() {
	cp := ccmd.CfgPlay
	app := cp.App
	if cp.Middlewares == nil {
		return
	}
	for _, v := range *cp.Middlewares {
		middleware := app.GetMiddleware(v.Name)
		ccmd.Middlewares = append(ccmd.Middlewares, middleware)
	}
}

func (ccmd *CfgCmd) ParseRunner() {
	cp := ccmd.CfgPlay
	app := cp.App
	var runner string
	if cp.Runner != "" {
		runner = cp.Runner
	} else {
		runner = cp.App.GetConfig().Runner
	}
	ccmd.Runner = app.GetRunner(runner)
}

func unexpectedTypeCmd(m map[string]interface{}, key string) {
	errors.UnexpectedType(m, key, "cmd")
}
