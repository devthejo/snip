package play

import (
	"strconv"
)

type CfgLoopRow struct {
	Name          string
	Key           string
	Index         int
	Vars          map[string]*Var
	IsLoopRowItem bool
}

func CreateCfgLoopRow(index int, key string, loopSet map[string]*Var) *CfgLoopRow {
	var name string
	if key == "" {
		name = "loop-index   : " + strconv.Itoa(index)

	} else {
		name = "loop-set     : " + key
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
