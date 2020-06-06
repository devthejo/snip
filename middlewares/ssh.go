package main

import (
	"strings"

	"gitlab.com/youtopia.earth/ops/snip/middleware"
	"gitlab.com/youtopia.earth/ops/snip/middlewares/ssh"
)

func Middleware(mutableCmd *middleware.MutableCmd, next func() error) error {

	commandBin := mutableCmd.OriginalCommand

	cfg := ssh.CreateConfig(mutableCmd.Vars)

	if strings.Contains(commandBin, "/") {
		err := ssh.Upload(cfg, commandBin)
		if err != nil {
			return err
		}
	}

	err := ssh.Exec(cfg, mutableCmd)
	if err != nil {
		return err
	}

	return nil
}
