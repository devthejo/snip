package runner

import (
	"context"

	cache "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"

	expect "github.com/devthejo/snip/goexpect"
	snipplugin "github.com/devthejo/snip/plugin"
	"github.com/devthejo/snip/registry"
)

type Config struct {
	AppConfig *snipplugin.AppConfig

	RunnerVars map[string]string

	Context       context.Context
	ContextCancel context.CancelFunc
	Logger        *logrus.Entry

	Cache        *cache.Cache
	VarsRegistry *registry.NsVars

	Command      []string
	Vars         map[string]string
	RegisterVars map[string]*registry.VarDef
	Quiet        bool

	TreeKeyParts []string

	RequiredFiles map[string]string
	Use           map[string]string
	Persist       map[string]string
	TmpdirName    string
	Expect        []expect.Batcher

	Dir string

	Closer *func(interface{}, *string) bool
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
