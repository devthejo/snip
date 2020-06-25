package runner

type Plugin struct {
	Run func(*Config) error
}
