package main

import (
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/kvz/logstreamer"

	"gitlab.com/youtopia.earth/ops/snip/plugin/runner"
	"gitlab.com/youtopia.earth/ops/snip/tools"
)

var (
	Runner = runner.Runner{
		Run: func(cfg *runner.Config) error {

			var err error

			usr, _ := user.Current()
			snipDir := usr.HomeDir + "/.snip/"
			for src, dest := range cfg.RequiredFiles {
				destAbs := snipDir + dest
				dir := filepath.Dir(destAbs)
				os.MkdirAll(dir, os.ModePerm)
				_, err := tools.CopyOnce(src, destAbs)
				if err != nil {
					return err
				}
			}

			commandSlice := make([]string, len(cfg.Command))
			for i, p := range cfg.Command {
				if strings.HasPrefix(p, "~/") {
					usr, err := user.Current()
					if err != nil {
						return err
					}
					p = filepath.Join(usr.HomeDir, p[2:])
				}
				commandSlice[i] = p
			}

			commandPath, err := exec.LookPath(commandSlice[0])
			if err != nil {
				return err
			}

			cmd := exec.CommandContext(*cfg.Context, commandPath, commandSlice[1:]...)

			if cfg.Stdin != nil {
				cmd.Stdin = cfg.Stdin
			}

			env := cfg.EnvMap()
			cmd.Env = tools.EnvToPairs(env)

			logger := cfg.Logger
			w := logger.Writer()
			defer w.Close()
			loggerOut := log.New(w, "", 0)
			logStreamer := logstreamer.NewLogstreamer(loggerOut, "", true)
			defer logStreamer.Close()
			cmd.Stdout = logStreamer
			cmd.Stderr = logStreamer
			logStreamer.FlushRecord()

			go func() {
				select {
				case <-(*cfg.Context).Done():
					logger.Debug(`closing process`)
					// cmd.Process.Signal(syscall.SIGTERM)
					if err := cmd.Process.Kill(); err != nil {
						logger.Warn("failed to kill process: ", err)
					}
					return
				}
			}()

			if err = cmd.Start(); err != nil {
				return err
			}

			if err = cmd.Wait(); err != nil {
				return err
			}

			return nil
		},
	}
)
