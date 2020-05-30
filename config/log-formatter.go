package config

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type LogFormatter struct {
	logrus.TextFormatter
}

func (f *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// this whole mess of dealing with ansi color codes is required if you want the colored output otherwise you will lose colors in the log levels
	var levelColor int
	var icon string
	switch entry.Level {
	case logrus.DebugLevel, logrus.TraceLevel:
		levelColor = 31 // gray
		// icon = "☢"
		icon = "🐝"
	case logrus.WarnLevel:
		levelColor = 33 // yellow
		// icon = "⚠"
		icon = "🛆"
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = 31 // red
		// icon = "☠"
		icon = "⮾"
	default:
		levelColor = 36 // blue
		icon = "🛈"
	}
	return []byte(fmt.Sprintf("\x1b[%dm%s \x1b[0m %s\n", levelColor, icon, entry.Message)), nil
}
