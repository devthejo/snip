package ssh

import (
	"log"
	"strings"

	shellquote "github.com/kballard/go-shellquote"
	"github.com/kvz/logstreamer"
	"github.com/sirupsen/logrus"

	"gitlab.com/youtopia.earth/ops/snip/middleware"
	"gitlab.com/youtopia.earth/ops/snip/sshclient"
)

func Exec(cfg *sshclient.Config, mutableCmd *middleware.MutableCmd, logger *logrus.Entry) error {
	client, err := sshclient.CreateClient(cfg)
	if err != nil {
		return err
	}

	if err := client.Connect(); err != nil {
		return err
	}
	defer client.Close()

	commandSlice := append([]string{}, mutableCmd.Args...)
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
	for k, v := range mutableCmd.Vars {
		setenvSlice = append(setenvSlice, k+"="+v)
	}
	setenv = shellquote.Join(setenvSlice...)

	runCmdSlice := []string{"cd", cwd, "&&", setenv, command}
	runCmd := strings.Join(runCmdSlice, " ")

	logger.Debugf("remote command: %v", runCmd)

	if err := session.Start(runCmd); err != nil {
		return err
	}

	if err := session.Wait(); err != nil {
		return err
	}

	return nil
}
