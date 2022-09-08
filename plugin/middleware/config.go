package middleware

import (
	"context"

	"github.com/sirupsen/logrus"

	snipplugin "github.com/devthejo/snip/plugin"
)

type Config struct {
	AppConfig *snipplugin.AppConfig

	MiddlewareVars map[string]string

	Context       context.Context
	ContextCancel context.CancelFunc
	MutableCmd    *MutableCmd
	Logger        *logrus.Entry
}
