package play

import (
	"time"

	"gitlab.com/youtopia.earth/ops/snip/config"
)

type App interface {
	GetConfig() *config.Config
	GetNow() time.Time
}
