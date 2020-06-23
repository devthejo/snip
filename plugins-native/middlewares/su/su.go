package mainNative

import (
	"os/exec"
	"strings"
	// "strings"

	shellquote "github.com/kballard/go-shellquote"
	expect "gitlab.com/youtopia.earth/ops/snip/goexpect"

	"gitlab.com/youtopia.earth/ops/snip/plugin"
	"gitlab.com/youtopia.earth/ops/snip/plugin/middleware"
)

// inspired from https://github.com/ansible/ansible/blob/devel/lib/ansible/plugins/become/su.py
var promptL10N = []string{
	"Password",
	"암호",
	"パスワード",
	"Adgangskode",
	"Contraseña",
	"Contrasenya",
	"Hasło",
	"Heslo",
	"Jelszó",
	"Lösenord",
	"Mật khẩu",
	"Mot de passe",
	"Parola",
	"Parool",
	"Pasahitza",
	"Passord",
	"Passwort",
	"Salasana",
	"Sandi",
	"Senha",
	"Wachtwoord",
	"ססמה",
	"Лозинка",
	"Парола",
	"Пароль",
	"गुप्तशब्द",
	"शब्दकूट",
	"సంకేతపదము",
	"හස්පදය",
	"密码",
	"密碼",
	"口令",
}

var (
	Middleware = middleware.Middleware{
		Apply: func(cfg *middleware.Config) (bool, error) {

			mutableCmd := cfg.MutableCmd

			command := []string{"su", "--preserve-env"}

			if pass, ok := mutableCmd.Vars["@SU_PASS"]; ok {
				mutableCmd.PrependExpect(
					&expect.BExp{R: strings.Join(promptL10N, "|") + " ?(:|：) ?"},
					&expect.BSnd{S: pass + "\n"},
				)
			}

			if user, ok := mutableCmd.Vars["@SU_USER"]; ok {
				command = append(command, user)
			}

			originalCommand := mutableCmd.Command
			mutableCmd.Command = append(command, `--command="`+shellquote.Join(originalCommand...)+`"`)

			f := func(iface interface{}) bool {
				switch v := iface.(type) {
				case *exec.Cmd:
					CloseCmd(v, mutableCmd)
				}
				return true
			}
			mutableCmd.Closer = &f

			return true, nil
		},
	}
)

func CloseCmd(cmd *exec.Cmd, mutableCmd *plugin.MutableCmd) {
	// if cmd.Process == nil {
	// 	return
	// }
	// killSlice := []string{"su"}
	// if user, ok := mutableCmd.Vars["@SU_USER"]; ok {
	// 	killSlice = append(killSlice, user)
	// }
	// killSlice = append(killSlice, "-c", "kill", "-TERM", "--", strconv.Itoa(-cmd.Process.Pid))
	// kill := exec.Command(killSlice[0], killSlice[1:]...)
	// // kill.Stdin = strings.NewReader(password)
	// if err := kill.Run(); err != nil {
	// 	logrus.Warnf("failed to kill: %v", err)
	// }
}
