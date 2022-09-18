package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/devthejo/snip/config"
	"github.com/devthejo/snip/tools"
)

func CmdPlay(app App) *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "play",
		Short: "Let's run the playbook !",
		Args:  cobra.ExactArgs(0),
		PreRun: func(cmd *cobra.Command, args []string) {
			app.OnPreRun(cmd)
		},
		Run: func(cmd *cobra.Command, args []string) {

			startTime := time.Now()

			_, runReport := RunPlay(app)

			if runReport != nil {
				au := app.GetAurora()
				logrus.Infof("🏁 play report:")
				resultMsg := "  result: %s %s %s"
				resultVars := []interface{}{
					au.BrightGreen(fmt.Sprintf("OK=%d", runReport.OK)),
					au.BrightMagenta(fmt.Sprintf("Changed=%d", runReport.Changed)),
				}
				if runReport.Failed > 0 {
					resultMsg += " %s"
					resultVars = append(resultVars, au.BrightRed(fmt.Sprintf("Failed=%d", runReport.Failed)))
				}
				resultVars = append(resultVars, au.BrightBlue(fmt.Sprintf("Total=%d", runReport.Total)))
				logrus.Infof(resultMsg, resultVars...)

				logrus.Infof("  duration %s", time.Since(startTime).Round(time.Second))

				tools.PrintMemUsage()
			}

			proc := app.GetMainProc()
			os.Exit(proc.ExitCode)

		},
	}

	flags := cmd.Flags()
	flags.String("snippets-dir", config.FlagSnippetsDirDefault, config.FlagSnippetsDirDesc)
	flags.String("shutdown-timeout", config.FlagShutdownTimeoutDefault, config.FlagShutdownTimeoutDesc)

	flags.String("runner", config.FlagRunnerDefault, config.FlagRunnerDesc)
	flags.String("loaders", config.FlagLoadersDefault, config.FlagLoadersDesc)

	flags.StringP("deployment-name", "n", config.FlagDeploymentNameDefault, config.FlagDeploymentNameDesc)

	flags.StringSliceP("key", "k", config.FlagPlayKeyDefault, config.FlagPlayKeyDesc)
	flags.Bool("key-deps", config.FlagPlayKeyDepsDefault, config.FlagPlayKeyDepsDesc)
	flags.Bool("key-post", config.FlagPlayKeyPostDefault, config.FlagPlayKeyPostDesc)
	flags.String("key-start", config.FlagPlayKeyStartDefault, config.FlagPlayKeyStartDesc)
	flags.String("key-end", config.FlagPlayKeyEndDefault, config.FlagPlayKeyEndDesc)
	flags.Bool("no-clean", config.FlagPlayNoCleanDefault, config.FlagPlayNoCleanDesc)
	flags.BoolP("resume", "r", config.FlagPlayResumeDefault, config.FlagPlayResumeDesc)

	return cmd
}
