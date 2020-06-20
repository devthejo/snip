package mainNative

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"gitlab.com/youtopia.earth/ops/snip/plugin"
	"gitlab.com/youtopia.earth/ops/snip/plugin/middleware"
	"golang.org/x/crypto/ssh"
)

var (
	Middleware = middleware.Middleware{
		Apply: func(cfg *middleware.Config) (bool, error) {

			mutableCmd := cfg.MutableCmd

			command := []string{"sudo", "--preserve-env"}

			if pass, ok := mutableCmd.Vars["@SUDO_PASS"]; ok {
				mutableCmd.Stdin = strings.NewReader(pass + "\n")
				command = append(command, "--stdin")
			}

			if user, ok := mutableCmd.Vars["@SUDO_USER"]; ok {
				command = append(command, "--user="+user)
			}

			command = append(command, "--")

			mutableCmd.Command = append(command, mutableCmd.Command...)

			f := func(iface interface{}) bool {
				switch v := iface.(type) {
				case *exec.Cmd:
					CloseCmd(v, mutableCmd)
				case *ssh.Session:
					CloseSSH(v, mutableCmd)
				}
				return true
			}
			mutableCmd.Closer = &f

			return true, nil
		},
	}
)

func CloseCmd(cmd *exec.Cmd, mutableCmd *plugin.MutableCmd) {
	if cmd.Process == nil {
		return
	}
	kill := exec.Command("sudo", "kill", "-TERM", "--", strconv.Itoa(-cmd.Process.Pid))
	kill.Stdin = mutableCmd.Stdin
	if err := kill.Run(); err != nil {
		logrus.Warn(err)
	}
}

func CloseSSH(session *ssh.Session, mutableCmd *plugin.MutableCmd) {

}
