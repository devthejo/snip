package cmd

import (
	"github.com/spf13/cobra"

	"gitlab.com/youtopia.earth/ops/snip/config"
	"gitlab.com/youtopia.earth/ops/snip/play"
)

func CmdPlay(app App, rootCmd *cobra.Command) *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "play",
		Short: "run playbook",
		Args:  cobra.ExactArgs(0),
		PreRun: func(cmd *cobra.Command, args []string) {
			app.OnPreRun(cmd)
		},
		Run: func(cmd *cobra.Command, args []string) {
			cfg := app.GetConfig()

			playbook := play.LoadPlaybook(app, cfg.Playbook)

			play.BuildPlaybook(app, playbook)
			play.PromptVars(app, playbook)
			play.Run(app, playbook)

		},
	}

	flags := cmd.Flags()
	flags.StringP("snippets-dir", "", config.FlagSnippetsDirDefault, config.FlagSnippetsDirDesc)
	flags.StringP("build-dir", "", config.FlagBuildDirDefault, config.FlagBuildDirDesc)

	return cmd
}
