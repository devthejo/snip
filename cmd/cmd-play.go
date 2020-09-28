package cmd

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"gitlab.com/ytopia/ops/snip/config"
	"gitlab.com/ytopia/ops/snip/play"
	"gitlab.com/ytopia/ops/snip/tools"
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

			startTime := time.Now()

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
				runReport := playCfg.GlobalRunCtx.RunReport
				logrus.Infof("🏁 play report:")
				logrus.Infof("  result: %s %s %s",
					au.BrightGreen(fmt.Sprintf("OK=%d", runReport.OK)),
					au.BrightMagenta(fmt.Sprintf("Changed=%d", runReport.Changed)),
					au.BrightBlue(fmt.Sprintf("Total=%d", runReport.Total)))

				logrus.Infof("  duration %s", time.Since(startTime).Round(time.Second))

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
	flags.Bool("key-no-deps", config.FlagPlayKeyNoDepsDefault, config.FlagPlayKeyNoDepsDesc)
	flags.Bool("key-no-post", config.FlagPlayKeyNoPostDefault, config.FlagPlayKeyNoPostDesc)
	flags.String("key-start", config.FlagPlayKeyStartDefault, config.FlagPlayKeyStartDesc)
	flags.String("key-end", config.FlagPlayKeyEndDefault, config.FlagPlayKeyEndDesc)

	return cmd
}
