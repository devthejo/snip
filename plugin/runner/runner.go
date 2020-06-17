package runner

type Runner struct {
	Run func(*Config) error
}
