package variable

type FromType int

const (
	FromValue FromType = iota
	FromVar
	FromFile
)

type RunVar struct {
	FromType  FromType
	FromParam string
}

func CreateRunVar() *RunVar {
	runVar := &RunVar{
		FromType: FromValue,
	}
	return runVar
}

func (runVar *RunVar) GetValue() string {
	var r string
	switch runVar.FromType {
	case FromValue:
		r = runVar.FromParam
	case FromVar:

	case FromFile:

	}
	return r
}
