package mainNative

import (
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/kvz/logstreamer"

	"gitlab.com/youtopia.earth/ops/snip/plugin/runner"
	"gitlab.com/youtopia.earth/ops/snip/tools"
)

var (
	Runner = runner.Runner{
		Run: func(cfg *runner.Config) error {

			var err error

			usr, _ := user.Current()
			snipDir := filepath.Join(usr.HomeDir, ".snip", cfg.AppConfig.DeploymentName)
			for src, dest := range cfg.RequiredFiles {
				destAbs := filepath.Join(snipDir, dest)
				dir := filepath.Dir(destAbs)
				os.MkdirAll(dir, os.ModePerm)
				_, err := tools.RequiredOnce(cfg.Cache, []string{"local", destAbs}, src, func() (interface{}, error) {
					return tools.Copy(src, destAbs)
				})
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

			cmd := exec.CommandContext(cfg.Context, commandPath, commandSlice[1:]...)

			if cfg.Stdin != nil {
				cmd.Stdin = cfg.Stdin
			}

			env := cfg.EnvMap()
			cmd.Env = tools.EnvToPairs(env)

			cmd.Dir = cfg.Dir

			logger := cfg.Logger
			w := logger.Writer()
			defer w.Close()
			logStreamer := logstreamer.NewLogstreamer(log.New(w, "", 0), "", false)
			defer logStreamer.Close()
			cmd.Stdout = logStreamer
			cmd.Stderr = logStreamer

			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

			go func() {
				select {
				case <-cfg.Context.Done():
					logger.Debug(`closing process`)
					if cfg.Closer != nil {
						if !(*cfg.Closer)(cmd) {
							return
						}
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
