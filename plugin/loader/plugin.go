package loader

type Plugin struct {
	UseVars []string
	Load    func(*Config) error
	Check   func(*Config) bool
}
