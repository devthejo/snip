package proc

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/devthejo/snip/config"
)

type Main struct {
	App           App
	Context       context.Context
	ContextCancel context.CancelFunc
	MainChan      chan os.Signal
	WaitGroup     *sync.WaitGroup
	Ended         bool
	ExitCode      int
	Success       bool
}

func CreateMain(app App) *Main {
	proc := &Main{}

	proc.App = app

	ctx, cancel := context.WithCancel(context.Background())
	proc.Context = ctx
	proc.ContextCancel = cancel

	proc.WaitGroup = &sync.WaitGroup{}

	return proc
}

func (proc *Main) GetConfig() *config.Config {
	return proc.App.GetConfig()
}

func (proc *Main) GetWaitGroup() *sync.WaitGroup {
	return proc.WaitGroup
}

func (proc *Main) WaitShutdown() bool {
	app := proc.App
	cfg := app.GetConfig()
	app.Exiting()

	chForce := make(chan os.Signal)
	signal.Notify(chForce, syscall.SIGINT)

	c := make(chan struct{})
	go func() {
		defer close(c)
		proc.WaitGroup.Wait()
	}()
	select {
	case <-c:
		logrus.Info("workers exited gracefully")
		return true
	case <-time.After(cfg.ShutdownTimeout):
		logrus.Warn("workers killed by timeout")
		return false
	case <-chForce:
		logrus.Warn("workers killed by user")
		return false
	}

}

func (proc *Main) MainOpener() {
	proc.MainChan = make(chan os.Signal)
	signal.Notify(proc.MainChan, syscall.SIGINT, syscall.SIGTERM)
}
func (proc *Main) MainCloser() {
	<-proc.MainChan
	if proc.Ended {
		if proc.ExitCode == 0 {
			proc.Success = true
			logrus.Debug("success")
		} else {
			logrus.Error("failed")
		}
	} else {
		logrus.Info("shutdown signal received, cancelling workers...")
		proc.Cancel()
		proc.WaitShutdown()
		logrus.Info("shutting down")
	}
}

func (proc *Main) GetContext() context.Context {
	return proc.Context
}

func (proc *Main) Cancel() {
	proc.ContextCancel()
}

func (proc *Main) Done() <-chan struct{} {
	return proc.Context.Done()
}

func (proc *Main) Run(f func() error) {
	proc.RunMain(f)
	os.Exit(proc.ExitCode)
}

func (proc *Main) RunMain(f func() error) {

	proc.MainOpener()

	go func() {
		err := f()
		if err != nil {
			logrus.Error(err)
			proc.ExitCode = 1
		}
		proc.WaitGroup.Wait()
		proc.End()
	}()

	proc.MainCloser()
}

func (proc *Main) End() {
	proc.Ended = true
	close(proc.MainChan)
}

func (proc *Main) Exit(exitCode int) {
	proc.ExitCode = exitCode
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
}
