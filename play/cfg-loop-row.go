package play

import (
	"strconv"

	"gitlab.com/ytopia/ops/snip/variable"
)

type CfgLoopRow struct {
	Name          string
	Key           string
	Index         int
	Vars          map[string]*variable.Var
	IsLoopRowItem bool
}

func CreateCfgLoopRow(index int, key string, vars map[string]*variable.Var) *CfgLoopRow {
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
		Vars:          vars,
		IsLoopRowItem: true,
	}
	return cfgLoopRow
}
