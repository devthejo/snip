package middleware

import (
	expect "github.com/devthejo/snip/goexpect"
	snipplugin "github.com/devthejo/snip/plugin"
	"github.com/devthejo/snip/plugin/processor"
	"github.com/devthejo/snip/plugin/runner"
)

type MutableCmd struct {
	AppConfig *snipplugin.AppConfig

	Command []string
	// Vars    map[string]string

	OriginalCommand []string

	RequiredFiles              map[string]string
	RequiredFilesSrcProcessors map[string][]func(*processor.Config, *string) error
	Expect                     []expect.Batcher
	Runner                     *runner.Runner

	Dir string

	Closer *func(interface{}, *string) bool
}

func (cmd *MutableCmd) PrependExpect(b ...expect.Batcher) {
	cmd.Expect = append(b, cmd.Expect...)
}

func (cmd *MutableCmd) AppendExpect(b ...expect.Batcher) {
	cmd.Expect = append(cmd.Expect, b...)
}
