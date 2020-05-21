package app

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"gitlab.com/youtopia.earth/ops/snip/cmd"
	"gitlab.com/youtopia.earth/ops/snip/config"
)

type App struct {
	Config          *config.Config
	ConfigFile      *string
	ConfigEnvPrefix string
	ConfigLoader    *config.ConfigLoader
	Viper           *viper.Viper
	RootCmd         *cobra.Command
	Now             time.Time
}

func New() *App {
	app := NewApp()
	app.RunCmd()
	return app
}

func NewApp() *App {
	app := &App{}

	app.ConfigEnvPrefix = "SNIP"

	app.Now = time.Now()

	var configFile string
	app.ConfigFile = &configFile

	app.ConfigLoader = config.NewConfigLoader()
	app.ConfigLoader.SetEnvPrefix(app.ConfigEnvPrefix)
	app.ConfigLoader.SetFile(app.ConfigFile)
	app.Config = app.ConfigLoader.Config
	app.Viper = app.ConfigLoader.Viper

	return app
}

func (app *App) GetViper() *viper.Viper {
	return app.ConfigLoader.GetViper()
}

func (app *App) GetConfig() *config.Config {
	return app.ConfigLoader.GetConfig()
}

func (app *App) GetConfigLoader() *config.ConfigLoader {
	return app.ConfigLoader
}

func (app *App) GetConfigFile() *string {
	return app.ConfigFile
}

func (app *App) OnInitialize() {
	app.ConfigLoader.OnInitialize()
}

func (app *App) OnPreRun(cmd *cobra.Command) {
	app.ConfigLoader.OnPreRun(cmd)
}

func (app *App) RunCmd() {
	cobra.OnInitialize(app.OnInitialize)

	RootCmd := cmd.NewCmd(app)
	app.RootCmd = RootCmd
	app.ConfigLoader.RootCmd = RootCmd

	if err := RootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

func (app *App) GetNow() time.Time {
	return app.Now
}
