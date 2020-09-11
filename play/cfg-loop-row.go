package play

import (
	"strconv"

	"gitlab.com/youtopia.earth/ops/snip/variable"
)

type CfgLoopRow struct {
	Name          string
	Key           string
	Index         int
	Vars          map[string]*variable.Var
	IsLoopRowItem bool
}

func CreateCfgLoopRow(index int, key string, loopSet map[string]*variable.Var) *CfgLoopRow {
	var name string
	if key == "" {
		name = "loop-index : " + strconv.Itoa(index)

	} else {
		name = "loop-set   : " + key
	}
	cfgLoopRow := &CfgLoopRow{
		Name:          name,
		Key:           key,
		Index:         index,
		Vars:          loopSet,
		IsLoopRowItem: true,
	}
	return cfgLoopRow
}
