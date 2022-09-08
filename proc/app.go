package proc

import (
	"github.com/devthejo/snip/config"
)

type App interface {
	GetConfig() *config.Config
	GetMainProc() *Main
	IsExiting() bool
	Exiting()
}
