package play

import (
	"time"

	"github.com/logrusorgru/aurora"
	cache "github.com/patrickmn/go-cache"

	"github.com/devthejo/snip/config"
	"github.com/devthejo/snip/plugin/loader"
	"github.com/devthejo/snip/plugin/middleware"
	"github.com/devthejo/snip/plugin/runner"
	"github.com/devthejo/snip/proc"
	"github.com/devthejo/snip/registry"
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
