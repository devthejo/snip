package play

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	expect "gitlab.com/youtopia.earth/ops/snip/goexpect"

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
	Stdin  io.Reader

	Dir string

	Closer *func(interface{}) bool

	RegisterVars []string
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

	vars := make(map[string]string)
	for k, v := range ctx.VarsDefault.Items() {
		vars[k] = v.(string)
	}
	for k, v := range ctx.Vars.Items() {
		vars[k] = v.(string)
	}
	cmd.Vars = vars

	logKey := cmd.GetTreeKey()
	logger := logrus.WithFields(logrus.Fields{
		"tree": logKey,
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

func (cmd *Cmd) GetTreeKey() string {
	parts := cmd.GetTreeKeyParts()
	for i, v := range parts {
		parts[i] = strings.ReplaceAll(v, "|", "-")
	}
	return strings.Join(parts, "|")

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
		parts = append([]string{part}, parts...)
	}
	return parts
}

func (cmd *Cmd) ExpandCmdEnvMapper(key string) string {
	if val, ok := cmd.Vars[key]; ok {
		return val
	}
	return ""
}
func (cmd *Cmd) ExpandCmdEnv(commandSlice []string) []string {
	expandedCmd := make([]string, len(commandSlice))
	for i, str := range commandSlice {
		expandedCmd[i] = os.Expand(str, cmd.ExpandCmdEnvMapper)
	}
	return expandedCmd
}

func (cmd *Cmd) Run() error {
	return cmd.Thread.Run(cmd.Main)
}

func (cmd *Cmd) CreateMutableCmd() *snipplugin.MutableCmd {
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

	mutableCmd := &snipplugin.MutableCmd{
		AppConfig:       cmd.AppConfig,
		Command:         cmd.Command,
		Vars:            cmd.Vars,
		OriginalCommand: originalCommand,
		OriginalVars:    originalVars,
		RequiredFiles:   requiredFiles,
		Expect:          cmd.Expect,
		Runner:          cmd.CfgCmd.CfgPlay.Runner,
		Stdin:           cmd.Stdin,
		Dir:             cmd.Dir,
		Closer:          cmd.Closer,
	}
	return mutableCmd
}

func (cmd *Cmd) ApplyMiddlewares() error {
	var middlewareStack []*middleware.Middleware

	for _, middleware := range cmd.Middlewares {
		middlewareStack = append(middlewareStack, middleware)
	}

	mutableCmd := cmd.CreateMutableCmd()
	middlewareConfig := &middleware.Config{
		AppConfig:     cmd.AppConfig,
		MutableCmd:    mutableCmd,
		Context:       cmd.Thread.Context,
		ContextCancel: cmd.Thread.ContextCancel,
		Logger:        cmd.Logger,
	}

	wrapped := func() (bool, error) {
		return false, nil
	}
	for i := len(middlewareStack) - 1; i >= 0; i-- {
		current := middlewareStack[i]
		next := wrapped
		wrapped = func() (bool, error) {
			ok, err := current.Apply(middlewareConfig)
			if err != nil || !ok {
				return ok, err
			}
			return next()
		}
	}

	if _, err := wrapped(); err != nil {
		return err
	}

	cmd.Command = mutableCmd.Command
	cmd.Vars = mutableCmd.Vars
	cmd.RequiredFiles = mutableCmd.RequiredFiles
	cmd.Expect = mutableCmd.Expect
	cmd.Stdin = mutableCmd.Stdin
	cmd.Closer = mutableCmd.Closer
	cmd.Dir = mutableCmd.Dir
	if mutableCmd.Runner != cmd.CfgCmd.CfgPlay.Runner {
		cmd.Runner = cmd.App.GetRunner(mutableCmd.Runner)
	}

	return nil
}

func (cmd *Cmd) RunRunner() error {

	r := cmd.Runner

	runCfg := &runner.Config{
		AppConfig:     cmd.AppConfig,
		Context:       cmd.Thread.Context,
		ContextCancel: cmd.Thread.ContextCancel,
		Logger:        cmd.Logger,
		Cache:         cmd.App.GetCache(),
		Vars:          cmd.Vars,
		Command:       cmd.Command,
		RequiredFiles: cmd.RequiredFiles,
		Expect:        cmd.Expect,
		Stdin:         cmd.Stdin,
		Closer:        cmd.Closer,
		Dir:           cmd.Dir,
	}
	return r.Run(runCfg)
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
