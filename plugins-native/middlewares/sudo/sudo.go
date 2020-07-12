package mainNative

import (
	"strings"

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
			command = append(command, "--prompt=[sudo]")
			if user, ok := vars["user"]; ok && user != "" {
				command = append(command, "--user="+user)
			}

			if _, ok := vars["pass"]; ok {
				command = append(command, "--stdin")
			}

			command = append(command, "--")
			command = append(command, "/bin/sh")
			// command = append(command, "su")
			mutableCmd.ExpectBeforeCommand = append(mutableCmd.ExpectBeforeCommand, &expect.BSnd{S: strings.Join(command, " ") + "\n"})

			if pass, ok := vars["pass"]; ok {
				mutableCmd.ExpectBeforeCommand = append(mutableCmd.ExpectBeforeCommand, &expect.BExp{R: "[sudo]"})
				mutableCmd.ExpectBeforeCommand = append(mutableCmd.ExpectBeforeCommand, &expect.BSnd{S: pass + "\n"})
			}

			// command := []string{"sudo", "--preserve-env"}
			//
			// command = append(command, "--prompt=[sudo]")
			//
			// if pass, ok := vars["pass"]; ok {
			// 	mutableCmd.PrependExpect(
			// 		&expect.BExp{R: "[sudo]"},
			// 		&expect.BSnd{S: pass + "\n"},
			// 	)
			// 	command = append(command, "--stdin")
			// }
			//
			// if user, ok := vars["user"]; ok && user != "" {
			// 	command = append(command, "--user="+user)
			// }
			//
			// command = append(command, "--")
			//
			// mutableCmd.Command = append(command, mutableCmd.Command...)

			return true, nil
		},
	}
)
