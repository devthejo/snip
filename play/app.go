package play

import (
	"time"

	"gitlab.com/youtopia.earth/ops/snip/config"
	"gitlab.com/youtopia.earth/ops/snip/middleware"
	"gitlab.com/youtopia.earth/ops/snip/proc"
)

type App interface {
	GetConfig() *config.Config
	GetNow() time.Time
	GetMainProc() *proc.Main
	GetMiddlewareApply(k string) middleware.Apply
}
