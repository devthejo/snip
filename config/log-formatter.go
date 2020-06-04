package config

import (
	"strings"

	"github.com/sirupsen/logrus"
)

type LogContextKey string

type LogFormatter struct {
	logrus.TextFormatter
	NativeTextFormatter *logrus.TextFormatter
}

func (f *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	b, err := f.NativeTextFormatter.Format(entry)
	if err != nil {
		return b, err
	}
	msg := string(b)

	var icon string
	var levelText string
	switch entry.Level {
	case logrus.DebugLevel:
		icon = "⏿"
		levelText = "DEBU"
	case logrus.TraceLevel:
		icon = "👁"
		levelText = "TRAC"
	case logrus.WarnLevel:
		icon = "🛆"
		levelText = "WARN"
	case logrus.ErrorLevel:
		icon = "⮾"
		levelText = "ERRO"
	case logrus.FatalLevel:
		icon = "⮾"
		levelText = "FATA"
	case logrus.PanicLevel:
		icon = "☠"
		levelText = "PANI"
	default:
		levelText = "INFO"
		icon = "🛈"
	}

	var indentationInt int
	if entry.Context != nil {
		indentationKey := LogContextKey("indentation")
		if v := entry.Context.Value(indentationKey); v != nil {
			indentationInt = v.(int)
		}
	}

	indentationString := strings.Repeat("  ", indentationInt)
	msg = strings.Replace(msg, levelText, icon+" "+indentationString, 1)

	return []byte(msg), err
}
