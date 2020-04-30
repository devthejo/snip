package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/youtopia.earth/ops/snip/config"
)

type App interface {
	GetConfig() *config.Config
	GetViper() *viper.Viper
	GetConfigLoader() *config.ConfigLoader
	GetConfigFile() *string
	OnPreRun(*cobra.Command)
}
