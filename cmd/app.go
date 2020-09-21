package cmd

import (
	"time"

	cache "github.com/patrickmn/go-cache"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/logrusorgru/aurora"

	"gitlab.com/youtopia.earth/ops/snip/config"
	"gitlab.com/youtopia.earth/ops/snip/plugin/loader"
	"gitlab.com/youtopia.earth/ops/snip/plugin/middleware"
	"gitlab.com/youtopia.earth/ops/snip/plugin/runner"
	"gitlab.com/youtopia.earth/ops/snip/proc"
	"gitlab.com/youtopia.earth/ops/snip/registry"
)

type App interface {
	GetConfig() *config.Config
	GetViper() *viper.Viper
	GetConfigLoader() *config.ConfigLoader
	GetConfigFile() *string
	GetAurora() aurora.Aurora
	OnPreRun(*cobra.Command)
	GetNow() time.Time
	GetCache() *cache.Cache
	GetVarsRegistry() *registry.NsVars
	GetMainProc() *proc.Main
	GetLoader(string) *loader.Plugin
	GetMiddleware(string) *middleware.Plugin
	GetRunner(string) *runner.Plugin
	IsExiting() bool
	Exiting()
}
