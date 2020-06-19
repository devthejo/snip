package plugin

import (
	"io"

	expect "gitlab.com/youtopia.earth/ops/snip/goexpect"
)

type MutableCmd struct {
	AppConfig *AppConfig

	Command []string
	Vars    map[string]string

	OriginalCommand []string
	OriginalVars    map[string]string

	RequiredFiles map[string]string
	Expect        []expect.Batcher
	Stdin         io.Reader
	Closer        *func(interface{}) bool
	Runner        string
}

func (cmd *MutableCmd) EnvMap() map[string]string {
	m := make(map[string]string)
	for k, v := range cmd.Vars {
		if k[0:1] != "@" {
			m[k] = v
		}
	}
	return m
}
