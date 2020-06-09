package play

import (
	"context"
	"os/exec"
	"strings"
	"time"

	shellquote "github.com/kballard/go-shellquote"
	"github.com/opencontainers/runc/libcontainer/user"
	"github.com/sirupsen/logrus"

	"gitlab.com/youtopia.earth/ops/snip/config"
	"gitlab.com/youtopia.earth/ops/snip/middleware"
	"gitlab.com/youtopia.earth/ops/snip/proc"
	"gitlab.com/youtopia.earth/ops/snip/tools"
)

type Cmd struct {
	App App

	Thread *proc.Thread

	ParentLoopRow *LoopRow
	CfgCmd        *CfgCmd

	Command string
	Args    []string
	Vars    map[string]string

	Middlewares []string

	ExecUser    *user.ExecUser
	ExecTimeout *time.Duration

	IsMD   bool
	Logger *logrus.Entry
	Depth  int
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

	cmd := &Cmd{
		App:           app,
		ParentLoopRow: parentLoopRow,
		CfgCmd:        ccmd,
		Command:       ccmd.Command,
		Args:          ccmd.Args,
		IsMD:          ccmd.IsMD,
		Middlewares:   parentPlay.Middlewares,
		Thread:        proc.CreateThread(app),
		ExecUser:      parentPlay.ExecUser,
		ExecTimeout:   parentPlay.ExecTimeout,
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
	cmd.Thread.Logger = logger

	cmd.Thread.ExecUser = cmd.ExecUser
	cmd.Thread.ExecTimeout = cmd.ExecTimeout

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

func (cmd *Cmd) Run() error {
	return cmd.Thread.Run(cmd.Main)
}

func (cmd *Cmd) RunFunc() error {
	commandSlice := append([]string{cmd.Command}, cmd.Args...)
	commandHook := func(c *exec.Cmd) error {
		c.Env = tools.EnvToPairs(cmd.EnvMap())
		return nil
	}
	cmd.Logger.Debugf("command: %v", shellquote.Join(commandSlice...))
	return cmd.Thread.RunCmd(commandSlice, cmd.Logger, commandHook)
}

func (cmd *Cmd) Main() error {

	app := cmd.App

	cmd.Logger.Info("⮞ playing")
	cmd.Logger.Debugf("vars: %v", tools.JsonEncode(cmd.Vars))

	var runStack []middleware.Func

	for _, k := range cmd.Middlewares {
		middleware := app.GetMiddleware(k)
		runStack = append(runStack, middleware)
	}
	runStack = append(runStack, func(middlewareConfig *middleware.Config, next func() error) error {
		mutableCmd := middlewareConfig.MutableCmd
		cmd.Command = mutableCmd.Command
		cmd.Args = mutableCmd.Args
		cmd.Vars = mutableCmd.Vars
		return cmd.RunFunc()
	})

	originalVars := make(map[string]string)
	for k, v := range cmd.Vars {
		originalVars[k] = v
	}
	originalArgs := make([]string, len(cmd.Args))
	for i, v := range cmd.Args {
		originalArgs[i] = v
	}
	mutableCmd := &middleware.MutableCmd{
		Command:         cmd.Command,
		Args:            cmd.Args,
		Vars:            cmd.Vars,
		OriginalCommand: cmd.Command,
		OriginalArgs:    originalArgs,
		OriginalVars:    originalVars,
	}
	middlewareConfig := &middleware.Config{
		MutableCmd:    mutableCmd,
		Context:       cmd.Thread.Context,
		ContextCancel: cmd.Thread.ContextCancel,
		Logger:        cmd.Logger,
	}

	wrapped := func() error {
		return nil
	}
	for i := len(runStack) - 1; i >= 0; i-- {
		current := runStack[i]
		next := wrapped
		wrapped = func() error {
			return current(middlewareConfig, next)
		}
	}

	return wrapped()

}
