package mainNative

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/google/goterm/term"
	"github.com/kvz/logstreamer"

	expect "gitlab.com/youtopia.earth/ops/snip/goexpect"
	"gitlab.com/youtopia.earth/ops/snip/plugin/runner"
	"gitlab.com/youtopia.earth/ops/snip/tools"
)

var (
	Runner = runner.Plugin{
		UseVars:     []string{"pty"},
		GetRootPath: getRootPath,
		Run: func(cfg *runner.Config) error {

			if err := installRequiredFiles(cfg); err != nil {
				return err
			}

			if err := registerVarsCreateFiles(cfg); err != nil {
				return err
			}

			var err error

			logger := cfg.Logger

			commandSlice := []string{"/bin/sh", "-c", strings.Join(cfg.Command, " ")}

			logger.Debugf("final command: %v", commandSlice)

			cmd := exec.CommandContext(cfg.Context, commandSlice[0], commandSlice[1:]...)

			cmd.Dir = cfg.Dir

			env := cfg.EnvMap()
			cmd.Env = tools.EnvToPairs(env)

			var sIn io.WriteCloser
			var sOut io.Reader

			var wait func() error
			var clean func()

			var enablePTY bool
			if enablePTYStr, ok := cfg.RunnerVars["pty"]; ok {
				enablePTY = enablePTYStr == "true"
			}

			var pty *term.PTY

			var pr io.ReadCloser
			var pw io.WriteCloser

			if enablePTY {
				pty, err = term.OpenPTY()
				if err != nil {
					return err
				}
				var t term.Termios
				t.Raw()
				t.Set(pty.Slave)

				pw = pty.Slave
				pr = pty.Slave
				cmd.Stdin = pty.Slave

				cmd.SysProcAttr = &syscall.SysProcAttr{
					Setsid:  true,
					Setctty: true,
				}
				sIn = pty.Master
				sOut = pty.Master
				clean = func() {
					pty.Master.Close()
				}
				wait = func() error {
					if err := cmd.Wait(); err != nil {
						return err
					}
					if err := pty.Slave.Close(); err != nil {
						return err
					}
					return nil
				}

			} else {
				cmd.SysProcAttr = &syscall.SysProcAttr{
					Setpgid: true,
				}
				sIn, err = cmd.StdinPipe()
				if err != nil {
					return err
				}

				pr, pw = io.Pipe()
				sOut = pr

				wait = func() error {
					err := cmd.Wait()
					pr.Close()
					return err
				}
			}

			cmd.Stdout = pw
			cmd.Stderr = pw

			if err := cmd.Start(); err != nil {
				return err
			}

			spawnOpts := &expect.SpawnOptions{
				In:  sIn,
				Out: sOut,
				Close: func() error {
					sIn.Close()
					if pty != nil {
						pty.Close()
					}
					if cmd.Process != nil {
						return cmd.Process.Kill()
					}
					return nil
				},
				Wait: wait,
				Check: func() bool {
					if cmd.Process == nil {
						return false
					}
					// Sending Signal 0 to a process returns nil if process can take a signal , something else if not.
					return cmd.Process.Signal(syscall.Signal(0)) == nil
				},
				// Verbose: true,
				// VerboseWriter: logStreamer,
				Clean: clean,
			}

			if !cfg.Quiet {
				w := logger.Writer()
				defer w.Close()
				logStreamer := logstreamer.NewLogstreamer(log.New(w, "", 0), "", false)
				defer logStreamer.Close()

				spawnOpts.Tee = logStreamer
			}

			e, ch, err := expect.Spawn(spawnOpts)

			var isClosed bool
			defer func() {
				isClosed = true
				e.Close()
			}()

			go func() {
				select {
				case <-cfg.Context.Done():
					if isClosed {
						return
					}
					logger.Debug(`closing process`)
					if cfg.Closer != nil {
						if !(*cfg.Closer)(cmd) {
							return
						}
					}
					e.Close()
					return
				}
			}()

			var expected []expect.Batcher

			for _, v := range cfg.Expect {
				expected = append(expected, v)
			}

			e.ExpectBatch(expected, -1)

			err = <-ch
			if err != nil {
				return err
			}

			err = registerVarsRetrieve(cfg)
			if err != nil {
				return err
			}

			return nil
		},
	}
)

func getRootPath(cfg *runner.Config) string {
	usr, _ := user.Current()
	return filepath.Join(usr.HomeDir, ".snip", cfg.AppConfig.DeploymentName)
}

func getVarsPath(cfg *runner.Config) string {
	kp := cfg.TreeKeyParts
	appCfg := cfg.AppConfig
	varDir := appCfg.TreeDirVars(kp)
	rootPath := getRootPath(cfg)
	return filepath.Join(rootPath, "vars", varDir)
}

func installRequiredFiles(cfg *runner.Config) error {
	rootPath := getRootPath(cfg)
	for src, dest := range cfg.RequiredFiles {
		destAbs := filepath.Join(rootPath, dest)
		dir := filepath.Dir(destAbs)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
		_, err := tools.RequiredOnce(cfg.Cache, []string{"local", destAbs}, src, func() (interface{}, error) {
			if src == destAbs {
				return nil, nil
			}
			return tools.Copy(src, destAbs)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func registerVarsCreateFiles(cfg *runner.Config) error {
	varsPath := getVarsPath(cfg)
	var vars []string
	vars = append(vars, cfg.RegisterVars...)
	if cfg.RegisterOutput != "" {
		vars = append(vars, cfg.RegisterOutput)
	}
	for _, vr := range vars {
		file := filepath.Join(varsPath, vr)
		dir := filepath.Dir(file)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
		if _, err := os.OpenFile(file, os.O_RDONLY|os.O_CREATE, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func registerVarsRetrieve(cfg *runner.Config) error {
	varsPath := getVarsPath(cfg)
	kp := cfg.TreeKeyParts
	dp := kp[0 : len(kp)-2]
	var vars []string
	vars = append(vars, cfg.RegisterVars...)
	if cfg.RegisterOutput != "" {
		vars = append(vars, cfg.RegisterOutput)
	}
	r := cfg.VarsRegistry
	for _, vr := range vars {
		file := filepath.Join(varsPath, vr)
		dat, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		value := string(dat)
		value = strings.TrimSuffix(value, "\n")
		r.SetVarBySlice(dp, vr, value)
	}
	return nil
}
