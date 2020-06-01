package config

import (
	"strings"

	"github.com/sirupsen/logrus"
)

func init() {
	ConfigureLogrusLogType("JSON", false)
}

var currentLogType *string
var currentLogForceColors *bool

func ConfigureLogrusLogType(strLogType string, logForceColors bool) {

	upLogType := strings.ToUpper(strLogType)
	if upLogType != "" &&
		((currentLogType == nil || upLogType != *currentLogType) ||
			(currentLogForceColors == nil || logForceColors != *currentLogForceColors)) {
		currentLogType = &upLogType
		currentLogForceColors = &logForceColors
		switch upLogType {
		case "JSON":
			logrus.SetFormatter(&logrus.JSONFormatter{})
		case "TEXT":
			// logrus.SetFormatter(&logrus.TextFormatter{
			// 	ForceColors:      true,
			// 	DisableTimestamp: true,
			// })
			logrus.SetFormatter(&LogFormatter{
				NativeTextFormatter: &logrus.TextFormatter{
					ForceColors:      true,
					DisableTimestamp: true,
				},
			})
		default:
			logrus.Fatalf(`Invalid LOG_TYPE: "%v"`, upLogType)
		}
	}
}

var currentLogLevel *string

func ConfigureLogrusLogLevel(strLogLevel string) {

	upLogLevel := strings.ToUpper(strLogLevel)

	if upLogLevel != "" && (currentLogLevel == nil || upLogLevel != *currentLogLevel) {
		currentLogLevel = &upLogLevel
		logLevel, err := logrus.ParseLevel(upLogLevel)
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.SetLevel(logLevel)
	}

}
