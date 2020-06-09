package middleware

type MutableCmd struct {
	Command         string
	Args            []string
	Vars            map[string]string
	OriginalCommand string
	OriginalArgs    []string
	OriginalVars    map[string]string
}

func (cmd *MutableCmd) EnvMap() map[string]string {
	m := make(map[string]string)
	for k, v := range cmd.Vars {
		if k[0:1] != "@" {
			m[k] = v
		}
	}
	return m
}
