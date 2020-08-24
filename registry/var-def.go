package registry

type VarDef struct {
	To           string
	From         string
	Source       string
	SourceStdout bool
	Enable       bool
	Persist      bool
}

func (v *VarDef) GetSource() string {
	var src string
	if v.Source != "" {
		src = v.Source
	} else {
		src = v.GetFrom()
	}
	if src[0:1] == "@" {
		src = "_" + src[1:]
	}
	return src
}

func (v *VarDef) GetFrom() string {
	var src string
	if v.From != "" {
		src = v.From
	}
	src = v.To
	return src
}

func (v *VarDef) GetTo() string {
	return v.To
}
