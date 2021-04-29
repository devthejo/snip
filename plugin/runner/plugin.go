package runner

type Plugin struct {
	UseVars     []string
	Run         func(*Config) error
	UpUse       func(*Config) error
	DownPersist func(*Config) error
	GetRootPath func(*Config) string
}
