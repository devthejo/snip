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
		UseVars: []string{"pty"},
		Run: func(cfg *runner.Config) error {

			var err error

			usr, _ := user.Current()
			snipDir := filepath.Join(usr.HomeDir, ".snip", cfg.AppConfig.DeploymentName)
			for src, dest := range cfg.RequiredFiles {
				destAbs := filepath.Join(snipDir, dest)
				dir := filepath.Dir(destAbs)
				if err := os.MkdirAll(dir, os.ModePerm); err != nil {
					return err
				}
				_, err := tools.RequiredOnce(cfg.Cache, []string{"local", destAbs}, src, func() (interface{}, error) {
					return tools.Copy(src, destAbs)
				})
				if err != nil {
					return err
				}
			}

			logger := cfg.Logger

			w := logger.Writer()
			defer w.Close()
			logStreamer := logstreamer.NewLogstreamer(log.New(w, "", 0), "", false)
			defer logStreamer.Close()

			// commandSlice := []string{"."}
			commandSlice := []string{}
			commandSlice = append(commandSlice, strings.Join(cfg.Command, " ")+";")

			varDirAbs := GetVarsDir(cfg)
			if err := os.MkdirAll(varDirAbs, os.ModePerm); err != nil {
				return err
			}
			var rVars []string
			for _, vr := range cfg.RegisterVars {
				file := filepath.Join(varDirAbs, vr)
				rVars = append(rVars, `echo -n "${`+strings.ToUpper(vr)+`}">`+file+";")
			}
			commandSlice = append(commandSlice, strings.Join(rVars, " "))

			commandSlice = []string{"/bin/sh", "-c", strings.Join(commandSlice, " ")}

			// logrus.Warn(commandSlice)

			cmd := exec.CommandContext(cfg.Context, commandSlice[0], commandSlice[1:]...)

			cmd.Dir = cfg.Dir

			env := cfg.EnvMap()

			appCfg := cfg.AppConfig
			snipPath := filepath.Join(usr.HomeDir, ".snip", appCfg.DeploymentName, appCfg.BuildDir, "snippets")
			env["SNIP_PATH"] = snipPath

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
			if enablePTY {
				pty, err = term.OpenPTY()
				if err != nil {
					return err
				}
				var t term.Termios
				t.Raw()
				t.Set(pty.Slave)
				cmd.Stdin, cmd.Stdout, cmd.Stderr = pty.Slave, pty.Slave, pty.Slave
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
				stdout, err := cmd.StdoutPipe()
				if err != nil {
					return err
				}
				stderr, err := cmd.StderrPipe()
				if err != nil {
					return err
				}
				sOut = io.MultiReader(stdout, stderr)
				wait = cmd.Wait
			}

			if err := cmd.Start(); err != nil {
				return err
			}

			e, ch, err := expect.Spawn(&expect.SpawnOptions{
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
				Tee: logStreamer,
				// Verbose: true,
				// VerboseWriter: logStreamer,
				Clean: clean,
			})

			defer e.Close()

			go func() {
				select {
				case <-cfg.Context.Done():
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

func GetVarsDir(cfg *runner.Config) string {
	kp := cfg.TreeKeyParts
	dp := kp[0 : len(kp)-2]
	dirParts := make([]string, len(dp))
	for i := 0; i < len(dp); i++ {
		var p string
		if i%2 == 0 {
			p = "key"
		} else {
			p = "row"
		}
		p += "."
		p += dp[i]
		dirParts[i] = p
	}
	varDir := filepath.Join(dirParts...)

	appCfg := cfg.AppConfig
	usr, _ := user.Current()
	varsPath := filepath.Join(usr.HomeDir, ".snip", appCfg.DeploymentName, "vars")
	varDirAbs := filepath.Join(varsPath, varDir)
	return varDirAbs
}

func registerVarsRetrieve(cfg *runner.Config) error {
	r := cfg.VarsRegistry
	varDirAbs := GetVarsDir(cfg)
	kp := cfg.TreeKeyParts
	dp := kp[0 : len(kp)-2]
	for _, vr := range cfg.RegisterVars {
		file := filepath.Join(varDirAbs, vr)
		dat, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		r.SetVarBySlice(dp, vr, string(dat))
	}
	return nil
}
