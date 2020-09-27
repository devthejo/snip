package middleware

import (
	expect "gitlab.com/ytopia/ops/snip/goexpect"
	snipplugin "gitlab.com/ytopia/ops/snip/plugin"
	"gitlab.com/ytopia/ops/snip/plugin/runner"
	"gitlab.com/ytopia/ops/snip/plugin/processor"
)

type MutableCmd struct {
	AppConfig *snipplugin.AppConfig

	Command []string
	Vars    map[string]string

	OriginalCommand []string
	OriginalVars    map[string]string

	RequiredFiles           map[string]string
	RequiredFilesSrcProcessors map[string][]func(*processor.Config, *string) error
	Expect                  []expect.Batcher
	Runner                  *runner.Runner

	Dir string

	Closer *func(interface{}, *string) bool
}

func (cmd *MutableCmd) PrependExpect(b ...expect.Batcher) {
	cmd.Expect = append(b, cmd.Expect...)
}

func (cmd *MutableCmd) AppendExpect(b ...expect.Batcher) {
	cmd.Expect = append(cmd.Expect, b...)
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
