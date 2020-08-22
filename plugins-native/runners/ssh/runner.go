package mainNative

import (
	"bytes"
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
	cache "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"go.uber.org/multierr"

	expect "gitlab.com/youtopia.earth/ops/snip/goexpect"
	"gitlab.com/youtopia.earth/ops/snip/loggers"
	"gitlab.com/youtopia.earth/ops/snip/plugin/runner"
	"gitlab.com/youtopia.earth/ops/snip/registry"
	"gitlab.com/youtopia.earth/ops/snip/sshclient"
	"gitlab.com/youtopia.earth/ops/snip/sshutils"
	"gitlab.com/youtopia.earth/ops/snip/tools"
)

var (
	Runner = runner.Plugin{
		UseVars:     []string{"pty", "host", "port", "user", "pass", "file", "sock", "max_retry"},
		GetRootPath: getRootPath,
		Run: func(cfg *runner.Config) error {

			vars := cfg.RunnerVars

			connect := func(vars map[string]string) (*sshclient.Client, *sshclient.Config, error) {
				sshCfg := sshclient.CreateConfig(vars)
				client, err := sshclient.CreateClient(sshCfg)
				if err == nil {
					err = client.Connect()
				}
				return client, sshCfg, err
			}

			r := cfg.Cache

			var client *sshclient.Client
			var sshCfg *sshclient.Config
			if port, ok := vars["port"]; ok && strings.Contains(port, ",") {
				cacheKey := "host:port:" + vars["host"]
				var portFirst string
				if portI, found := r.Get(cacheKey); found {
					portFirst = portI.(string)
				}

				portsList := strings.Split(port, ",")
				var ports []string
				if portFirst != "" {
					ports = append(ports, portFirst)
				}
				for _, p := range portsList {
					if p != portFirst {
						ports = append(ports, p)
					}
				}
				var err error
				for _, p := range ports {
					varsP := make(map[string]string)
					for k, v := range vars {
						varsP[k] = v
					}
					varsP["port"] = p
					client, sshCfg, err = connect(varsP)
					if err == nil {
						r.Set(cacheKey, p, cache.DefaultExpiration)
						break
					}
				}
				if err != nil {
					return err
				}

			} else {
				var err error
				client, sshCfg, err = connect(vars)
				if err != nil {
					return err
				}
			}

			defer client.Close()

			if err := installRequiredFiles(cfg, sshCfg); err != nil {
				return err
			}

			if err := registerVarsCreateFiles(cfg, client); err != nil {
				return err
			}

			var enablePTY bool
			if enablePTYStr, ok := vars["pty"]; ok {
				enablePTY = enablePTYStr == "true"
			}

			logger := cfg.Logger

			var commandSlice []string
			if !enablePTY {
				commandSlice = append(commandSlice, "setsid")
			}
			commandSlice = append(commandSlice, "/bin/sh", "-c", "echo $$ PGID && "+strings.Join(cfg.Command, " "))

			command := shellquote.Join(commandSlice...)

			logger.Debugf("remote command: %v", command)

			session, err := client.NewSession()
			if err != nil {
				return err
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

			var pgid string

			spawnOpts := &expect.SpawnOptions{
				In:  sIn,
				Out: sOut,
				Close: func() error {
					// session.Signal(ssh.SIGKILL)
					if pgid != "" {
						if sess, err := client.NewSession(); err == nil {
							// sess.Run(`[ -n "$(ps -p ` + pgid + ` -o pid=)" ] && kill -s KILL ` + pgid)
							sess.Run(`kill -s KILL ` + pgid)
							pgid = ""
						}
					}
					if err := session.Close(); err != nil {
						return err
					}
					return nil
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
				wHeadStripper := makeWriterHeadStripperAndGetPGID(sep, &pgid)
				tee := &WriterModifier{
					Modifier: wHeadStripper,
					Writer:   logStreamer,
				}

				expected = append(expected, &expect.BSnd{S: `echo "` + sep + `"; `})

				spawnOpts.Tee = tee

				logStreamerErr := logstreamer.NewLogstreamer(log.New(&loggers.Warn{Entry: loggerSSH}, "", 0), "", false)
				defer logStreamerErr.Close()
				session.Stderr = io.MultiWriter(session.Stderr, logStreamerErr)
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
						if !(*cfg.Closer)(client.Client, &pgid) {
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

			if len(env) > 0 {
				for k, v := range env {
					setenvSlice = append(setenvSlice, k+"="+v)
				}
				setenv = "export " + shellquote.Join(setenvSlice...) + ";"
				expected = append(expected, &expect.BSnd{S: setenv})
			}

			expected = append(expected, &expect.BSnd{S: command + "; exit;\n"})

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

func getVarsPath(cfg *runner.Config) string {
	kp := cfg.TreeKeyParts
	appCfg := cfg.AppConfig
	varDir := appCfg.TreeDirVars(kp)
	rootPath := getRootPath(cfg)
	return filepath.Join(rootPath, "vars", varDir)
}

func installRequiredFiles(cfg *runner.Config, sshCfg *sshclient.Config) error {
	rootPath := getRootPath(cfg)
	for src, dest := range cfg.RequiredFiles {
		_, err := tools.RequiredOnce(cfg.Cache, []string{"host", sshCfg.Host, dest}, src, func() (interface{}, error) {
			destAbs := filepath.Join(rootPath, dest)
			err := sshutils.Upload(sshCfg, src, destAbs, cfg.Logger)
			return nil, err
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func registerVarsCreateFiles(cfg *runner.Config, client *sshclient.Client) error {
	varsPath := getVarsPath(cfg)

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	if _, err := session.Output(fmt.Sprintf("mkdir -p %s", varsPath)); err != nil {
		return err
	}
	session.Close()

	var wg sync.WaitGroup
	var errs []error

	var vars []string
	for _, vr := range cfg.RegisterVars {
		if vr.Enable {
			vars = append(vars, vr.GetFrom())
		}
	}
	vars = append(vars, "raw.stdout")

	for _, vr := range vars {
		wg.Add(1)
		go func(vr string) {
			defer wg.Done()

			session, err := client.NewSession()
			if err != nil {
				errs = append(errs, err)
				return
			}
			file := filepath.Join(varsPath, vr)
			if _, err := session.Output(fmt.Sprintf("touch %s", file)); err != nil {
				errs = append(errs, err)
				return
			}
			session.Close()

		}(vr)
	}
	wg.Wait()
	if len(errs) > 0 {
		return multierr.Combine(errs...)
	}
	return nil
}

func registerVarsRetrieve(cfg *runner.Config, client *sshclient.Client) error {
	kp := cfg.TreeKeyParts
	if len(kp) < 1 {
		return nil
	}
	dp := kp[0 : len(kp)-1]

	varsPath := getVarsPath(cfg)
	r := cfg.VarsRegistry

	var wg sync.WaitGroup
	var errs []error

	for _, vr := range cfg.RegisterVars {
		if !vr.Enable {
			continue
		}

		wg.Add(1)
		go func(vr *registry.VarDef) {
			defer wg.Done()

			var src string
			if !vr.SourceStdout {
				src = vr.GetFrom()
			} else {
				src = "raw.stdout"
			}

			session, err := client.NewSession()
			if err != nil {
				errs = append(errs, err)
				return
			}
			file := filepath.Join(varsPath, src)
			dat, err := session.Output(fmt.Sprintf("cat %s", file))
			if err != nil {
				errs = append(errs, err)
				return
			}
			session.Close()

			value := string(dat)
			value = strings.TrimSuffix(value, "\n")
			r.SetVarBySlice(dp, vr.GetFrom(), value)
		}(vr)
	}
	wg.Wait()
	if len(errs) > 0 {
		return multierr.Combine(errs...)
	}
	return nil
}

func makeWriterHeadStripperAndGetPGID(sep string, pgid *string) func(data []byte) []byte {
	sepL := len(sep)
	var enableWrite bool
	var retrievedPgid bool
	pgidSep := " PGID\n"
	pgidSepL := len(pgidSep)
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
		if enableWrite && !retrievedPgid {
			str := string(data)
			index := strings.Index(str, pgidSep)
			if index != -1 {
				retrievedPgid = true
				s := str[0:index]
				*pgid = s
				data = []byte(str[index+pgidSepL:])
			}
		}
		if len(bytes.TrimSpace(data)) == 0 {
			return make([]byte, 0)
		}
		return data
	}
}
