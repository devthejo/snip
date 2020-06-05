package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/youtopia.earth/ops/snip/config"
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

	pFlags.String("ssh-host", config.FlagSSHHostDefault, config.FlagSSHHostDesc)
	pFlags.Int("ssh-port", config.FlagSSHPortDefault, config.FlagSSHPortDesc)
	pFlags.String("ssh-user", config.FlagSSHUserDefault, config.FlagSSHUserDesc)
	pFlags.String("ssh-file", config.FlagSSHFileDefault, config.FlagSSHFileDesc)
	pFlags.String("ssh-pass", config.FlagSSHPassDefault, config.FlagSSHPassDesc)
	pFlags.Int("ssh-retry-max", config.FlagSSHRetryMaxDefault, config.FlagSSHRetryMaxDesc)

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
