package play

import (
	"strconv"

	"github.com/devthejo/snip/variable"
)

type LoopRow struct {
	Name          string
	Key           string
	Index         int
	Prefix        string
	Vars          map[string]*variable.Var
	ParentPlay    *Play
	Play          interface{}
	IsLoopRowItem bool
	PreChk        *Chk
	PostChk       *Chk
	RunVars       *RunVars
}

func (l *LoopRow) GetKey() string {
	return l.Key + "." + strconv.Itoa(l.Index)
}
