package tools

import (
	"runtime"

	"github.com/sirupsen/logrus"
)

// PrintMemUsage outputs the current, total and OS memory being used. As well as the number
// of garage collection cycles completed.
func PrintMemUsage(loggers ...*logrus.Entry) {

	if len(loggers) == 0 {
		logger := logrus.WithFields(logrus.Fields{})
		loggers = append(loggers, logger)
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	for _, logger := range loggers {
		logger.Debugf("Alloc = %v MiB", bToMb(m.Alloc))
		logger.Debugf("TotalAlloc = %v MiB", bToMb(m.TotalAlloc))
		logger.Debugf("Sys = %v MiB", bToMb(m.Sys))
		logger.Debugf("NumGC = %v\n", m.NumGC)
	}
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
