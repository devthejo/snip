package runner

import (
	"context"

	cache "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"

	expect "gitlab.com/youtopia.earth/ops/snip/goexpect"
	snipplugin "gitlab.com/youtopia.earth/ops/snip/plugin"
)

type Config struct {
	AppConfig *snipplugin.AppConfig

	Context       context.Context
	ContextCancel context.CancelFunc
	Logger        *logrus.Entry

	Cache *cache.Cache

	Command []string
	Vars    map[string]string

	RequiredFiles map[string]string
	Expect        []expect.Batcher

	Dir string

	Closer *func(interface{}) bool
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
