package cmd

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"gitlab.com/youtopia.earth/ops/snip/config"
	"gitlab.com/youtopia.earth/ops/snip/plugin/loader"
	"gitlab.com/youtopia.earth/ops/snip/plugin/middleware"
	"gitlab.com/youtopia.earth/ops/snip/plugin/runner"
	"gitlab.com/youtopia.earth/ops/snip/proc"
)

type App interface {
	GetConfig() *config.Config
	GetViper() *viper.Viper
	GetConfigLoader() *config.ConfigLoader
	GetConfigFile() *string
	OnPreRun(*cobra.Command)
	GetNow() time.Time
	GetMainProc() *proc.Main
	GetLoader(string) *loader.Loader
	GetMiddleware(string) *middleware.Middleware
	GetRunner(string) *runner.Runner
}
