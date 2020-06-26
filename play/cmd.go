package play

import (
	"context"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	expect "gitlab.com/youtopia.earth/ops/snip/goexpect"
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

	RegisterVars []string

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
	app := ccmd.CfgPlay.App
	cfg := app.GetConfig()

	thr := proc.CreateThread(app)

	cmd := &Cmd{
		App: app,
		AppConfig: &snipplugin.AppConfig{
			DeploymentName: cfg.DeploymentName,
			BuildDir:       cfg.BuildDir,
			SnippetsDir:    cfg.SnippetsDir,
			Runner:         cfg.Runner,
		},

		ParentLoopRow: parentLoopRow,
		CfgCmd:        ccmd,
		Command:       ccmd.Command,

		Middlewares: ccmd.Middlewares,
		Runner:      ccmd.Runner,

		Dir: ccmd.Dir,

		RequiredFiles: ccmd.RequiredFiles,

		Thread:      thr,
		ExecTimeout: parentPlay.ExecTimeout,

		RegisterVars: parentPlay.RegisterVars,
	}

	depth := ccmd.Depth
	if parentLoopRow.IsLoopRowItem {
		depth = depth + 1
	}
	cmd.Depth = depth

	cmd.TreeKeyParts = cmd.GetTreeKeyParts()
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

func (cmd *Cmd) GetTreeKeyParts() []string {
	var parts []string
	var parent interface{}
	parent = cmd.ParentLoopRow
	for {
		var part string
		switch p := parent.(type) {
		case *LoopRow:
			if p == nil {
				parent = nil
				break
			}
			part = p.GetKey()
			parent = p.ParentPlay
		case *Play:
			if p == nil {
				parent = nil
				break
			}
			part = p.GetKey()
			parent = p.ParentLoopRow
		case nil:
			parent = nil
		}
		if parent == nil {
			break
		}
		part = strings.ReplaceAll(part, "|", "-")
		part = strings.ReplaceAll(part, "/", "_")
		parts = append([]string{part}, parts...)
	}
	return parts
}

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
	kp := cmd.TreeKeyParts
	kp = kp[0 : len(kp)-2]
	varsRegistry := cmd.App.GetVarsRegistry()
	regVarsMap := varsRegistry.GetMapBySlice(kp)
	for k, v := range regVarsMap {
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

func (cmd *Cmd) RunRunner() error {

	r := cmd.Runner

	if r.Plugin == nil {
		r.Plugin = cmd.App.GetRunner(r.Name)
	}

	runnerVars := cmd.GetPluginVarsMap("runner", r.Name, r.Plugin.UseVars, r.Vars)

	runCfg := &runner.Config{
		AppConfig:     cmd.AppConfig,
		RunnerVars:    runnerVars,
		Context:       cmd.Thread.Context,
		ContextCancel: cmd.Thread.ContextCancel,
		Logger: cmd.Logger.WithFields(logrus.Fields{
			"runner": cmd.Runner.Name,
		}),
		Cache:         cmd.App.GetCache(),
		VarsRegistry:  cmd.App.GetVarsRegistry(),
		Command:       cmd.Command,
		Vars:          cmd.Vars,
		RegisterVars:  cmd.RegisterVars,
		TreeKeyParts:  cmd.TreeKeyParts,
		RequiredFiles: cmd.RequiredFiles,
		Expect:        cmd.Expect,
		Closer:        cmd.Closer,
		Dir:           cmd.Dir,
	}

	return r.Plugin.Run(runCfg)
}

func (cmd *Cmd) Main() error {

	logger := cmd.Logger
	logger.Info("⮞ playing")

	var err error

	err = cmd.ApplyMiddlewares()
	if err != nil {
		return err
	}

	// logger.Debugf("vars: %v", tools.JsonEncode(cmd.Vars))
	logger.Debugf("env: %v", tools.JsonEncode(cmd.EnvMap()))
	logger.Debugf("command: %v", strings.Join(cmd.Command, " "))

	err = cmd.RunRunner()
	if err != nil {
		return err
	}

	return nil

}
