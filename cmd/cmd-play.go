package cmd

import (
	"github.com/spf13/cobra"

	"gitlab.com/youtopia.earth/ops/snip/config"
	"gitlab.com/youtopia.earth/ops/snip/play"
	"gitlab.com/youtopia.earth/ops/snip/tools"
)

func CmdPlay(app App) *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "play",
		Short: "let's play !",
		Args:  cobra.ExactArgs(0),
		PreRun: func(cmd *cobra.Command, args []string) {
			app.OnPreRun(cmd)
		},
		Run: func(cmd *cobra.Command, args []string) {

			mainFunc := func() error {
				if err := play.Clean(app); err != nil {
					return err
				}

				config := play.BuildConfig(app)
				p := play.BuildPlay(config)

				if err := p.Start(); err != nil {
					return err
				}
				if err := play.Clean(app); err != nil {
					return err
				}

				tools.PrintMemUsage()

				return nil
			}

			main := app.GetMainProc()
			main.Run(mainFunc)

		},
	}

	flags := cmd.Flags()
	flags.String("snippets-dir", config.FlagSnippetsDirDefault, config.FlagSnippetsDirDesc)
	flags.String("shutdown-timeout", config.FlagShutdownTimeoutDefault, config.FlagShutdownTimeoutDesc)

	flags.String("runner", config.FlagRunnerDefault, config.FlagRunnerDesc)
	flags.String("loaders", config.FlagLoadersDefault, config.FlagLoadersDesc)

	flags.StringP("deployment-name", "", config.FlagDeploymentNameDefault, config.FlagDeploymentNameDesc)

	return cmd
}
