package middleware

import (
	"context"

	"github.com/sirupsen/logrus"
)

type Config struct {
	Context       context.Context
	ContextCancel context.CancelFunc
	MutableCmd    *MutableCmd
	Logger        *logrus.Entry
}
