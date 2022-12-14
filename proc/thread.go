package proc

import (
	"context"
	"os/exec"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/devthejo/snip/errors"
)

type Thread struct {
	App App

	ExecRunning  bool
	ExecExited   bool
	ExecExitCode int

	ExecTimeout *time.Duration

	Context       context.Context
	ContextCancel context.CancelFunc
	MainProc      *Main

	Logger *logrus.Entry

	Error error
}

func CreateThread(app App) *Thread {
	thr := &Thread{}
	thr.App = app
	thr.MainProc = app.GetMainProc()
	thr.initContext()
	return thr
}

func (thr *Thread) Run(runMain func() error) error {

	mainProc := thr.App.GetMainProc()

	go func() {
		<-mainProc.Done()
		thr.Cancel()
	}()

	thr.Exec(runMain)

	return thr.Error
}
func (c *Thread) Cancel() {
	c.ContextCancel()
}
func (c *Thread) Done() <-chan struct{} {
	return c.Context.Done()
}

func (thr *Thread) SetTimeout(timeout *time.Duration) {
	thr.ExecTimeout = timeout
	thr.setContextWithTimeout()
}

func (thr *Thread) initContext() {
	ctx, cancel := context.WithCancel(context.Background())
	thr.Context = ctx
	thr.ContextCancel = cancel
}

func (thr *Thread) setContextWithTimeout() {
	if thr.ExecTimeout == nil {
		return
	}
	ctx, cancel := context.WithTimeout(thr.Context, *thr.ExecTimeout)
	thr.Context = ctx
	thr.ContextCancel = cancel
}

func (thr *Thread) LogErrors() {
	if thr.Context.Err() == context.DeadlineExceeded {
		thr.Logger.WithFields(logrus.Fields{
			"timeout": thr.ExecTimeout,
		}).Warnf("thread exec timeout fail")
	} else if thr.ExecExitCode > 0 {
		switch thr.ExecExitCode {
		case 129, 141:
			thr.Logger.WithFields(logrus.Fields{
				"exitCode": thr.ExecExitCode,
			}).Warnf("thread exec error: %v", thr.Error)
		default:
			thr.Logger.WithFields(logrus.Fields{
				"exitCode": thr.ExecExitCode,
			}).Errorf("thread exec error: %v", thr.Error)
		}
	}
}

func (thr *Thread) Reset() {
	thr.ExecRunning = false
	thr.ExecExited = false
	thr.Error = nil
	thr.ExecExitCode = 0
	thr.initContext()
	thr.setContextWithTimeout()
}

func (thr *Thread) Exec(runMain func() error) {

	app := thr.App

	if app.IsExiting() {
		return
	}

	thr.ExecRunning = true

	mainWg := app.GetMainProc().WaitGroup
	mainWg.Add(1)
	defer mainWg.Done()

	err := runMain()

	if err != nil {
		_, err = errors.CreateErrorWithCodeFromError(err)

		thr.Error = err

		if exitError, ok := err.(*exec.ExitError); ok {
			thr.ExecExitCode = exitError.ExitCode()
		} else if exitError, ok := err.(*errors.ErrorWithCode); ok {
			thr.ExecExitCode = exitError.Code
		} else {
			thr.ExecExitCode = 1
		}

	}

	thr.ExecRunning = false
	thr.ExecExited = true

	thr.Cancel()
}
