package play

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"gitlab.com/ytopia/ops/snip/config"
	"gitlab.com/ytopia/ops/snip/goenv"
	expect "gitlab.com/ytopia/ops/snip/goexpect"
	snipplugin "gitlab.com/ytopia/ops/snip/plugin"
	"gitlab.com/ytopia/ops/snip/plugin/middleware"
	"gitlab.com/ytopia/ops/snip/plugin/processor"
	"gitlab.com/ytopia/ops/snip/plugin/runner"
	"gitlab.com/ytopia/ops/snip/proc"
	"gitlab.com/ytopia/ops/snip/tools"
	"gitlab.com/ytopia/ops/snip/variable"
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

	RequiredFiles              map[string]string
	RequiredFilesSrcProcessors map[string][]func(*processor.Config, *string) error

	Expect []expect.Batcher

	Dir string

	Closer *func(interface{}, *string) bool

	Quiet bool

	TreeKeyParts []string
	TreeKey      string

	PreflightRunnedOnce bool

	RunReport *RunReport
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

func CreateChk(cchk *CfgChk, ctx *RunCtx, parentLoopRow *LoopRow, isPreRun bool) *Chk {
	parentPlay := parentLoopRow.ParentPlay
	cp := cchk.CfgPlay
	app := cp.App
	cfg := app.GetConfig()

	thr := proc.CreateThread(app)

	command := make([]string, len(cchk.Command))
	copy(command, cchk.Command)

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

		RequiredFiles:              cchk.RequiredFiles,
		RequiredFilesSrcProcessors: cchk.RequiredFilesSrcProcessors,

		Thread:      thr,
		ExecTimeout: parentPlay.ExecTimeout,

		Quiet: cp.Quiet != nil && (*cp.Quiet),

		RunReport: cp.RunReport,
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
	for k, v := range vars {
		vars[k], _ = goenv.Expand(v, vars)
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

	return chk
}

func (chk *Chk) RunThread() error {
	if chk.Thread.ExecExited {
		chk.Thread.Reset()
	}

	return chk.Thread.Run(chk.Main)
}
func (chk *Chk) Run() (bool, error) {

	app := chk.App
	if app.IsExiting() {
		return false, nil
	}

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

	ok := err == nil

	au := app.GetAurora()
	runReport := chk.RunReport

	if !app.IsExiting() {
		if ok {
			if isPreRun {
				logger.Info(au.BrightGreen(icon + " check ok"))
				runReport.OK++
			} else {
				logger.Info(au.BrightMagenta(icon + " check ok (changed)"))
				runReport.Changed++
			}
		} else {
			if isPreRun {
				logger.Info(au.BrightCyan(icon + " check unready"))
			} else {
				logger.Errorf(icon+" error: %v", err)
			}
		}
	}

	return ok, err
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

	requiredFilesProcessors := make(map[string][]func(*processor.Config, *string) error)
	for k, v := range chk.RequiredFilesSrcProcessors {
		requiredFilesProcessors[k] = v
	}

	vars := make(map[string]string)
	for k, v := range chk.Vars {
		vars[k] = v
	}

	mutableCmd := &middleware.MutableCmd{
		AppConfig:                  chk.AppConfig,
		Command:                    chk.Command,
		Vars:                       vars,
		OriginalCommand:            originalCommand,
		OriginalVars:               originalVars,
		RequiredFiles:              requiredFiles,
		RequiredFilesSrcProcessors: requiredFilesProcessors,
		Expect:                     chk.Expect,
		Runner:                     chk.Runner,
		Dir:                        chk.Dir,
		Closer:                     chk.Closer,
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

func (chk *Chk) BuildLauncher() error {

	appCfg := chk.AppConfig
	rootPath := GetRootPath(chk.App)
	treeDir := appCfg.TreeDirLauncher(chk.TreeKeyParts)
	launcherFilename := "check.bash"
	launcherDir := filepath.Join("build", "launcher", treeDir)
	launcherFile := filepath.Join(launcherDir, launcherFilename)
	launcherDirAbs := filepath.Join(rootPath, launcherDir)
	launcherFileAbs := filepath.Join(rootPath, launcherFile)

	if err := os.MkdirAll(launcherDirAbs, os.ModePerm); err != nil {
		return err
	}

	launcherContent := "#!/usr/bin/env bash\n"
	launcherContent += "set -e\n"
	launcherContent += strings.Join(chk.Command, " ")

	err := ioutil.WriteFile(launcherFileAbs, []byte(launcherContent), 0755)
	if err != nil {
		return err
	}

	chk.Logger.Debugf("writed check launcher to %s", launcherFile)

	bin := filepath.Join("${SNIP_LAUNCHER_PATH}", launcherFilename)
	chk.Command = []string{bin}

	chk.RequiredFiles[launcherFile] = launcherFileAbs

	return nil
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
	chk.RequiredFilesSrcProcessors = mutableCmd.RequiredFilesSrcProcessors
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

	if chk.ExecTimeout != nil {
		chk.Thread.SetTimeout(chk.ExecTimeout)
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

	if !chk.PreflightRunnedOnce {
		if err := chk.BuildLauncher(); err != nil {
			return err
		}
		if err := chk.ApplyMiddlewares(); err != nil {
			return err
		}
		chk.PreflightRunnedOnce = true
	}

	// logger.Debugf("vars: %v", tools.JsonEncode(chk.Vars))
	logger.Debugf("env: %v", tools.JsonEncode(chk.EnvMap()))
	logger.Debugf("command: %v", strings.Join(chk.Command, " "))

	if err := chk.RunRunner(); err != nil {
		return err
	}

	return nil

}
