package middleware

type MutableCmd struct {
	Command         string
	Args            []string
	Vars            map[string]string
	OriginalCommand string
	OriginalArgs    []string
	OriginalVars    map[string]string
}
