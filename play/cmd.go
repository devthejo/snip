package play

import (
	"context"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"gitlab.com/ytopia/ops/snip/config"
	expect "gitlab.com/ytopia/ops/snip/goexpect"
	snipplugin "gitlab.com/ytopia/ops/snip/plugin"
	"gitlab.com/ytopia/ops/snip/plugin/middleware"
	"gitlab.com/ytopia/ops/snip/plugin/processor"
	"gitlab.com/ytopia/ops/snip/plugin/runner"
	"gitlab.com/ytopia/ops/snip/proc"
	"gitlab.com/ytopia/ops/snip/registry"
)

type Cmd struct {
	App       App
	AppConfig *snipplugin.AppConfig

	Thread *proc.Thread

	ParentLoopRow *LoopRow

	Command []string
	RunVars *RunVars

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

	RegisterVars map[string]*registry.VarDef
	Quiet        bool

	TreeKeyParts []string
	TreeKey      string

	PreflightRunnedOnce bool

	Tmpdir bool
}

func CreateCmd(ccmd *CfgCmd, parentLoopRow *LoopRow) *Cmd {
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
		Command:       command,

		Middlewares: ccmd.Middlewares,
		Runner:      ccmd.Runner,

		Dir: ccmd.Dir,

		RequiredFiles:              ccmd.RequiredFiles,
		RequiredFilesSrcProcessors: ccmd.RequiredFilesSrcProcessors,

		Thread:      thr,
		ExecTimeout: parentPlay.ExecTimeout,

		RegisterVars: registerVars,
		Quiet:        cp.Quiet != nil && (*cp.Quiet),

		RunVars: parentLoopRow.RunVars,
	}
	if cp.Tmpdir == nil {
		cmd.Tmpdir = true
	} else {
		cmd.Tmpdir = (*cp.Tmpdir)
	}

	depth := ccmd.Depth
	if parentLoopRow.IsLoopRowItem {
		depth = depth + 1
	}
	cmd.Depth = depth

	cmd.TreeKeyParts = GetTreeKeyParts(cmd.ParentLoopRow)
	cmd.TreeKey = strings.Join(cmd.TreeKeyParts, "|")

	logger := logrus.WithFields(logrus.Fields{
		"tree":   cmd.TreeKey,
		"action": "running",
	})
	loggerCtx := context.WithValue(context.Background(), config.LogContextKey("indentation"), cmd.Depth+1)
	logger = logger.WithContext(loggerCtx)
	cmd.Logger = logger
	thr.Logger = logger

	return cmd
}

func (cmd *Cmd) Run() error {
	cmd.RegisterVarsLoad()

	if cmd.Thread.ExecExited {
		cmd.Thread.Reset()
	}

	err := cmd.Thread.Run(cmd.Main)
	cmd.Thread.LogErrors()
	return err
}

func (cmd *Cmd) CreateMutableCmd() *middleware.MutableCmd {
	originalCommand := make([]string, len(cmd.Command))
	copy(originalCommand, cmd.Command)

	requiredFiles := make(map[string]string)
	for k, v := range cmd.RequiredFiles {
		requiredFiles[k] = v
	}

	requiredFilesProcessors := make(map[string][]func(*processor.Config, *string) error)
	for k, v := range cmd.RequiredFilesSrcProcessors {
		requiredFilesProcessors[k] = v
	}

	mutableCmd := &middleware.MutableCmd{
		AppConfig:                  cmd.AppConfig,
		Command:                    cmd.Command,
		OriginalCommand:            originalCommand,
		RequiredFiles:              requiredFiles,
		RequiredFilesSrcProcessors: requiredFilesProcessors,
		Expect:                     cmd.Expect,
		Runner:                     cmd.Runner,
		Dir:                        cmd.Dir,
		Closer:                     cmd.Closer,
	}
	return mutableCmd
}

func (cmd *Cmd) ApplyMiddlewares() error {

	middlewareStack := cmd.Middlewares

	mutableCmd := cmd.CreateMutableCmd()

	for i := len(middlewareStack) - 1; i >= 0; i-- {

		cfgMiddleware := middlewareStack[i]

		middlewareVars := cmd.RunVars.GetPluginVars("middleware", cfgMiddleware.Name, cfgMiddleware.Plugin.UseVars, cfgMiddleware.Vars)

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
	cmd.RequiredFiles = mutableCmd.RequiredFiles
	cmd.RequiredFilesSrcProcessors = mutableCmd.RequiredFilesSrcProcessors
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
	launcherFilename := "play.sh"
	launcherDir := filepath.Join("build", "launcher", treeDir)
	launcherFile := filepath.Join(launcherDir, launcherFilename)
	launcherDirAbs := filepath.Join(rootPath, launcherDir)
	launcherFileAbs := filepath.Join(rootPath, launcherFile)

	if err := os.MkdirAll(launcherDirAbs, os.ModePerm); err != nil {
		return err
	}

	launcherContent := "#!/usr/bin/env bash\n"

	if cmd.Tmpdir {
		usr, _ := user.Current()
		rootPath := filepath.Join(usr.HomeDir, ".snip", appCfg.DeploymentName, "tmpdir")
		if err := os.MkdirAll(rootPath, os.ModePerm); err != nil {
			return err
		}
		tempDir, err := ioutil.TempDir(rootPath, "tmpdir-")
		if err != nil {
			return err
		}
		// bugfix hack, sometime the tempdir is not created by ioutil.TempDir
		if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
			return err
		}
		launcherContent += "cd " + tempDir + "\n"
	}

	launcherContent += "exec " + cmd.Command[0] + " $@> >(tee ${SNIP_VARS_TREEPATH}/raw.stdout)"

	err := ioutil.WriteFile(launcherFileAbs, []byte(launcherContent), 0755)
	if err != nil {
		return err
	}

	cmd.Logger.Debugf("writed play launcher to %s", launcherFile)

	bin := filepath.Join("${SNIP_LAUNCHER_PATH}", launcherFilename)
	cmd.Command[0] = bin

	cmd.RequiredFiles[launcherFile] = launcherFileAbs

	return nil
}

func (cmd *Cmd) RegisterVarsLoad() {
	varsRegistry := cmd.App.GetVarsRegistry()
	for i := 0; i < len(cmd.TreeKeyParts); i++ {
		kp := cmd.TreeKeyParts[0 : i+1]
		regVarsMap := varsRegistry.GetMapBySlice(kp)
		for k, v := range regVarsMap {
			cmd.RunVars.SetValueString(k, v)
		}
	}
}

func (cmd *Cmd) RunRunner() error {

	r := cmd.Runner

	if r.Plugin == nil {
		r.Plugin = cmd.App.GetRunner(r.Name)
	}

	runnerVars := cmd.RunVars.GetPluginVars("runner", r.Name, r.Plugin.UseVars, r.Vars)

	registerVars := make(map[string]*registry.VarDef)
	for k, v := range cmd.RegisterVars {
		registerVars[k] = v
	}

	if cmd.ExecTimeout != nil {
		cmd.Thread.SetTimeout(cmd.ExecTimeout)
	}

	vars := cmd.RunVars.GetAll()

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
		Vars:          vars,
		RegisterVars:  registerVars,
		Quiet:         cmd.Quiet,
		TreeKeyParts:  cmd.TreeKeyParts,
		RequiredFiles: cmd.RequiredFiles,
		Expect:        cmd.Expect,
		Closer:        cmd.Closer,
		Dir:           cmd.Dir,
	}

	appCfg := cmd.AppConfig
	rootPath := r.Plugin.GetRootPath(runCfg)
	vars["SNIP_SNIPPETS_PATH"] = filepath.Join(rootPath, "build", "snippets")
	vars["SNIP_LAUNCHER_PATH"] = filepath.Join(rootPath, "build", "launcher",
		appCfg.TreeDirLauncher(cmd.TreeKeyParts))

	kp := cmd.TreeKeyParts
	vars["SNIP_VARS_TREEPATH"] = filepath.Join(rootPath, "vars",
		appCfg.TreeDirVars(kp))

	return r.Plugin.Run(runCfg)
}

func (cmd *Cmd) PreflightRun() error {
	if cmd.PreflightRunnedOnce {
		return nil
	}
	cmd.PreflightRunnedOnce = true
	logger := cmd.Logger
	logger.Debug("⮞ preflight")
	if err := cmd.BuildLauncher(); err != nil {
		return err
	}
	if err := cmd.ApplyMiddlewares(); err != nil {
		return err
	}
	processorCfg := &processor.Config{
		RunVars: cmd.RunVars,
	}
	for dest, src := range cmd.RequiredFiles {
		if processors, ok := cmd.RequiredFilesSrcProcessors[src]; ok {
			for _, processor := range processors {
				err := processor(processorCfg, &src)
				if err != nil {
					logrus.Fatal(err)
					return err
				}
			}
			cmd.RequiredFiles[dest] = src
		}
	}

	return nil
}

func (cmd *Cmd) Main() error {

	logger := cmd.Logger
	logger.Info("⮞ playing")

	if err := cmd.RunRunner(); err != nil {
		return err
	}

	return nil

}
