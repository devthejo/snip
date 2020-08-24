package mod

type Plugin struct {
	Mod func(*Config) error
}
