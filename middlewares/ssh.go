package main

import (
	"strings"

	"gitlab.com/youtopia.earth/ops/snip/middleware"
	"gitlab.com/youtopia.earth/ops/snip/middlewares/ssh"
	"gitlab.com/youtopia.earth/ops/snip/sshclient"
)

func Middleware(middlewareConfig *middleware.Config, next func() error) error {

	mutableCmd := middlewareConfig.MutableCmd
	logger := middlewareConfig.Logger

	commandBin := mutableCmd.OriginalCommand

	cfg := sshclient.CreateConfig(mutableCmd.Vars)

	if strings.Contains(commandBin, "/") && !strings.HasPrefix(commandBin, "/") {
		err := ssh.Upload(cfg, commandBin, logger)
		if err != nil {
			return err
		}
	}

	err := ssh.Exec(cfg, middlewareConfig)
	if err != nil {
		return err
	}

	return nil
}
