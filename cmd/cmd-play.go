package cmd

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
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

			cfg := app.GetConfig()

			startTime := time.Now()

			main := app.GetMainProc()

			mainFunc := func() error {
				if err := play.Clean(app); err != nil {
					return err
				}

				resumeFile := filepath.Join(play.GetRootPath(app), "resume")
				if cfg.PlayResume {
					if resume, err := ioutil.ReadFile(resumeFile); err == nil {
						logrus.Infof("🔁 resuming from %s", resume)
						cfg.PlayKeyStart = string(resume)
					}
				}

				playCfg := play.BuildConfig(app)

				p := play.BuildPlay(playCfg)

				gRunCtx := playCfg.GlobalRunCtx

				err := p.Start()

				if main.Success {
					gRunCtx.CurrentTreeKey = ""
				}

				logrus.Debugf("resume saved: %v", gRunCtx.CurrentTreeKey)
				ioutil.WriteFile(resumeFile, []byte(gRunCtx.CurrentTreeKey), 0644)

				if !cfg.PlayNoClean {
					play.Clean(app)
				}

				if err != nil {
					return err
				}

				au := app.GetAurora()
				runReport := gRunCtx.RunReport
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

				return nil
			}

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
	flags.Bool("key-deps", config.FlagPlayKeyDepsDefault, config.FlagPlayKeyDepsDesc)
	flags.Bool("key-post", config.FlagPlayKeyPostDefault, config.FlagPlayKeyPostDesc)
	flags.String("key-start", config.FlagPlayKeyStartDefault, config.FlagPlayKeyStartDesc)
	flags.String("key-end", config.FlagPlayKeyEndDefault, config.FlagPlayKeyEndDesc)
	flags.Bool("no-clean", config.FlagPlayNoCleanDefault, config.FlagPlayNoCleanDesc)
	flags.BoolP("resume", "r", config.FlagPlayResumeDefault, config.FlagPlayResumeDesc)

	return cmd
}
