package runner

type Plugin struct {
	UseVars []string
	Run     func(*Config) error
}
