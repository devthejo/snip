package parse

type Env struct {
	Mapping func(string) (string, bool)
}

func (e Env) Get(name string) string {
	v, _ := e.Lookup(name)
	return v
}

func (e Env) Has(name string) bool {
	_, ok := e.Lookup(name)
	return ok
}

func (e Env) Lookup(name string) (string, bool) {
	return e.Mapping(name)
}
