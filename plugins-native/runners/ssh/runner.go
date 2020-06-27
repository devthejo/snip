package mainNative

import (
	"fmt"
	"io"
	"log"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/goterm/term"
	shellquote "github.com/kballard/go-shellquote"
	"github.com/kvz/logstreamer"
	"github.com/sirupsen/logrus"
	"go.uber.org/multierr"

	expect "gitlab.com/youtopia.earth/ops/snip/goexpect"
	"gitlab.com/youtopia.earth/ops/snip/plugin/runner"
	"gitlab.com/youtopia.earth/ops/snip/sshclient"
	"gitlab.com/youtopia.earth/ops/snip/sshutils"
	"gitlab.com/youtopia.earth/ops/snip/tools"
)

var (
	Runner = runner.Plugin{
		UseVars:     []string{"pty", "host", "port", "user", "pass", "file", "sock", "max_retry"},
		GetRootPath: getRootPath,
		Run: func(cfg *runner.Config) error {

			logger := cfg.Logger

			vars := cfg.RunnerVars
			sshCfg := sshclient.CreateConfig(vars)

			rootPath := getRootPath(cfg)

			for src, dest := range cfg.RequiredFiles {
				_, err := tools.RequiredOnce(cfg.Cache, []string{"host", sshCfg.Host, dest}, src, func() (interface{}, error) {
					destAbs := filepath.Join(rootPath, dest)
					err := sshutils.Upload(sshCfg, src, destAbs, logger)
					return nil, err
				})
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

			commandSlice := []string{"/bin/sh", "-c", strings.Join(cfg.Command, " ")}

			command := shellquote.Join(commandSlice...)

			logger.Debugf("remote command: %v", command)

			session, err := client.NewSession()
			if err != nil {
				return err
			}

			var enablePTY bool
			if enablePTYStr, ok := vars["pty"]; ok {
				enablePTY = enablePTYStr == "true"
			}

			if enablePTY {
				tios := term.Termios{}
				tios.Raw()
				tios.Wz.WsCol, tios.Wz.WsRow = sshTermWidth, sshTermHeight
				err = session.RequestPty(sshTerm, int(tios.Wz.WsRow), int(tios.Wz.WsCol), tios.ToSSH())
				if err != nil {
					return err
				}
			}

			sIn, err := session.StdinPipe()
			if err != nil {
				return err
			}

			pr, pw := io.Pipe()
			sOut := pr
			session.Stdout = pw
			session.Stderr = pw

			err = session.Shell()
			if err != nil {
				return err
			}

			spawnOpts := &expect.SpawnOptions{
				In:  sIn,
				Out: sOut,
				Close: func() error {
					// session.Signal(ssh.SIGTERM)
					return session.Close()
				},
				Check: func() bool {
					_, err := session.SendRequest("dummy", false, nil)
					return err == nil
				},
				Wait: func() error {
					err := session.Wait()
					pr.Close()
					return err
				},
			}

			var expected []expect.Batcher

			if !cfg.Quiet {
				loggerSSH := logger.WithFields(logrus.Fields{
					"host": sshCfg.Host,
				})
				w := loggerSSH.Writer()
				defer w.Close()
				logStreamer := logstreamer.NewLogstreamer(log.New(w, "", 0), "", false)
				defer logStreamer.Close()

				sep := "#" + strconv.FormatInt(time.Now().UnixNano(), 10) + "#"
				wHeadStripper := wHeadStripperMake(sep)
				tee := &WriterModifier{
					Modifier: wHeadStripper,
					Writer:   logStreamer,
				}

				expected = append(expected, &expect.BSnd{S: `echo "` + sep + `"; `})

				spawnOpts.Tee = tee
			}

			e, ch, err := expect.Spawn(spawnOpts)
			if err != nil {
				return err
			}

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
						if !(*cfg.Closer)(session) {
							return
						}
					}
					e.Close()
					return
				}
			}()

			if cfg.Dir != "" {
				expected = append(expected, &expect.BSnd{S: "cd " + cfg.Dir + ";"})
			}

			var setenvSlice []string
			var setenv string
			env := cfg.EnvMap()

			appCfg := cfg.AppConfig

			vars["SNIP_SNIPPETS_PATH"] = filepath.Join(rootPath, "build", "snippets")
			vars["SNIP_LAUNCHERS_PATH"] = filepath.Join(rootPath, "build", "launchers")
			varsDir := appCfg.TreepathVarsDir(cfg.TreeKeyParts)
			vars["SNIP_VARS_TREEPATH"] = filepath.Join(rootPath, "vars", varsDir)

			if len(env) > 0 {
				for k, v := range env {
					setenvSlice = append(setenvSlice, k+"="+v)
				}
				setenv = "export " + shellquote.Join(setenvSlice...) + ";"
				expected = append(expected, &expect.BSnd{S: setenv})
			}

			expected = append(expected, &expect.BSnd{S: command + "; "})

			expected = append(expected, &expect.BSnd{S: "exit\n"})

			for _, v := range cfg.Expect {
				expected = append(expected, v)
			}

			e.ExpectBatch(expected, -1)

			err = <-ch
			if err != nil {
				return err
			}

			err = registerVarsRetrieve(cfg, client)
			if err != nil {
				return err
			}

			return nil
		},
	}
)

const (
	sshTerm       = "xterm"
	sshTermWidth  = 132
	sshTermHeight = 43
)

type WriterModifier struct {
	Modifier func([]byte) []byte
	Writer   io.WriteCloser
}

func (w *WriterModifier) Write(data []byte) (n int, err error) {
	data = w.Modifier(data)
	return w.Writer.Write(data)
}
func (w *WriterModifier) Close() error {
	return w.Writer.Close()
}

func getRootPath(cfg *runner.Config) string {
	username := cfg.RunnerVars["user"]
	if username == "" {
		usr, _ := user.Current()
		username = usr.Username
	}
	return filepath.Join("/home", username, ".snip", cfg.AppConfig.DeploymentName)
}

func registerVarsRetrieve(cfg *runner.Config, client *sshclient.Client) error {

	r := cfg.VarsRegistry
	kp := cfg.TreeKeyParts
	appCfg := cfg.AppConfig
	varDir := appCfg.TreepathVarsDir(kp)
	rootPath := getRootPath(cfg)
	varDirAbs := filepath.Join(rootPath, "vars", varDir)
	dp := kp[0 : len(kp)-2]
	var wg sync.WaitGroup
	var errs []error
	var vars []string
	vars = append(vars, cfg.RegisterVars...)
	if cfg.RegisterOutput != "" {
		vars = append(vars, cfg.RegisterOutput)
	}
	for _, vr := range vars {
		wg.Add(1)
		go func(vs string) {
			defer wg.Done()
			session, err := client.NewSession()
			if err != nil {
				errs = append(errs, err)
				return
			}
			file := filepath.Join(varDirAbs, vr)
			dat, err := session.Output(fmt.Sprintf("cat %s", file))
			if err != nil {
				errs = append(errs, err)
				return
			}
			session.Close()
			value := string(dat)
			value = strings.TrimSuffix(value, "\n")
			r.SetVarBySlice(dp, vr, value)
		}(vr)
	}
	wg.Wait()
	if len(errs) > 0 {
		return multierr.Combine(errs...)
	}
	return nil
}

func wHeadStripperMake(sep string) func(data []byte) []byte {
	sepL := len(sep)
	var enableWrite bool
	return func(data []byte) []byte {
		if !enableWrite {
			str := string(data)
			index := strings.Index(str, sep)
			if index == -1 {
				return []byte{}
			}
			enableWrite = true
			data = []byte(str[index+sepL:])
		}
		return data
	}
}
