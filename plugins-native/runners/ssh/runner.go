package mainNative

import (
	"io"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/goterm/term"
	shellquote "github.com/kballard/go-shellquote"
	"github.com/kvz/logstreamer"
	"github.com/sirupsen/logrus"
	expect "gitlab.com/youtopia.earth/ops/snip/goexpect"
	"gitlab.com/youtopia.earth/ops/snip/plugin/runner"
	"gitlab.com/youtopia.earth/ops/snip/sshclient"
	"gitlab.com/youtopia.earth/ops/snip/sshutils"
	"gitlab.com/youtopia.earth/ops/snip/tools"
)

var (
	Runner = runner.Plugin{
		Run: func(cfg *runner.Config) error {

			logger := cfg.Logger

			sshCfg := sshclient.CreateConfig(cfg.Vars)

			for src, dest := range cfg.RequiredFiles {
				_, err := tools.RequiredOnce(cfg.Cache, []string{"host", sshCfg.Host, dest}, src, func() (interface{}, error) {
					destAbs := filepath.Join("/home", sshCfg.User, ".snip", cfg.AppConfig.DeploymentName, dest)
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
			if enablePTYStr, ok := cfg.Vars["@PTY"]; ok {
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
			stdout, err := session.StdoutPipe()
			if err != nil {
				return err
			}
			stderr, err := session.StderrPipe()
			if err != nil {
				return err
			}
			sOut := io.MultiReader(stdout, stderr)

			loggerSSH := logger.WithFields(logrus.Fields{
				"host": sshCfg.Host,
			})
			w := loggerSSH.Writer()
			defer w.Close()
			logStreamer := logstreamer.NewLogstreamer(log.New(w, "", 0), "", false)
			defer logStreamer.Close()

			err = session.Shell()
			if err != nil {
				return err
			}

			sep := "#" + strconv.FormatInt(time.Now().UnixNano(), 10) + "#"
			sepL := len(sep)
			var enableWrite bool
			tee := &WriterModifier{
				Modifier: func(data []byte) []byte {
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
				},
				Writer: logStreamer,
			}

			e, ch, err := expect.Spawn(&expect.SpawnOptions{
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
				Wait: session.Wait,
				Tee:  tee,
			})
			if err != nil {
				return err
			}
			defer e.Close()

			go func() {
				if cfg.Closer != nil {
					if !(*cfg.Closer)(session) {
						return
					}
				}
				select {
				case <-cfg.Context.Done():
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

			var expected []expect.Batcher

			expected = append(expected, &expect.BSnd{S: `echo "` + sep + `"; `})

			if cfg.Dir != "" {
				expected = append(expected, &expect.BSnd{S: "cd " + cfg.Dir + ";"})
			}

			var setenvSlice []string
			var setenv string
			env := cfg.EnvMap()

			appCfg := cfg.AppConfig
			snipPath := filepath.Join("/home", sshCfg.User, ".snip", appCfg.DeploymentName, appCfg.BuildDir, "snippets")
			env["SNIP_PATH"] = snipPath

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

			return <-ch
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
