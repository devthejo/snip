package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"gitlab.com/youtopia.earth/ops/snip/config"
	"gitlab.com/youtopia.earth/ops/snip/play"
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

			for _, playI := range playbook {
				p := play.ParseInterface(playI, app)
				logrus.Debug(p.Name)
			}

		},
	}

	flags := cmd.Flags()
	flags.StringP("snippets-dir", "", config.FlagSnippetsDirDefault, config.FlagSnippetsDirDesc)

	return cmd
}
