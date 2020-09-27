package play

import (
	"time"

	"github.com/logrusorgru/aurora"
	cache "github.com/patrickmn/go-cache"

	"gitlab.com/ytopia/ops/snip/config"
	"gitlab.com/ytopia/ops/snip/plugin/loader"
	"gitlab.com/ytopia/ops/snip/plugin/middleware"
	"gitlab.com/ytopia/ops/snip/plugin/runner"
	"gitlab.com/ytopia/ops/snip/proc"
	"gitlab.com/ytopia/ops/snip/registry"
)

type App interface {
	GetConfig() *config.Config
	GetNow() time.Time
	GetCache() *cache.Cache
	GetVarsRegistry() *registry.NsVars
	GetMainProc() *proc.Main
	GetLoader(string) *loader.Plugin
	GetMiddleware(string) *middleware.Plugin
	GetRunner(string) *runner.Plugin
	GetAurora() aurora.Aurora
	IsExiting() bool
	Exiting()
}
