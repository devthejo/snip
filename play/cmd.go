package play

import (
	"strings"

	"github.com/sirupsen/logrus"
	"gitlab.com/youtopia.earth/ops/snip/tools"
)

type Cmd struct {
	ParentLoopRow *LoopRow
	CfgCmd        *CfgCmd

	Command string
	Args    []string
	Vars    map[string]string

	IsMD bool

	Logger *logrus.Entry
	Depth  int
}

func (cmd *Cmd) GetLogKey() string {
	parts := cmd.GetLogKeyParts()
	return strings.Join(parts, "|")

}
func (cmd *Cmd) GetLogKeyParts() []string {
	var parts []string
	var parent interface{}
	parent = cmd.ParentLoopRow
	for {
		var part string
		switch p := parent.(type) {
		case *LoopRow:
			if p == nil {
				parent = nil
				break
			}
			part = p.GetKey()
			parent = p.ParentPlay
		case *Play:
			if p == nil {
				parent = nil
				break
			}
			part = p.GetKey()
			parent = p.ParentLoopRow
		case nil:
			parent = nil
		}
		if parent == nil {
			break
		}
		parts = append([]string{part}, parts...)
	}
	return parts
}

func (cmd *Cmd) Run() {

	cmd.Logger.Info("Hello")
	cmd.Logger.Debugf(strings.Repeat("  ", cmd.Depth+2)+" vars: %v", tools.JsonEncode(cmd.Vars))

}
