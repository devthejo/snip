package main

import (
	"gitlab.com/youtopia.earth/ops/snip/middleware"
)

func Middleware(middlewareConfig *middleware.Config, next func() error) error {

	mutableCmd := middlewareConfig.MutableCmd
	// logger := middlewareConfig.Logger

	mutableCmd.Args = append([]string{mutableCmd.Command}, mutableCmd.Args...)
	mutableCmd.Command = "sudo"
	return next()
}
