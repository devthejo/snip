package play

import (
	"time"

	"gitlab.com/youtopia.earth/ops/snip/config"
	"gitlab.com/youtopia.earth/ops/snip/plugin/loader"
	"gitlab.com/youtopia.earth/ops/snip/plugin/middleware"
	"gitlab.com/youtopia.earth/ops/snip/plugin/runner"
	"gitlab.com/youtopia.earth/ops/snip/proc"
)

type App interface {
	GetConfig() *config.Config
	GetNow() time.Time
	GetMainProc() *proc.Main
	GetLoader(string) *loader.Loader
	GetMiddleware(string) *middleware.Middleware
	GetRunner(string) *runner.Runner
}
