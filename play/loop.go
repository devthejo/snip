package play

type Loop struct {
	Name       string
	Vars       map[string]*Var
	Play       interface{}
	IsLoopItem bool
}
