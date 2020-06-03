package proc

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"

	"gitlab.com/youtopia.earth/ops/snip/config"
	"gitlab.com/youtopia.earth/ops/snip/tools"
)

type Main struct {
	App           App
	Context       *context.Context
	ContextCancel *context.CancelFunc
	MainChan      chan os.Signal
	WaitGroup     *sync.WaitGroup
	ExitCode      int
	Ended         bool
}

func CreateMain(app App) *Main {
	proc := &Main{}

	proc.App = app

	ctx, cancel := context.WithCancel(context.Background())
	proc.Context = &ctx
	proc.ContextCancel = &cancel

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
	cfg := proc.App.GetConfig()
	return tools.WaitTimeout(proc.WaitGroup, cfg.ShutdownTimeout)
}

func (proc *Main) MainOpener() {
	proc.MainChan = make(chan os.Signal)
	signal.Notify(proc.MainChan, syscall.SIGINT, syscall.SIGTERM)
}
func (proc *Main) MainCloser() {
	<-proc.MainChan
	if proc.Ended {
		logrus.Info("done")
	} else {
		logrus.Info("shutdown signal received")
		proc.Cancel()

		if proc.WaitShutdown() {
			logrus.Info("workers done, shutting down")
		} else {
			logrus.Warn("workers timeout, shutting down")
		}
	}
}

func (proc *Main) GetContext() *context.Context {
	return proc.Context
}

func (proc *Main) Cancel() {
	(*proc.ContextCancel)()
}

func (proc *Main) Done() <-chan struct{} {
	return (*proc.Context).Done()
}

func (proc *Main) Run(f func() error) {
	if err := proc.RunMain(f); err != nil {
		proc.ExitCode = 1
	}
	os.Exit(proc.ExitCode)
}

func (proc *Main) RunMain(f func() error) error {

	proc.MainOpener()

	if err := f(); err != nil {
		return err
	}

	go func() {
		proc.WaitGroup.Wait()
		proc.End()
	}()

	proc.MainCloser()

	return nil
}

func (proc *Main) End() {
	proc.Ended = true
	close(proc.MainChan)
}

func (proc *Main) Exit(exitCode int) {
	proc.ExitCode = exitCode
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
}
