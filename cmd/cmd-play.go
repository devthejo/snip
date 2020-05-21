package cmd

import (
	"github.com/spf13/cobra"

	"gitlab.com/youtopia.earth/ops/snip/config"
	"gitlab.com/youtopia.earth/ops/snip/play"
)

func CmdPlay(app App, rootCmd *cobra.Command) *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "play",
		Short: "let's play !",
		Args:  cobra.ExactArgs(0),
		PreRun: func(cmd *cobra.Command, args []string) {
			app.OnPreRun(cmd)
		},
		Run: func(cmd *cobra.Command, args []string) {
			cfg := app.GetConfig()

			p := play.ParsePlay(app, cfg.Play, nil)
			// logrus.Infof("%v", tools.JsonEncode(p))
			// logrus.Infof("%v", p)

			p.PromptVars(nil)
			// play.Run(app, playbook, vars, vars_loops)

		},
	}

	flags := cmd.Flags()
	flags.StringP("snippets-dir", "", config.FlagSnippetsDirDefault, config.FlagSnippetsDirDesc)
	flags.StringP("build-dir", "", config.FlagBuildDirDefault, config.FlagBuildDirDesc)

	return cmd
}
