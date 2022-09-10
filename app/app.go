package app

import (
	"os"
	"os/user"
	"path/filepath"
	"plugin"
	"time"

	auroraPackage "github.com/logrusorgru/aurora"
	"github.com/mattn/go-isatty"
	cmap "github.com/orcaman/concurrent-map"
	cache "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/devthejo/snip/cmd"
	"github.com/devthejo/snip/config"
	"github.com/devthejo/snip/errors"
	"github.com/devthejo/snip/plugin/loader"
	"github.com/devthejo/snip/plugin/middleware"
	"github.com/devthejo/snip/plugin/runner"
	"github.com/devthejo/snip/proc"
	"github.com/devthejo/snip/registry"

	pluginLoaderMarkdown "github.com/devthejo/snip/plugins-native/loaders/markdown"
	pluginMiddlewareSu "github.com/devthejo/snip/plugins-native/middlewares/su"
	pluginMiddlewareSudo "github.com/devthejo/snip/plugins-native/middlewares/sudo"
	pluginRunnerSH "github.com/devthejo/snip/plugins-native/runners/sh"
	pluginRunnerSSH "github.com/devthejo/snip/plugins-native/runners/ssh"
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

	Plugins cmap.ConcurrentMap

	Cache *cache.Cache

	VarsRegistry *registry.NsVars

	ExitingState bool

	Aurora auroraPackage.Aurora

	Version string
}

func New(version string) *App {
	app := NewApp(version)
	app.RunCmd()
	return app
}

func NewApp(version string) *App {
	app := &App{
		Version: version,
	}

	app.ConfigEnvPrefix = "SNIP"

	app.Now = time.Now()

	app.Cache = cache.New(5*time.Minute, 10*time.Minute)

	var configFile string
	app.ConfigFile = &configFile

	app.ConfigLoader = config.NewConfigLoader()
	app.ConfigLoader.SetEnvPrefix(app.ConfigEnvPrefix)
	app.ConfigLoader.SetFile(app.ConfigFile)
	app.Config = app.ConfigLoader.Config
	app.Viper = app.ConfigLoader.Viper

	app.Plugins = cmap.New()
	app.LoadNativePlugins()

	return app
}

func (app *App) GetViper() *viper.Viper {
	return app.ConfigLoader.GetViper()
}

func (app *App) GetConfig() *config.Config {
	return app.ConfigLoader.GetConfig()
}

func (app *App) GetVersion() string {
	return app.Version
}

func (app *App) GetConfigLoader() *config.ConfigLoader {
	return app.ConfigLoader
}

func (app *App) GetConfigFile() *string {
	return app.ConfigFile
}

func (app *App) OnInitialize() {
	app.ConfigLoader.OnInitialize()
	app.InitAurora()
}

func (app *App) InitAurora() {
	var enableColors bool
	if app.Config.LogForceColors {
		enableColors = true
	} else {
		enableColors = isatty.IsTerminal(os.Stdout.Fd())
	}
	app.Aurora = auroraPackage.NewAurora(enableColors)
}

func (app *App) GetAurora() auroraPackage.Aurora {
	return app.Aurora
}

func (app *App) OnPreRun(cmd *cobra.Command) {
	app.ConfigLoader.OnPreRun(cmd)

	usr, _ := user.Current()
	app.VarsRegistry = registry.CreateNsVars(&registry.NsVarsOptions{
		BasePath: filepath.Join(usr.HomeDir, ".snip", app.Config.DeploymentName, "vars_persist"),
	})
}

func (app *App) IsExiting() bool {
	return app.ExitingState
}
func (app *App) Exiting() {
	app.ExitingState = true
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

func (app *App) GetCache() *cache.Cache {
	return app.Cache
}

func (app *App) GetVarsRegistry() *registry.NsVars {
	return app.VarsRegistry
}

func (app *App) GetMainProc() *proc.Main {
	if app.MainProc == nil {
		app.MainProc = proc.CreateMain(app)
	}
	return app.MainProc
}

func (app *App) GetPlugin(k string) interface{} {
	plugInterface, ok := app.Plugins.Get(k)
	var plug interface{}
	if ok {
		plug = plugInterface
	} else {
		mod := "./plugins/" + k + ".so"
		var err error
		plug, err = plugin.Open(mod)
		errors.Check(err)
		app.Plugins.Set(k, plug)
	}
	return plug
}

func (app *App) GetLoader(k string) *loader.Plugin {
	plug := app.GetPlugin("loaders/" + k)
	var run *loader.Plugin
	switch v := plug.(type) {
	case *plugin.Plugin:
		sym, err := v.Lookup("Loader")
		errors.Check(err)
		var ok bool
		run, ok = sym.(*loader.Plugin)
		if !ok {
			logrus.Fatalf("unexpected type from module symbol on loader plugin %s: %T", k, sym)
		}
	case *loader.Plugin:
		run = v
	}
	return run
}

func (app *App) GetMiddleware(k string) *middleware.Plugin {
	plug := app.GetPlugin("middlewares/" + k)
	var run *middleware.Plugin
	switch v := plug.(type) {
	case *plugin.Plugin:
		sym, err := v.Lookup("Middleware")
		errors.Check(err)
		var ok bool
		run, ok = sym.(*middleware.Plugin)
		if !ok {
			logrus.Fatalf("unexpected type from module symbol on middleware plugin %s: %T", k, sym)
		}
	case *middleware.Plugin:
		run = v
	}
	return run
}

func (app *App) GetRunner(k string) *runner.Plugin {
	plug := app.GetPlugin("runners/" + k)
	var run *runner.Plugin
	switch v := plug.(type) {
	case *plugin.Plugin:
		sym, err := v.Lookup("Runner")
		errors.Check(err)
		var ok bool
		run, ok = sym.(*runner.Plugin)
		if !ok {
			logrus.Fatalf("unexpected type from module symbol on runner plugin %s: %T", k, sym)
		}
	case *runner.Plugin:
		run = v
	}
	return run
}

func (app *App) LoadNativePlugins() {
	app.Plugins.Set("loaders/markdown", &pluginLoaderMarkdown.Loader)
	app.Plugins.Set("middlewares/sudo", &pluginMiddlewareSudo.Middleware)
	app.Plugins.Set("middlewares/su", &pluginMiddlewareSu.Middleware)
	app.Plugins.Set("runners/sh", &pluginRunnerSH.Runner)
	app.Plugins.Set("runners/ssh", &pluginRunnerSSH.Runner)
}
