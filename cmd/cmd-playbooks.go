package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	snipApp "github.com/devthejo/snip/app"
	"github.com/devthejo/snip/config"
	"github.com/devthejo/snip/play"
	"github.com/devthejo/snip/tools"
)

func CmdPlaybooks(app App) *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "playbooks [files, ...]",
		Short: "Let's run the playbooks !",
		Args:  cobra.ArbitraryArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
			app.OnPreRun(cmd)
		},
		Run: func(cmd *cobra.Command, args []string) {

			cfg := app.GetConfig()

			var files []string
			if len(args) > 0 {
				files = args
			} else {
				files = cfg.Playbooks
			}

			startTime := time.Now()

			runReport := &play.RunReport{}

			for _, file := range files {
				playbookApp := snipApp.NewApp(app.GetVersion())
				cl := playbookApp.ConfigLoader
				pbCfg := cl.Config
				pbCfgFile := "playbooks/" + file
				cl.File = &pbCfgFile
				// cl.ConfigPaths = []string{"playbooks"}
				// cl.ConfigName = file
				pbCfg.DeploymentName = cfg.DeploymentName + "." + file

				pbCfg.CWD = cfg.CWD
				if pbCfg.CWD != "" {
					err := os.Chdir(pbCfg.CWD)
					if err != nil {
						panic(err)
					}
				}

				pbCfg.LogLevel = cfg.LogLevel
				pbCfg.LogType = cfg.LogType
				pbCfg.LogForceColors = cfg.LogForceColors
				pbCfg.SnippetsDir = cfg.SnippetsDir
				pbCfg.MarkdownOutput = cfg.MarkdownOutput
				pbCfg.ShutdownTimeout = cfg.ShutdownTimeout
				pbCfg.Runner = cfg.Runner
				pbCfg.Loaders = cfg.Loaders
				pbCfg.PlayKey = cfg.PlayKey
				pbCfg.PlayKeyDeps = cfg.PlayKeyDeps
				pbCfg.PlayKeyPost = cfg.PlayKeyPost
				pbCfg.PlayKeyStart = cfg.PlayKeyStart
				pbCfg.PlayKeyEnd = cfg.PlayKeyEnd
				pbCfg.PlayResume = cfg.PlayResume
				pbCfg.PlayNoClean = cfg.PlayNoClean

				playbookApp.Aurora = app.GetAurora()

				playbookApp.InitVarsRegistry()
				cl.LoadJsonnet()
				cl.InitViper()
				cl.LoadViperConfigFile()
				cl.LoadViper()

				err, playbookRunReport := RunPlay(playbookApp)
				if err != nil {
					break
				}

				if playbookRunReport != nil {
					runReport.OK += playbookRunReport.OK
					runReport.Failed += playbookRunReport.Failed
					runReport.Changed += playbookRunReport.Changed
					runReport.Total += playbookRunReport.Total
				}
			}

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

	return cmd
}
