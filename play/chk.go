package play

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	expect "gitlab.com/youtopia.earth/ops/snip/goexpect"
	"gitlab.com/youtopia.earth/ops/snip/registry"
	"gitlab.com/youtopia.earth/ops/snip/variable"

	"gitlab.com/youtopia.earth/ops/snip/config"
	snipplugin "gitlab.com/youtopia.earth/ops/snip/plugin"
	"gitlab.com/youtopia.earth/ops/snip/plugin/middleware"
	"gitlab.com/youtopia.earth/ops/snip/plugin/runner"
	"gitlab.com/youtopia.earth/ops/snip/proc"
	"gitlab.com/youtopia.earth/ops/snip/tools"
)

type Chk struct {
	App       App
	AppConfig *snipplugin.AppConfig

	Thread *proc.Thread

	ParentLoopRow *LoopRow

	IsPreRun bool

	Command []string
	Vars    map[string]string

	ExecTimeout *time.Duration

	Logger *logrus.Entry
	Depth  int

	Middlewares []*middleware.Middleware
	Runner      *runner.Runner

	RequiredFiles map[string]string

	Expect []expect.Batcher

	Dir string

	Closer *func(interface{}) bool

	RegisterVars map[string]*registry.VarDef
	Quiet        bool

	TreeKeyParts []string
	TreeKey      string
}

func (chk *Chk) EnvMap() map[string]string {
	m := make(map[string]string)
	for k, v := range chk.Vars {
		if k[0:1] != "@" {
			m[k] = v
		}
	}
	return m
}

func CreateChk(cchk *CfgChkCmd, ctx *RunCtx, parentLoopRow *LoopRow, isPreRun bool) *Chk {
	parentPlay := parentLoopRow.ParentPlay
	cp := cchk.CfgPlay
	app := cp.App
	cfg := app.GetConfig()

	thr := proc.CreateThread(app)

	command := make([]string, len(cchk.Command))
	copy(command, cchk.Command)

	registerVars := make(map[string]*registry.VarDef)
	for k, v := range parentPlay.RegisterVars {
		registerVars[k] = v
	}

	chk := &Chk{
		App: app,
		AppConfig: &snipplugin.AppConfig{
			DeploymentName: cfg.DeploymentName,
			SnippetsDir:    cfg.SnippetsDir,
			Runner:         cfg.Runner,
		},

		ParentLoopRow: parentLoopRow,
		Command:       command,

		IsPreRun: isPreRun,

		Middlewares: cchk.Middlewares,
		Runner:      cchk.Runner,

		Dir: cchk.Dir,

		RequiredFiles: cchk.RequiredFiles,

		Thread:      thr,
		ExecTimeout: parentPlay.ExecTimeout,

		RegisterVars: registerVars,
		Quiet:        cp.Quiet != nil && (*cp.Quiet),
	}

	depth := cchk.Depth
	if parentLoopRow.IsLoopRowItem {
		depth = depth + 1
	}
	chk.Depth = depth

	chk.TreeKeyParts = GetTreeKeyParts(chk.ParentLoopRow)
	chk.TreeKey = strings.Join(chk.TreeKeyParts, "|")

	vars := make(map[string]string)
	for k, v := range ctx.VarsDefault.Items() {
		vars[k] = v.(string)
	}
	for k, v := range ctx.Vars.Items() {
		vars[k] = v.(string)
	}
	chk.Vars = vars

	logger := logrus.WithFields(logrus.Fields{
		"tree":   chk.TreeKey,
		"action": "checking",
	})
	loggerCtx := context.WithValue(context.Background(), config.LogContextKey("indentation"), chk.Depth+1)
	logger = logger.WithContext(loggerCtx)
	chk.Logger = logger
	thr.Logger = logger

	if chk.ExecTimeout != nil {
		thr.SetTimeout(chk.ExecTimeout)
	}

	return chk
}

func (chk *Chk) RunThread() error {
	return chk.Thread.Run(chk.Main)
}
func (chk *Chk) Run() (bool, error) {
	logger := chk.Logger
	isPreRun := chk.IsPreRun

	var checkingAction string
	if isPreRun {
		checkingAction = "pre"
		logger.Info("⮞ pre-checking")
	} else {
		logger.Info("⮞ post-checking")
		checkingAction = "post"
	}

	err := chk.RunThread()

	var pState string
	var cState string
	var icon string
	if err == nil {
		icon = "🗹"
		cState = "success"
		if isPreRun {
			pState = "unchanged"
		} else {
			pState = "changed"
		}
	} else {
		cState = "failed"
		if isPreRun {
			icon = "𐄂"
			pState = "unready"
		} else {
			icon = "❎"
			pState = "error"
			chk.Thread.LogErrors()
		}
	}
	logger = logger.WithFields(logrus.Fields{
		"checkState":     cState,
		"playState":      pState,
		"checkingAction": checkingAction,
	})

	if err != nil {
		if isPreRun {
			logger.Info(icon + " check unready")
		} else {
			logger.Errorf(icon+" error: %v", err)
		}
		return false, err
	}

	logger.Info(icon + " check ok")
	return true, err
}

func (chk *Chk) CreateMutableCmd() *middleware.MutableCmd {
	originalVars := make(map[string]string)
	for k, v := range chk.Vars {
		originalVars[k] = v
	}
	originalCommand := make([]string, len(chk.Command))
	copy(originalCommand, chk.Command)

	requiredFiles := make(map[string]string)
	for k, v := range chk.RequiredFiles {
		requiredFiles[k] = v
	}

	vars := make(map[string]string)
	for k, v := range chk.Vars {
		vars[k] = v
	}

	mutableCmd := &middleware.MutableCmd{
		AppConfig:       chk.AppConfig,
		Command:         chk.Command,
		Vars:            vars,
		OriginalCommand: originalCommand,
		OriginalVars:    originalVars,
		RequiredFiles:   requiredFiles,
		Expect:          chk.Expect,
		Runner:          chk.Runner,
		Dir:             chk.Dir,
		Closer:          chk.Closer,
	}
	return mutableCmd
}

func (chk *Chk) GetPluginVarsMap(pluginType string, pluginName string, useVars []string, mVar map[string]*variable.Var) map[string]string {
	pVars := make(map[string]string)
	for _, useV := range useVars {

		var val string

		key := strings.ToUpper(useV)

		v := mVar[key]
		if v != nil && v.Default != "" {
			val = v.Default
		}

		k1 := strings.ToUpper("@" + key)
		if cv, ok := chk.Vars[k1]; ok {
			val = cv
		}
		k2 := strings.ToUpper("@" + pluginName + "_" + key)
		if cv, ok := chk.Vars[k2]; ok {
			val = cv
		}
		k3 := strings.ToUpper("@" + pluginType + "_" + pluginName + "_" + key)
		if cv, ok := chk.Vars[k3]; ok {
			val = cv
		}

		if v != nil && v.Value != "" {
			val = v.Value
		}

		pVars[strings.ToLower(key)] = val
	}
	return pVars
}

func (chk *Chk) ApplyMiddlewares() error {

	middlewareStack := chk.Middlewares

	mutableCmd := chk.CreateMutableCmd()

	for i := len(middlewareStack) - 1; i >= 0; i-- {

		cfgMiddleware := middlewareStack[i]

		middlewareVars := chk.GetPluginVarsMap("middleware", cfgMiddleware.Name, cfgMiddleware.Plugin.UseVars, cfgMiddleware.Vars)

		middlewareConfig := &middleware.Config{
			AppConfig:      chk.AppConfig,
			MiddlewareVars: middlewareVars,
			MutableCmd:     mutableCmd,
			Context:        chk.Thread.Context,
			ContextCancel:  chk.Thread.ContextCancel,
			Logger:         chk.Logger,
		}

		ok, err := cfgMiddleware.Plugin.Apply(middlewareConfig)
		if err != nil {
			return err
		}
		if !ok {
			break
		}
	}

	chk.Command = mutableCmd.Command
	chk.Vars = mutableCmd.Vars
	chk.RequiredFiles = mutableCmd.RequiredFiles
	chk.Expect = mutableCmd.Expect
	chk.Closer = mutableCmd.Closer
	chk.Dir = mutableCmd.Dir
	chk.Runner = mutableCmd.Runner

	return nil
}

func (chk *Chk) RegisterVarsLoad() {
	varsRegistry := chk.App.GetVarsRegistry()
	for i := 0; i < len(chk.TreeKeyParts); i++ {
		kp := chk.TreeKeyParts[0 : i+1]
		regVarsMap := varsRegistry.GetMapBySlice(kp)
		for k, v := range regVarsMap {
			chk.Vars[k] = v
		}
	}
}

func (chk *Chk) RunRunner() error {

	r := chk.Runner

	if r.Plugin == nil {
		r.Plugin = chk.App.GetRunner(r.Name)
	}

	runnerVars := chk.GetPluginVarsMap("runner", r.Name, r.Plugin.UseVars, r.Vars)

	vars := make(map[string]string)
	for k, v := range chk.Vars {
		vars[k] = v
	}

	registerVars := make(map[string]*registry.VarDef)
	for k, v := range chk.RegisterVars {
		registerVars[k] = v
	}

	runCfg := &runner.Config{
		AppConfig:     chk.AppConfig,
		RunnerVars:    runnerVars,
		Context:       chk.Thread.Context,
		ContextCancel: chk.Thread.ContextCancel,
		Logger: chk.Logger.WithFields(logrus.Fields{
			"runner": chk.Runner.Name,
		}),
		Cache:         chk.App.GetCache(),
		VarsRegistry:  chk.App.GetVarsRegistry(),
		Command:       chk.Command,
		Vars:          vars,
		RegisterVars:  registerVars,
		Quiet:         chk.Quiet,
		TreeKeyParts:  chk.TreeKeyParts,
		RequiredFiles: chk.RequiredFiles,
		Expect:        chk.Expect,
		Closer:        chk.Closer,
		Dir:           chk.Dir,
	}

	appCfg := chk.AppConfig
	rootPath := r.Plugin.GetRootPath(runCfg)
	vars["SNIP_SNIPPETS_PATH"] = filepath.Join(rootPath, "build", "snippets")
	vars["SNIP_LAUNCHER_PATH"] = filepath.Join(rootPath, "build", "launcher",
		appCfg.TreeDirLauncher(chk.TreeKeyParts))

	kp := chk.TreeKeyParts
	vars["SNIP_VARS_TREEPATH"] = filepath.Join(rootPath, "vars",
		appCfg.TreeDirVars(kp))

	return r.Plugin.Run(runCfg)
}

func (chk *Chk) Main() error {

	logger := chk.Logger

	if err := chk.ApplyMiddlewares(); err != nil {
		return err
	}

	// logger.Debugf("vars: %v", tools.JsonEncode(chk.Vars))
	logger.Debugf("env: %v", tools.JsonEncode(chk.EnvMap()))
	logger.Debugf("command: %v", strings.Join(chk.Command, " "))

	if err := chk.RunRunner(); err != nil {
		return err
	}

	return nil

}
