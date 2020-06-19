package middleware

import (
	"context"

	"github.com/sirupsen/logrus"

	snipplugin "gitlab.com/youtopia.earth/ops/snip/plugin"
)

type Config struct {
	Context       context.Context
	ContextCancel context.CancelFunc
	MutableCmd    *snipplugin.MutableCmd
	Logger        *logrus.Entry
}
