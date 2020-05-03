package play

import (
	"gitlab.com/youtopia.earth/ops/snip/config"
)

type App interface {
	GetConfig() *config.Config
}
