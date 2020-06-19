package mainNative

import (
	"log"
	"path/filepath"
	"strings"

	shellquote "github.com/kballard/go-shellquote"
	"github.com/kvz/logstreamer"
	"github.com/sirupsen/logrus"
	"gitlab.com/youtopia.earth/ops/snip/plugin/runner"
	"gitlab.com/youtopia.earth/ops/snip/sshclient"
	"gitlab.com/youtopia.earth/ops/snip/sshutils"
)

var (
	Runner = runner.Runner{
		Run: func(cfg *runner.Config) error {

			logger := cfg.Logger

			sshCfg := sshclient.CreateConfig(cfg.Vars)

			for src, dest := range cfg.RequiredFiles {
				err := sshutils.Upload(sshCfg, src, dest, logger)
				if err != nil {
					return err
				}
			}

			client, err := sshclient.CreateClient(sshCfg)
			if err != nil {
				return err
			}

			if err := client.Connect(); err != nil {
				return err
			}
			defer client.Close()

			commandSlice := make([]string, len(cfg.Command))
			for i, p := range cfg.Command {
				if strings.HasPrefix(p, "~/") {
					p = filepath.Join("/home", sshCfg.User, p[2:])
				}
				commandSlice[i] = p
			}

			command := shellquote.Join(commandSlice...)

			var cd string
			if cfg.Dir != "" {
				cd = "cd " + cfg.Dir + ";"
			}

			var setenvSlice []string
			var setenv string
			env := cfg.EnvMap()
			if len(env) > 0 {
				for k, v := range env {
					setenvSlice = append(setenvSlice, k+"="+v)
				}
				setenv = "export " + shellquote.Join(setenvSlice...) + ";"
			}

			logger.Debugf("remote command: %v", command)

			session, err := client.NewSession()
			if err != nil {
				return err
			}
			defer session.Close()

			loggerSSH := logger.WithFields(logrus.Fields{
				"ssh":  true,
				"host": sshCfg.Host,
			})
			w := loggerSSH.Writer()
			defer w.Close()
			logStreamer := logstreamer.NewLogstreamer(log.New(w, "", 0), "", false)
			defer logStreamer.Close()
			session.Stdout = logStreamer
			session.Stderr = logStreamer

			go func() {
				select {
				case <-cfg.Context.Done():
					logger.Debug(`closing process`)
					if cfg.Closer != nil {
						if !(*cfg.Closer)(session) {
							return
						}
					}
					// session.Signal(ssh.SIGTERM)
					session.Close()
					return
				}
			}()

			session.Stdin = cfg.Stdin

			runCmd := strings.Join([]string{cd, setenv, command}, " ")

			if err := session.Start(runCmd); err != nil {
				return err
			}

			if err := session.Wait(); err != nil {
				return err
			}

			return nil
		},
	}
)
