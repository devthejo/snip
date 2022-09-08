package play

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/devthejo/snip/config"
	expect "github.com/devthejo/snip/goexpect"
	snipplugin "github.com/devthejo/snip/plugin"
	"github.com/devthejo/snip/plugin/middleware"
	"github.com/devthejo/snip/plugin/processor"
	"github.com/devthejo/snip/plugin/runner"
	"github.com/devthejo/snip/proc"
)

type Chk struct {
	App       App
	AppConfig *snipplugin.AppConfig

	Thread *proc.Thread

	ParentLoopRow *LoopRow

	IsPreRun bool

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

	Quiet bool

	TreeKeyParts []string
	TreeKey      string

	PreflightRunnedOnce bool

	GlobalRunCtx *GlobalRunCtx

	Retry    interface{}
	Interval time.Duration
	Timeout  time.Duration
}

func CreateChk(cchk *CfgChk, parentLoopRow *LoopRow, isPreRun bool) *Chk {
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

		GlobalRunCtx: cp.GlobalRunCtx,

		RunVars: parentLoopRow.RunVars,
	}

	depth := cchk.Depth
	if parentLoopRow.IsLoopRowItem {
		depth = depth + 1
	}
	chk.Depth = depth

	chk.TreeKeyParts = GetTreeKeyParts(chk.ParentLoopRow)
	chk.TreeKey = strings.Join(chk.TreeKeyParts, "|")

	if isPreRun {
		chk.Retry = cp.PreCheckRetry
		chk.Interval = cp.PreCheckInterval
		chk.Timeout = cp.PreCheckTimeout
	} else {
		chk.Retry = cp.PostCheckRetry
		chk.Interval = cp.PostCheckInterval
		chk.Timeout = cp.PostCheckTimeout
	}

	if isPreRun {
		chk.Quiet = cp.PreCheckQuiet != nil && (*cp.PreCheckQuiet)
	} else {
		chk.Quiet = cp.PostCheckQuiet != nil && (*cp.PostCheckQuiet)
	}

	loggerCtx := context.WithValue(context.Background(), config.LogContextKey("indentation"), chk.Depth+1)
	logger := logrus.WithFields(logrus.Fields{
		"tree":   chk.TreeKey,
		"action": "checking",
	}).WithContext(loggerCtx)
	chk.Logger = logger
	thr.Logger = logger

	return chk
}

func (chk *Chk) RunThreadLoop() error {

	log := chk.Logger

	startedTime := time.Now()
	var timeoutTime time.Time
	ctx := context.Background()
	var cancel context.CancelFunc
	if chk.Timeout != 0 {
		timeoutTime = startedTime.Add(chk.Timeout)
		ctx, cancel = context.WithTimeout(ctx, chk.Timeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}

	mainProc := chk.Thread.App.GetMainProc()

	go func() {
		<-mainProc.Done()
		cancel()
	}()

	retry := chk.Retry
	var isRetryTypeInt bool
	var retryTypeInt int
	var retryTypeBool bool
	switch r := retry.(type) {
	case *int:
		retryTypeInt = *r
		isRetryTypeInt = true
	case *bool:
		retryTypeBool = *r
	}

	var err error

	try := 0
	for {
		select {
		case <-ctx.Done():
			chk.Thread.Cancel()
			return err
		default:
			if try > 0 {
				log.Debugf("interval %v", chk.Interval)
				time.Sleep(chk.Interval)
				log.Infof("retry: %v...", try)
			}
			log.Debugf("try: %v...", try+1)

			err = chk.RunThread()

			if err == nil {
				return nil
			} else if isRetryTypeInt && try >= retryTypeInt {
				log.Debugf("failed retry=%v", retryTypeInt)
				return err
			} else if !isRetryTypeInt && !retryTypeBool {
				log.Debugf("failed retry=%v", retryTypeBool)
				return err
			} else if chk.Timeout != 0 && time.Now().After(timeoutTime) {
				log.Debugf("failed timeout=%v", chk.Timeout)
				return err
			} else {
				try++
			}
		}
	}

	return err
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

	chk.RegisterVarsLoad()

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

	err := chk.RunThreadLoop()

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
	runReport := chk.GlobalRunCtx.RunReport

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

	mutableCmd := &middleware.MutableCmd{
		AppConfig:                  chk.AppConfig,
		Command:                    chk.Command,
		OriginalCommand:            originalCommand,
		RequiredFiles:              requiredFiles,
		RequiredFilesSrcProcessors: requiredFilesProcessors,
		Expect:                     chk.Expect,
		Runner:                     chk.Runner,
		Dir:                        chk.Dir,
		Closer:                     chk.Closer,
	}
	return mutableCmd
}

func (chk *Chk) BuildLauncher() error {

	appCfg := chk.AppConfig
	rootPath := GetRootPath(chk.App)
	treeDir := appCfg.TreeDirLauncher(chk.TreeKeyParts)

	var launcherFilename string
	if chk.IsPreRun {
		launcherFilename = "pre"
	} else {
		launcherFilename = "post"
	}
	launcherFilename += "_check.bash"

	launcherDir := filepath.Join("build", "launcher", treeDir)
	launcherFile := filepath.Join(launcherDir, launcherFilename)
	launcherDirAbs := filepath.Join(rootPath, launcherDir)
	launcherFileAbs := filepath.Join(rootPath, launcherFile)

	if err := os.MkdirAll(launcherDirAbs, os.ModePerm); err != nil {
		return err
	}

	launcherContent := "#!/usr/bin/env bash\n"
	launcherContent += "BASH_ENV=${BASH_ENV:-/etc/profile}\n"
	launcherContent += `[ -f "$BASH_ENV" ] && source $BASH_ENV` + "\n"
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

		middlewareVars := chk.RunVars.GetPluginVars("middleware", cfgMiddleware.Name, cfgMiddleware.Plugin.UseVars, cfgMiddleware.Vars)

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
			chk.RunVars.SetValueString(k, v)
		}
	}
}

func (chk *Chk) RunRunner() error {

	r := chk.Runner

	if r.Plugin == nil {
		r.Plugin = chk.App.GetRunner(r.Name)
	}

	runnerVars := chk.RunVars.GetPluginVars("runner", r.Name, r.Plugin.UseVars, r.Vars)

	if chk.ExecTimeout != nil {
		chk.Thread.SetTimeout(chk.ExecTimeout)
	}

	vars := chk.RunVars.GetAll()

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

	if !chk.PreflightRunnedOnce {
		if err := chk.BuildLauncher(); err != nil {
			return err
		}
		if err := chk.ApplyMiddlewares(); err != nil {
			return err
		}
		chk.PreflightRunnedOnce = true
	}

	if err := chk.RunRunner(); err != nil {
		return err
	}

	return nil

}
