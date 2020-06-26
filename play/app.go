package play

import (
	"time"

	cache "github.com/patrickmn/go-cache"

	"gitlab.com/youtopia.earth/ops/snip/config"
	"gitlab.com/youtopia.earth/ops/snip/plugin/loader"
	"gitlab.com/youtopia.earth/ops/snip/plugin/middleware"
	"gitlab.com/youtopia.earth/ops/snip/plugin/runner"
	"gitlab.com/youtopia.earth/ops/snip/proc"
	"gitlab.com/youtopia.earth/ops/snip/registry"
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
}
