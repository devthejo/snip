package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func CmdBash(app App, rootCmd *cobra.Command) *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "bash",
		Short: "Generates bash scripts",
		Args:  cobra.ExactArgs(0),
		PreRun: func(cmd *cobra.Command, args []string) {
			app.OnPreRun(cmd)
		},
		Run: func(cmd *cobra.Command, args []string) {
			cfg := app.GetConfig()
			playbook := cfg.Playbook
			logrus.Info(playbook)

		},
	}

	return cmd
}
