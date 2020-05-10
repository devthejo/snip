package tools

import (
	"github.com/sirupsen/logrus"
)

// Formatter implements logrus.Formatter interface.
type LogrusFormatterMsgOnly struct {
}

func (f *LogrusFormatterMsgOnly) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte(entry.Message), nil
}
