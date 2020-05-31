package play

type Cmd struct {
	Parent interface{}
	CfgCmd *CfgCmd

	Command string
	Args    []string
	Vars    map[string]string

	IsMD bool
}

func (cmd *Cmd) Run() {

	// logrus.Debugf(strings.Repeat("  ", cmd.Parent.Depth+2)+" vars: %v", tools.JsonEncode(cmd.Vars))

}
