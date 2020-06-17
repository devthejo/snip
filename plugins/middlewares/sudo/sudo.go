package main

import (
	"strings"

	"gitlab.com/youtopia.earth/ops/snip/plugin/middleware"
)

var (
	Middleware = middleware.Middleware{
		Apply: func(cfg *middleware.Config) (bool, error) {

			mutableCmd := cfg.MutableCmd

			command := []string{"sudo"}

			if pass, ok := mutableCmd.Vars["@SUDO_PASS"]; ok {
				mutableCmd.Stdin = strings.NewReader(pass + "\n")
				command = append(command, "--stdin", "--preserve-env")
			}

			mutableCmd.Command = append(command, mutableCmd.Command...)
			return true, nil
		},
	}
)
