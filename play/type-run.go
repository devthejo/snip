package play

import (
	"strings"

	"github.com/sirupsen/logrus"
	"gitlab.com/youtopia.earth/ops/snip/tools"
)

type Run struct {
	Cmd  *Cmd
	Play *Play

	Vars map[string]string

	State StateType

	Log *logrus.Logger
}

func CreateRun(cmd *Cmd) *Run {
	r := &Run{
		Cmd:  cmd,
		Play: cmd.Play,
	}

	// r.Log = logrus.WithFields(logrus.Fields{
	//   "key": ,
	// })

	return r
}

func (r *Run) Exec() {
	logrus.Debugf(strings.Repeat("  ", r.Play.Depth+2)+" vars: %v", tools.JsonEncode(r.Vars))

}
