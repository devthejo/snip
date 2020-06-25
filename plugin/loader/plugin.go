package loader

type Plugin struct {
	Load  func(*Config) error
	Check func([]string) bool
}
