package mainNative

import (
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
	expect "gitlab.com/youtopia.earth/ops/snip/goexpect"
	"gitlab.com/youtopia.earth/ops/snip/plugin/middleware"
	"golang.org/x/crypto/ssh"
)

var (
	Middleware = middleware.Plugin{
		UseVars: []string{"user", "pass"},
		Apply: func(cfg *middleware.Config) (bool, error) {

			vars := cfg.MiddlewareVars
			mutableCmd := cfg.MutableCmd

			command := []string{"sudo", "--preserve-env"}

			command = append(command, "--prompt=[sudo]")

			if pass, ok := vars["pass"]; ok && pass != "" {
				mutableCmd.PrependExpect(
					&expect.BExp{R: "[sudo]"},
					&expect.BSnd{S: pass + "\n"},
				)
				command = append(command, "--stdin")
			}

			if user, ok := vars["user"]; ok && user != "" {
				command = append(command, "--user="+user)
			}

			command = append(command, "--")

			mutableCmd.Command = append(command, mutableCmd.Command...)

			f := func(iface interface{}, pPgid *string) bool {
				var err error
				switch v := iface.(type) {
				case *exec.Cmd:
					err = CloseCmd(v, cfg)
				case *ssh.Client:
					err = CloseSSH(v, cfg, pPgid)
				}
				if err != nil && err.Error() != "no such process" {
					logrus.Warnf("failed to kill: %v", err)
				}
				return true
			}
			mutableCmd.Closer = &f

			return true, nil
		},
	}
)

func CloseCmd(cmd *exec.Cmd, cfg *middleware.Config) error {
	if cmd.Process == nil || !cmd.SysProcAttr.Setpgid {
		return nil
	}
	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err != nil {
		return err
	}

	vars := cfg.MiddlewareVars
	command := []string{"sudo"}
	if _, ok := vars["pass"]; ok {
		command = append(command, "--stdin")
	}
	command = append(command, "--")
	command = append(command, "kill", "-KILL")
	command = append(command, "--")
	command = append(command, strconv.Itoa(-pgid))
	killCmd := exec.Command(command[0], command[1:]...)
	if pass, ok := vars["pass"]; ok {
		killCmd.Stdin = strings.NewReader(pass)
	}
	if err := killCmd.Run(); err != nil {
		return err
	}
	return nil
}

func CloseSSH(client *ssh.Client, cfg *middleware.Config, pPgid *string) error {
	var pgid string
	if pPgid != nil {
		pgid = *pPgid
	}
	if pgid == "" {
		return nil
	}
	session, err := client.NewSession()
	if err != nil {
		return err
	}

	vars := cfg.MiddlewareVars
	command := []string{"sudo"}
	if _, ok := vars["pass"]; ok {
		command = append(command, "--stdin")
	}
	command = append(command, "--")
	command = append(command, "kill", "-KILL")
	command = append(command, "--")
	command = append(command, "-"+pgid)

	if pass, ok := vars["pass"]; ok {
		session.Stdin = strings.NewReader(pass + "\n")
	}

	if err := session.Run(strings.Join(command, " ")); err != nil {
		return err
	}
	*pPgid = ""

	return nil
}
