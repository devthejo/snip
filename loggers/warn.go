package loggers

import "github.com/sirupsen/logrus"

type Warn struct {
	Entry *logrus.Entry
}

func (w *Warn) Write(b []byte) (int, error) {
	n := len(b)
	w.Entry.Warn(string(b))
	return n, nil
}
