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
	return runVar.FromParam
}
