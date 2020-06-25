package middleware

type Plugin struct {
	UseVars []string
	Apply   func(*Config) (bool, error)
}
