package play

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.com/youtopia.earth/ops/snip/errors"
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

type Cmd struct {
	App       App
	AppConfig *snipplugin.AppConfig

	Thread *proc.Thread

	ParentLoopRow *LoopRow
	CfgCmd        *CfgCmd

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

	RegisterVars   map[string]*registry.VarDef
	RegisterOutput string
	Quiet          bool

	TreeKeyParts []string
	TreeKey      string
}

func (cmd *Cmd) EnvMap() map[string]string {
	m := make(map[string]string)
	for k, v := range cmd.Vars {
		if k[0:1] != "@" {
			m[k] = v
		}
	}
	return m
}

func CreateCmd(ccmd *CfgCmd, ctx *RunCtx, parentLoopRow *LoopRow) *Cmd {
	parentPlay := parentLoopRow.ParentPlay
	cp := ccmd.CfgPlay
	app := cp.App
	cfg := app.GetConfig()

	thr := proc.CreateThread(app)

	command := make([]string, len(ccmd.Command))
	copy(command, ccmd.Command)

	registerVars := make(map[string]*registry.VarDef)
	for k, v := range parentPlay.RegisterVars {
		registerVars[k] = v
	}

	cmd := &Cmd{
		App: app,
		AppConfig: &snipplugin.AppConfig{
			DeploymentName: cfg.DeploymentName,
			SnippetsDir:    cfg.SnippetsDir,
			Runner:         cfg.Runner,
		},

		ParentLoopRow: parentLoopRow,
		CfgCmd:        ccmd,
		Command:       command,

		Middlewares: ccmd.Middlewares,
		Runner:      ccmd.Runner,

		Dir: ccmd.Dir,

		RequiredFiles: ccmd.RequiredFiles,

		Thread:      thr,
		ExecTimeout: parentPlay.ExecTimeout,

		RegisterVars:   registerVars,
		RegisterOutput: cp.RegisterOutput,
		Quiet:          cp.Quiet != nil && (*cp.Quiet),
	}

	depth := ccmd.Depth
	if parentLoopRow.IsLoopRowItem {
		depth = depth + 1
	}
	cmd.Depth = depth

	cmd.TreeKeyParts = GetTreeKeyParts(cmd.ParentLoopRow)
	cmd.TreeKey = strings.Join(cmd.TreeKeyParts, "|")

	vars := make(map[string]string)
	for k, v := range ctx.VarsDefault.Items() {
		vars[k] = v.(string)
	}
	for k, v := range ctx.Vars.Items() {
		vars[k] = v.(string)
	}
	cmd.Vars = vars

	logger := logrus.WithFields(logrus.Fields{
		"tree": cmd.TreeKey,
	})
	loggerCtx := context.WithValue(context.Background(), config.LogContextKey("indentation"), cmd.Depth+1)
	logger = logger.WithContext(loggerCtx)
	cmd.Logger = logger
	thr.Logger = logger

	if cmd.ExecTimeout != nil {
		thr.SetTimeout(cmd.ExecTimeout)
	}

	return cmd
}

var regNormalizeTreeKeyParts = regexp.MustCompile("[^a-zA-Z0-9-_.]+")

func (cmd *Cmd) Run() error {
	return cmd.Thread.Run(cmd.Main)
}

func (cmd *Cmd) CreateMutableCmd() *middleware.MutableCmd {
	originalVars := make(map[string]string)
	for k, v := range cmd.Vars {
		originalVars[k] = v
	}
	originalCommand := make([]string, len(cmd.Command))
	copy(originalCommand, cmd.Command)

	requiredFiles := make(map[string]string)
	for k, v := range cmd.RequiredFiles {
		requiredFiles[k] = v
	}

	vars := make(map[string]string)
	for k, v := range cmd.Vars {
		vars[k] = v
	}

	mutableCmd := &middleware.MutableCmd{
		AppConfig:       cmd.AppConfig,
		Command:         cmd.Command,
		Vars:            vars,
		OriginalCommand: originalCommand,
		OriginalVars:    originalVars,
		RequiredFiles:   requiredFiles,
		Expect:          cmd.Expect,
		Runner:          cmd.Runner,
		Dir:             cmd.Dir,
		Closer:          cmd.Closer,
	}
	return mutableCmd
}

func (cmd *Cmd) GetPluginVarsMap(pluginType string, pluginName string, useVars []string, mVar map[string]*variable.Var) map[string]string {
	pVars := make(map[string]string)
	for _, useV := range useVars {

		var val string

		key := strings.ToUpper(useV)

		v := mVar[key]
		if v != nil && v.Default != "" {
			val = v.Default
		}

		k1 := strings.ToUpper("@" + key)
		if cv, ok := cmd.Vars[k1]; ok {
			val = cv
		}
		k2 := strings.ToUpper("@" + pluginName + "_" + key)
		if cv, ok := cmd.Vars[k2]; ok {
			val = cv
		}
		k3 := strings.ToUpper("@" + pluginType + "_" + pluginName + "_" + key)
		if cv, ok := cmd.Vars[k3]; ok {
			val = cv
		}

		if v != nil && v.Value != "" {
			val = v.Value
		}

		pVars[strings.ToLower(key)] = val
	}
	return pVars
}

func (cmd *Cmd) ApplyMiddlewares() error {

	middlewareStack := cmd.Middlewares

	mutableCmd := cmd.CreateMutableCmd()

	for i := len(middlewareStack) - 1; i >= 0; i-- {

		cfgMiddleware := middlewareStack[i]

		middlewareVars := cmd.GetPluginVarsMap("middleware", cfgMiddleware.Name, cfgMiddleware.Plugin.UseVars, cfgMiddleware.Vars)

		middlewareConfig := &middleware.Config{
			AppConfig:      cmd.AppConfig,
			MiddlewareVars: middlewareVars,
			MutableCmd:     mutableCmd,
			Context:        cmd.Thread.Context,
			ContextCancel:  cmd.Thread.ContextCancel,
			Logger:         cmd.Logger,
		}

		ok, err := cfgMiddleware.Plugin.Apply(middlewareConfig)
		if err != nil {
			return err
		}
		if !ok {
			break
		}
	}

	cmd.Command = mutableCmd.Command
	cmd.Vars = mutableCmd.Vars
	cmd.RequiredFiles = mutableCmd.RequiredFiles
	cmd.Expect = mutableCmd.Expect
	cmd.Closer = mutableCmd.Closer
	cmd.Dir = mutableCmd.Dir
	cmd.Runner = mutableCmd.Runner

	return nil
}

func (cmd *Cmd) BuildLauncher() error {

	appCfg := cmd.AppConfig
	rootPath := GetRootPath(cmd.App)
	treeDir := appCfg.TreeDirLauncher(cmd.TreeKeyParts)
	launcherFilename := "run.sh"
	launcherDir := filepath.Join("build", "launcher", treeDir)
	launcherFile := filepath.Join(launcherDir, launcherFilename)
	launcherDirAbs := filepath.Join(rootPath, launcherDir)
	launcherFileAbs := filepath.Join(rootPath, launcherFile)

	if err := os.MkdirAll(launcherDirAbs, os.ModePerm); err != nil {
		return err
	}

	f, err := os.OpenFile(launcherFileAbs, os.O_APPEND|os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	outputAppend := func(str string) {
		_, err := f.WriteString(str)
		errors.Check(err)
	}

	now := time.Now()
	nowText := now.Format("2006-01-02 15:04:05")
	outputAppend("#!/bin/sh\n")
	outputAppend("# play launcher generated by snip on " + nowText + "\n")
	outputAppend("exec " + cmd.Command[0] + ` $@`)
	if cmd.RegisterOutput != "" {
		vr := strings.ToUpper(cmd.RegisterOutput)
		outputAppend(` | tee ${SNIP_VARS_TREEPATH}/` + vr)
	}
	outputAppend("\n")

	cmd.Logger.Debugf("writed launcher to %s", launcherFile)

	bin := filepath.Join("${SNIP_LAUNCHER_PATH}", launcherFilename)
	cmd.Command[0] = bin

	cmd.RequiredFiles[launcherFileAbs] = launcherFile

	return nil
}

func (cmd *Cmd) RegisterVarsLoad() {
	varsRegistry := cmd.App.GetVarsRegistry()
	for i := 0; i < len(cmd.TreeKeyParts); i++ {
		kp := cmd.TreeKeyParts[0 : i+1]
		regVarsMap := varsRegistry.GetMapBySlice(kp)
		for k, v := range regVarsMap {
			cmd.Vars[k] = v
		}
	}
}

func (cmd *Cmd) RunRunner() error {

	r := cmd.Runner

	if r.Plugin == nil {
		r.Plugin = cmd.App.GetRunner(r.Name)
	}

	runnerVars := cmd.GetPluginVarsMap("runner", r.Name, r.Plugin.UseVars, r.Vars)

	vars := make(map[string]string)
	for k, v := range cmd.Vars {
		vars[k] = v
	}

	registerVars := make(map[string]*registry.VarDef)
	for k, v := range cmd.RegisterVars {
		registerVars[k] = v
	}

	runCfg := &runner.Config{
		AppConfig:     cmd.AppConfig,
		RunnerVars:    runnerVars,
		Context:       cmd.Thread.Context,
		ContextCancel: cmd.Thread.ContextCancel,
		Logger: cmd.Logger.WithFields(logrus.Fields{
			"runner": cmd.Runner.Name,
		}),
		Cache:          cmd.App.GetCache(),
		VarsRegistry:   cmd.App.GetVarsRegistry(),
		Command:        cmd.Command,
		Vars:           vars,
		RegisterVars:   registerVars,
		RegisterOutput: cmd.RegisterOutput,
		Quiet:          cmd.Quiet,
		TreeKeyParts:   cmd.TreeKeyParts,
		RequiredFiles:  cmd.RequiredFiles,
		Expect:         cmd.Expect,
		Closer:         cmd.Closer,
		Dir:            cmd.Dir,
	}

	appCfg := cmd.AppConfig
	rootPath := r.Plugin.GetRootPath(runCfg)
	vars["SNIP_SNIPPETS_PATH"] = filepath.Join(rootPath, "build", "snippets")
	vars["SNIP_LAUNCHER_PATH"] = filepath.Join(rootPath, "build", "launcher",
		appCfg.TreeDirLauncher(cmd.TreeKeyParts))
	vars["SNIP_VARS_TREEPATH"] = filepath.Join(rootPath, "vars",
		appCfg.TreeDirVars(cmd.TreeKeyParts))

	return r.Plugin.Run(runCfg)
}

func (cmd *Cmd) Main() error {

	logger := cmd.Logger
	logger.Info("⮞ playing")

	if err := cmd.BuildLauncher(); err != nil {
		return err
	}

	if err := cmd.ApplyMiddlewares(); err != nil {
		return err
	}

	// logger.Debugf("vars: %v", tools.JsonEncode(cmd.Vars))
	logger.Debugf("env: %v", tools.JsonEncode(cmd.EnvMap()))
	logger.Debugf("command: %v", strings.Join(cmd.Command, " "))

	if err := cmd.RunRunner(); err != nil {
		return err
	}

	return nil

}
