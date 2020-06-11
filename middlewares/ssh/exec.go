package ssh

import (
	"log"
	"strings"

	shellquote "github.com/kballard/go-shellquote"
	"github.com/kvz/logstreamer"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"

	"gitlab.com/youtopia.earth/ops/snip/errors"
	"gitlab.com/youtopia.earth/ops/snip/middleware"
	"gitlab.com/youtopia.earth/ops/snip/sshclient"
)

func Exec(cfg *sshclient.Config, middlewareConfig *middleware.Config) error {
	mutableCmd := middlewareConfig.MutableCmd
	logger := middlewareConfig.Logger
	client, err := sshclient.CreateClient(cfg)
	if err != nil {
		return err
	}

	if err := client.Connect(); err != nil {
		return err
	}
	defer client.Close()

	commandSlice := append([]string{mutableCmd.Command}, mutableCmd.Args...)
	command := shellquote.Join(commandSlice...)

	cwd := GetSnipPath(cfg.User)

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	loggerSSH := logger.WithFields(logrus.Fields{
		"ssh":  true,
		"host": cfg.Host,
	})

	w := loggerSSH.Writer()
	defer w.Close()

	loggerOut := log.New(w, "", 0)
	logStreamer := logstreamer.NewLogstreamer(loggerOut, "", true)
	defer logStreamer.Close()
	session.Stdout = logStreamer
	session.Stderr = logStreamer
	logStreamer.FlushRecord()

	var setenvSlice []string
	setenv := ""
	for k, v := range mutableCmd.EnvMap() {
		setenvSlice = append(setenvSlice, k+"="+v)
	}
	setenv = shellquote.Join(setenvSlice...)

	runCmdSlice := []string{"cd", cwd, "&&", setenv, command}
	runCmd := strings.Join(runCmdSlice, " ")

	logger.Debugf("remote command: %v", runCmd)

	go func() {
		select {
		case <-middlewareConfig.Context.Done():
			logger.Debug(`sending stopsignal`)
			session.Signal(ssh.SIGTERM)
			session.Close()
			return
		}
	}()

	if err := session.Start(runCmd); err != nil {
		return err
	}

	if err := session.Wait(); err != nil {
		if err.Error() == "Process exited with status 141 from signal PIPE" {
			return &errors.ErrorWithCode{
				Err:  err,
				Code: 141,
			}
		}
		return err
	}

	return nil
}
