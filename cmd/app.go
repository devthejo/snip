package cmd

import (
	"time"

	cache "github.com/patrickmn/go-cache"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/logrusorgru/aurora"

	"gitlab.com/ytopia/ops/snip/config"
	"gitlab.com/ytopia/ops/snip/plugin/loader"
	"gitlab.com/ytopia/ops/snip/plugin/middleware"
	"gitlab.com/ytopia/ops/snip/plugin/runner"
	"gitlab.com/ytopia/ops/snip/proc"
	"gitlab.com/ytopia/ops/snip/registry"
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
