package loader

type Plugin struct {
	UseVars  []string
	Load     func(*Config) error
	PostLoad func(*Config) error
	Check    func(*Config) bool
}
