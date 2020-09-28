package play

import (
	"strconv"

	"gitlab.com/ytopia/ops/snip/variable"
)

type LoopRow struct {
	Name          string
	Key           string
	Index         int
	Vars          map[string]*variable.Var
	ParentPlay    *Play
	Play          interface{}
	IsLoopRowItem bool
	HasChk        bool
	PreChk        *Chk
	PostChk       *Chk
	RunCtx        *RunCtx
}

func (l *LoopRow) GetKey() string {
	k := l.Key
	if k == "" {
		k = strconv.Itoa(l.Index)
	}
	return k
}
