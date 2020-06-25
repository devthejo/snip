package mainNative

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	expect "gitlab.com/youtopia.earth/ops/snip/goexpect"
	"gitlab.com/youtopia.earth/ops/snip/plugin/middleware"
)

var (
	Middleware = middleware.Plugin{
		UseVars: []string{"user", "pass"},
		Apply: func(cfg *middleware.Config) (bool, error) {

			vars := cfg.MiddlewareVars
			mutableCmd := cfg.MutableCmd

			command := []string{"sudo", "--preserve-env"}

			if pass, ok := vars["pass"]; ok {
				command = append(command, "--prompt=[sudo]")
				mutableCmd.PrependExpect(
					&expect.BExp{R: "[sudo]"},
					&expect.BSnd{S: pass + "\n"},
				)
				command = append(command, "--stdin")
			}

			if user, ok := vars["user"]; ok {
				command = append(command, "--user="+user)
			}

			command = append(command, "--")

			mutableCmd.Command = append(command, mutableCmd.Command...)

			f := func(iface interface{}) bool {
				switch v := iface.(type) {
				case *exec.Cmd:
					CloseCmd(v, cfg)
				}
				return true
			}
			mutableCmd.Closer = &f

			return true, nil
		},
	}
)

func CloseCmd(cmd *exec.Cmd, cfg *middleware.Config) {
	if cmd.Process == nil || !cmd.SysProcAttr.Setpgid {
		return
	}
	kill := exec.Command("sudo", "kill", "-TERM", "--", strconv.Itoa(-cmd.Process.Pid))
	if pass, ok := cfg.MiddlewareVars["pass"]; ok {
		kill.Stdin = strings.NewReader(pass)
	}
	if err := kill.Run(); err != nil {
		logrus.Warnf("failed to kill: %v", err)
	}
}
