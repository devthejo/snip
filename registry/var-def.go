package registry

type VarDef struct {
	To           string
	From         string
	Source       string
	SourceOutput bool
	Enable       bool
	Persist      bool
}

func (v *VarDef) GetSource() string {
	if v.Source != "" {
		return v.Source
	}
	return v.GetFrom()
}

func (v *VarDef) GetFrom() string {
	if v.From != "" {
		return v.From
	}
	return v.To
}

func (v *VarDef) GetTo() string {
	return v.To
}
