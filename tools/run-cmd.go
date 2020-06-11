package tools

import (
	"context"
	"log"
	"os/exec"
	"syscall"

	"github.com/kvz/logstreamer"
	"github.com/sirupsen/logrus"
)

func RunCmd(commandSlice []string, args ...interface{}) error {

	var ctx context.Context
	var hookFuncs []func(*exec.Cmd) error
	var logger *logrus.Entry
	for _, argif := range args {
		switch arg := argif.(type) {
		case *logrus.Entry:
			logger = arg
		case context.Context:
			ctx = arg
		case func(cmd *exec.Cmd) error:
			hookFuncs = append(hookFuncs, arg)
		default:
			logrus.Fatalf(`invalid argument for tools.RunCmd type:"%T",value:"%v"`, argif, argif)
		}
	}
	if ctx == nil {
		ctx = context.Background()
	}

	var commandPath string
	var argsSlice []string

	commandPath, lookErr := exec.LookPath(commandSlice[0])
	if lookErr != nil {
		return lookErr
	}

	if len(commandSlice) > 1 {
		argsSlice = commandSlice[1:]
	}

	cmd := exec.CommandContext(ctx, commandPath, argsSlice...)

	if logger == nil {
		logger = logrus.WithFields(logrus.Fields{})
	}

	w := logger.Writer()
	defer w.Close()

	loggerOut := log.New(w, "", 0)
	logStreamer := logstreamer.NewLogstreamer(loggerOut, "", true)
	defer logStreamer.Close()
	cmd.Stdout = logStreamer
	cmd.Stderr = logStreamer
	logStreamer.FlushRecord()

	// Create a dedicated pidgroup used to forward signals to main process and all children
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	for _, hookFunc := range hookFuncs {
		if err := hookFunc(cmd); err != nil {
			return err
		}
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}
