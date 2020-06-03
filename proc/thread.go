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

	Context       *context.Context
	ContextCancel *context.CancelFunc
	WaitGroup     *sync.WaitGroup
	MainProc      *Main

	Logger *logrus.Entry

	Vars map[string]string

	Error error

	ThreadRunMain func(ctx context.Context, hookFunc func(c *exec.Cmd) error) error
}

func (thr *Thread) ThreadRun() error {
	mainProc := thr.App.GetMainProc()

	go func() {
		<-mainProc.Done()
		thr.Cancel()
	}()

	go thr.ThreadExec()
	thr.WaitGroup.Wait()

	return thr.Error
}
func (c *Thread) Cancel() {
	(*c.ContextCancel)()
}
func (c *Thread) Done() <-chan struct{} {
	return (*c.Context).Done()
}

func (thr *Thread) ThreadExec() {

	execTimeout := thr.ExecTimeout
	var ctx context.Context
	if execTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), execTimeout)
		defer cancel()
	} else {
		ctx = context.Background()
	}

	thr.ExecRunning = true

	thr.WaitGroup.Add(1)

	ctx, cancel := context.WithCancel(context.Background())
	hookFunc := func(c *exec.Cmd) error {
		go func(c *exec.Cmd) {
			select {
			case <-thr.Done():
				if c.Process != nil {
					thr.Logger.Debug(`sending stopsignal`)
					c.Process.Signal(syscall.SIGTERM)
				}
			case <-ctx.Done():
				return
			}
		}(c)
		return nil
	}

	mainWg := thr.App.GetMainProc().WaitGroup
	mainWg.Add(1)
	defer mainWg.Done()

	err := thr.ThreadRunMain(ctx, hookFunc)

	cancel()
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

		if ctx.Err() == context.DeadlineExceeded {
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

func (thr *Thread) ThreadRunCmd(commandSlice []string, args ...interface{}) error {
	commandSlice = thr.ExpandCmdEnv(commandSlice)
	if thr.ExecUser != nil {
		args = append(args, thr.ExecUser)
	}
	return tools.RunCmd(commandSlice, args...)
}
