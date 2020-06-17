package runner

import (
	"context"
	"io"

	expect "github.com/google/goexpect"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Context       *context.Context
	ContextCancel *context.CancelFunc
	Logger        *logrus.Entry

	Command []string
	Vars    map[string]string

	RequiredFiles map[string]string
	Expect        []expect.Batcher
	Stdin         io.Reader
}

func (cfg *Config) EnvMap() map[string]string {
	m := make(map[string]string)
	for k, v := range cfg.Vars {
		if k[0:1] != "@" {
			m[k] = v
		}
	}
	return m
}
