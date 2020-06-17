package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	expect "github.com/google/goexpect"
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

			logger := cfg.Logger

			var opts []expect.Option

			env := cfg.EnvMap()
			opts = append(opts, expect.SetEnv(tools.EnvToPairs(env)))

			w := logger.Writer()
			defer w.Close()
			loggerOut := log.New(w, "", 0)
			logStreamer := logstreamer.NewLogstreamer(loggerOut, "", true)
			defer logStreamer.Close()
			opts = append(opts, expect.Tee(logStreamer))
			logStreamer.FlushRecord()

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
			commandSlice = append([]string{commandPath}, commandSlice[1:]...)

			e, ch, err := expect.Spawn(strings.Join(commandSlice, " "), -1, opts...)
			if err != nil {
				return err
			}
			defer e.Close()

			go func() {
				select {
				case <-(*cfg.Context).Done():
					logger.Debug(`closing process`)
					if err := e.Close(); err != nil {
						logger.Warn("failed to kill process: ", err)
					}
					return
				}
			}()

			if cfg.Stdin != nil {
				b, err := ioutil.ReadAll(cfg.Stdin)
				if err != nil {
					return err
				}
				e.Send(string(b))
			}

			e.ExpectBatch(cfg.Expect, -1)

			return <-ch
		},
	}
)
