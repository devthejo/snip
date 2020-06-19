package mainNative

import (
	"io"
	"io/ioutil"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/goterm/term"
	"github.com/kvz/logstreamer"
	"github.com/sirupsen/logrus"
	expect "gitlab.com/youtopia.earth/ops/snip/goexpect"
	"gitlab.com/youtopia.earth/ops/snip/plugin/runner"
	"gitlab.com/youtopia.earth/ops/snip/sshclient"
	"gitlab.com/youtopia.earth/ops/snip/sshutils"
)

var (
	Runner = runner.Runner{
		Run: func(cfg *runner.Config) error {

			logger := cfg.Logger

			env := cfg.EnvMap()
			sshCfg := sshclient.CreateConfig(env)

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

			command := strings.Join(commandSlice, " ")

			var setenvSlice []string
			setenv := ""
			for k, v := range env {
				setenvSlice = append(setenvSlice, k+"="+v)
			}
			setenv = strings.Join(setenvSlice, " ")

			runCmdSlice := []string{setenv, command}
			runCmd := strings.Join(runCmdSlice, " ")

			logger.Debugf("remote command: %v", runCmd)

			var opts []expect.Option

			session, err := client.NewSession()
			if err != nil {
				return err
			}

			tios := term.Termios{}
			tios.Raw()
			tios.Wz.WsCol, tios.Wz.WsRow = sshTermWidth, sshTermHeight
			err = session.RequestPty(sshTerm, int(tios.Wz.WsRow), int(tios.Wz.WsCol), tios.ToSSH())
			if err != nil {
				return err
			}

			sIn, err := session.StdinPipe()
			if err != nil {
				return err
			}
			sOut, err := session.StdoutPipe()
			if err != nil {
				return err
			}

			loggerSSH := logger.WithFields(logrus.Fields{
				"ssh":  true,
				"host": sshCfg.Host,
			})
			w := loggerSSH.Writer()
			defer w.Close()
			loggerOut := log.New(w, "", 0)
			logStreamer := logstreamer.NewLogstreamer(loggerOut, "", true)
			defer logStreamer.Close()
			logStreamer.FlushRecord()

			err = session.Shell()
			if err != nil {
				return err
			}

			e, ch, err := expect.SpawnGeneric(&expect.GenOptions{
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
			}, -1, opts...)
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

			e.ExpectBatch(cfg.Expect, -1)

			// e.Options(expect.Tee(logStreamer))
			sep := "#" + strconv.FormatInt(time.Now().UnixNano(), 10) + "#"
			sepL := len(sep)
			var enableWrite bool
			e.Options(expect.Tee(&WriterModifier{
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
			}))

			var expected []expect.Batcher

			expected = append(expected, &expect.BSnd{S: `echo "` + sep + `"` + " &&"})
			expected = append(expected, &expect.BSnd{S: runCmd + "\n"})

			if cfg.Stdin != nil {
				b, err := ioutil.ReadAll(cfg.Stdin)
				if err != nil {
					return err
				}
				expected = append(expected, &expect.BSnd{S: string(b)})
			}

			for _, v := range cfg.Expect {
				expected = append(expected, v)
			}

			expected = append(expected, &expect.BSnd{S: "exit\n"})

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
