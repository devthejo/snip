package proc

import (
	"gitlab.com/ytopia/ops/snip/config"
)

type App interface {
	GetConfig() *config.Config
	GetMainProc() *Main
	IsExiting() bool
	Exiting()
}
