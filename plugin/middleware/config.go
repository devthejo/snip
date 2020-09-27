package middleware

import (
	"context"

	"github.com/sirupsen/logrus"

	snipplugin "gitlab.com/ytopia/ops/snip/plugin"
)

type Config struct {
	AppConfig *snipplugin.AppConfig

	MiddlewareVars map[string]string

	Context       context.Context
	ContextCancel context.CancelFunc
	MutableCmd    *MutableCmd
	Logger        *logrus.Entry
}
