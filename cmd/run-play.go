package cmd

import (
	"io/ioutil"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/devthejo/snip/play"
)

func RunPlay(app App) (error, *play.RunReport) {

	cfg := app.GetConfig()

	main := app.GetMainProc()

	mainFunc := func() (error, interface{}) {
		if err := play.Clean(app); err != nil {
			return err, nil
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
			return err, nil
		}

		runReport := gRunCtx.RunReport

		return nil, runReport

	}

	err, runReportI := main.RunMain(mainFunc)

	var runReport *play.RunReport
	if runReportI != nil {
		runReport = runReportI.(*play.RunReport)
	}

	return err, runReport
}
