package play

import "strconv"

type LoopRow struct {
	Name          string
	Key           string
	Index         int
	Vars          map[string]*Var
	ParentPlay    *Play
	Play          interface{}
	IsLoopRowItem bool
}

func (l *LoopRow) GetKey() string {
	k := l.Key
	if k == "" {
		k = strconv.Itoa(l.Index)
	}
	return k
}
