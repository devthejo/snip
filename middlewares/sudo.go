package main

import (
	"gitlab.com/youtopia.earth/ops/snip/middleware"
)

func Middleware(mutableCmd *middleware.MutableCmd, next func() error) error {
	mutableCmd.Args = append([]string{mutableCmd.Command}, mutableCmd.Args...)
	mutableCmd.Command = "sudo"
	return next()
}
