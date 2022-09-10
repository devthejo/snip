package cmd

import (
	"time"

	"github.com/logrusorgru/aurora"
	cache "github.com/patrickmn/go-cache"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/devthejo/snip/config"
	"github.com/devthejo/snip/plugin/loader"
	"github.com/devthejo/snip/plugin/middleware"
	"github.com/devthejo/snip/plugin/runner"
	"github.com/devthejo/snip/proc"
	"github.com/devthejo/snip/registry"
)

type App interface {
	GetConfig() *config.Config
	GetViper() *viper.Viper
	GetConfigLoader() *config.ConfigLoader
	GetConfigFile() *string
	GetAurora() aurora.Aurora
	GetVersion() string
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
