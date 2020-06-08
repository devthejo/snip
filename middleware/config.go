package middleware

import "github.com/sirupsen/logrus"

type Config struct {
	MutableCmd *MutableCmd
	Logger     *logrus.Entry
}
