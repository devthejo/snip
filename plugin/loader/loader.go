package loader

type Loader struct {
	Load  func(*Config) error
	Check func([]string) bool
}
