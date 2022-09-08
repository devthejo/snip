package cmd

import (
	"os"

	"github.com/devthejo/snip/config"
	"github.com/spf13/cobra"
)

func NewCmd(app App) *cobra.Command {
	cmd := CmdRoot(app)

	cmd.AddCommand(CmdCompletion(app, cmd))

	cmd.AddCommand(CmdPlay(app))
	// cmd.AddCommand(CmdMarkdown(app))

	return cmd
}

func CmdRoot(app App) *cobra.Command {
	cl := app.GetConfigLoader()

	cmd := &cobra.Command{
		Use:                    "snip",
		Short:                  "Bash superset for DevOps 🚀",
		BashCompletionFunction: newBashCompletionFunc(cl),
	}

	configFile := app.GetConfigFile()

	pFlags := cmd.PersistentFlags()

	pFlags.StringVarP(configFile, "config", "", os.Getenv(cl.PrefixEnv("CONFIG")), config.FlagConfigDesc)
	pFlags.StringP("log-level", "l", config.FlagLogLevelDefault, config.FlagLogLevelDesc)
	pFlags.StringP("log-type", "", config.FlagLogTypeDefault, config.FlagLogTypeDesc)
	pFlags.BoolP("log-force-colors", "", config.FlagLogForceColorsDefault, config.FlagLogForceColorsDesc)
	pFlags.StringP("cwd", "", "", config.FlagCWDDesc)

	v := app.GetViper()

	v.BindPFlag("CONFIG", pFlags.Lookup("config"))
	v.BindPFlag("LOG_LEVEL", pFlags.Lookup("log-level"))
	v.BindPFlag("LOG_TYPE", pFlags.Lookup("log-type"))
	v.BindPFlag("LOG_FORCE_COLORS", pFlags.Lookup("log-force-colors"))

	v.BindEnv("CONFIG")
	v.BindEnv("LOG_LEVEL")
	v.BindEnv("LOG_TYPE")
	v.BindEnv("LOG_FORCE_COLORS")

	return cmd
}
