package proc

import (
	"context"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/opencontainers/runc/libcontainer/user"
	"github.com/sirupsen/logrus"

	"gitlab.com/youtopia.earth/ops/snip/tools"
)

type Thread struct {
	App App

	ExecRunning  bool
	ExecExited   bool
	ExecExitCode int
	ExecUser     *user.ExecUser
	ExecTimeout  time.Duration

	Context       context.Context
	ContextCancel context.CancelFunc
	WaitGroup     *sync.WaitGroup
	MainProc      *Main

	CommandStopper func(*exec.Cmd) error

	Logger *logrus.Entry

	Vars map[string]string

	Error error
}

func CreateThread(app App) *Thread {
	thr := &Thread{}
	thr.WaitGroup = &sync.WaitGroup{}

	var ctx context.Context
	var cancel context.CancelFunc
	if thr.ExecTimeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), thr.ExecTimeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	thr.Context = ctx
	thr.ContextCancel = cancel

	thr.App = app
	thr.MainProc = app.GetMainProc()
	return thr
}

func (thr *Thread) Run(runMain func() error) error {

	mainProc := thr.App.GetMainProc()

	go func() {
		<-mainProc.Done()
		thr.Cancel()
	}()

	go thr.Exec(runMain)
	thr.WaitGroup.Wait()

	return thr.Error
}
func (c *Thread) Cancel() {
	c.ContextCancel()
}
func (c *Thread) Done() <-chan struct{} {
	return c.Context.Done()
}

func (thr *Thread) Exec(runMain func() error) {

	thr.ExecRunning = true

	thr.WaitGroup.Add(1)

	thr.CommandStopper = func(c *exec.Cmd) error {
		go func(c *exec.Cmd) {
			select {
			case <-thr.Done():
				if c.Process != nil {
					thr.Logger.Debug(`sending stopsignal`)
					c.Process.Signal(syscall.SIGTERM)
				}
				return
			}
		}(c)
		return nil
	}

	mainWg := thr.App.GetMainProc().WaitGroup
	mainWg.Add(1)
	defer mainWg.Done()

	err := runMain()

	thr.Cancel()
	thr.WaitGroup.Done()

	if err != nil {

		thr.ExecRunning = false
		thr.ExecExited = true

		if exitError, ok := err.(*exec.ExitError); ok {
			thr.ExecExitCode = exitError.ExitCode()
		} else {
			thr.ExecExitCode = 1
		}

		if thr.ExecExitCode > 0 {
			thr.Logger.WithFields(logrus.Fields{
				"exitCode": thr.ExecExitCode,
			}).Errorf("thread exec error: %v", err)
		}

		if thr.Context.Err() == context.DeadlineExceeded {
			thr.Logger.WithFields(logrus.Fields{
				"timeout": thr.ExecTimeout,
			}).Warnf("thread exec timeout fail")
		}

		thr.Cancel()

		thr.Error = err
	}

	thr.ExecRunning = false
	thr.ExecExited = true
}

func (thr *Thread) ExpandCmdEnvMapper(key string) string {
	if val, ok := thr.Vars[key]; ok {
		return val
	}
	return ""
}
func (thr *Thread) ExpandCmdEnv(commandSlice []string) []string {
	expandedCmd := make([]string, len(commandSlice))
	for i, str := range commandSlice {
		expandedCmd[i] = os.Expand(str, thr.ExpandCmdEnvMapper)
	}
	return expandedCmd
}

func (thr *Thread) RunCmd(commandSlice []string, args ...interface{}) error {
	commandSlice = thr.ExpandCmdEnv(commandSlice)
	if thr.ExecUser != nil {
		args = append(args, thr.ExecUser)
	}
	args = append(args, thr.Context)
	args = append(args, thr.CommandStopper)
	return tools.RunCmd(commandSlice, args...)
}