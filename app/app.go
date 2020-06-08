package app

import (
	"plugin"
	"time"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"gitlab.com/youtopia.earth/ops/snip/cmd"
	"gitlab.com/youtopia.earth/ops/snip/config"
	"gitlab.com/youtopia.earth/ops/snip/errors"
	"gitlab.com/youtopia.earth/ops/snip/middleware"
	"gitlab.com/youtopia.earth/ops/snip/proc"
)

type App struct {
	Config          *config.Config
	ConfigFile      *string
	ConfigEnvPrefix string
	ConfigLoader    *config.ConfigLoader
	Viper           *viper.Viper
	RootCmd         *cobra.Command

	Now      time.Time
	MainProc *proc.Main

	Middlewares cmap.ConcurrentMap
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

	app.Middlewares = cmap.New()

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

func (app *App) GetMainProc() *proc.Main {
	if app.MainProc == nil {
		app.MainProc = proc.CreateMain(app)
	}
	return app.MainProc
}

func (app *App) GetMiddlewarePlugin(k string) *plugin.Plugin {
	plugInterface, ok := app.Middlewares.Get(k)
	var plug *plugin.Plugin
	if ok {
		plug = plugInterface.(*plugin.Plugin)
	} else {
		mod := "./middlewares/" + k + ".so"
		var err error
		plug, err = plugin.Open(mod)
		errors.Check(err)
		app.Middlewares.Set(k, plug)
	}
	return plug
}
func (app *App) GetMiddleware(k string) middleware.Func {
	plug := app.GetMiddlewarePlugin(k)
	symRun, err := plug.Lookup("Middleware")
	errors.Check(err)
	// run, ok := symRun.(middleware.Func)
	run, ok := symRun.(func(*middleware.Config, func() error) error)
	if !ok {
		logrus.Fatalf("unexpected type from module symbol on middleware %v", k)
	}
	return run
}
