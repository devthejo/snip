package mainNative

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/kvz/logstreamer"
	expect "gitlab.com/youtopia.earth/ops/snip/goexpect"

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

			env := cfg.EnvMap()

			w := logger.Writer()
			defer w.Close()
			logStreamer := logstreamer.NewLogstreamer(log.New(w, "", 0), "", false)
			defer logStreamer.Close()

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

			var opts []expect.Option
			opts = append(opts, expect.Tee(logStreamer))

			// opts = append(opts, expect.Verbose(true))
			// opts = append(opts, expect.VerboseWriter(logStreamer))

			cmd := exec.CommandContext(cfg.Context, commandSlice[0], commandSlice[1:]...)
			opts = append(opts, expect.SetSysProcAttr(&syscall.SysProcAttr{Setpgid: true}))
			opts = append(opts, expect.SetEnv(tools.EnvToPairs(env)))
			e, ch, err := expect.SpawnCommand(cmd, -1, opts...)

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
					return
				}
			}()

			expected := cfg.Expect
			if cfg.Stdin != nil {
				b, err := ioutil.ReadAll(cfg.Stdin)
				if err != nil {
					return err
				}
				expected = append(expected, &expect.BSnd{S: string(b)})
			}
			e.ExpectBatch(expected, -1)

			return <-ch
		},
	}
)
