package play

import (
	"strconv"
	"strings"

	"github.com/devthejo/snip/variable"
)

type CfgLoopRow struct {
	Name          string
	Key           string
	Index         int
	Vars          map[string]*variable.Var
	Prefix        string
	IsLoopRowItem bool
}

func CreateCfgLoopRow(index int, key string, vars map[string]*variable.Var, prefix string) *CfgLoopRow {
	var name string
	if key == "" {
		name = "loop-index : " + strconv.Itoa(index)

	} else {
		name = "loop-set   : " + key
	}

	if prefix != "" {
		prefix = strings.ToUpper(prefix) + "_"
	}

	cfgLoopRow := &CfgLoopRow{
		Name:          name,
		Key:           key,
		Index:         index,
		Vars:          vars,
		Prefix:        prefix,
		IsLoopRowItem: true,
	}
	return cfgLoopRow
}
