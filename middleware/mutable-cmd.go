package middleware

type MutableCmd struct {
	Command string
	Args    []string
	Vars    map[string]string
}
