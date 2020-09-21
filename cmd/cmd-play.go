package cmd

import (
	"fmt"

	"github.com/sirupsen/logrus"
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

				playCfg := play.BuildConfig(app)
				p := play.BuildPlay(playCfg)

				if err := p.Start(); err != nil {
					return err
				}
				if err := play.Clean(app); err != nil {
					return err
				}

				au := app.GetAurora()
				runReport := playCfg.RunReport
				logrus.Infof("Report of Play: %s %s %s",
					au.BrightGreen(fmt.Sprintf("OK=%d", runReport.OK)),
					au.BrightMagenta(fmt.Sprintf("Changed=%d", runReport.Changed)),
					au.BrightBlue(fmt.Sprintf("Total=%d", runReport.Total)))

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

	flags.StringP("deployment-name", "n", config.FlagDeploymentNameDefault, config.FlagDeploymentNameDesc)

	flags.StringSliceP("key", "k", config.FlagPlayKeyDefault, config.FlagPlayKeyDesc)

	return cmd
}
