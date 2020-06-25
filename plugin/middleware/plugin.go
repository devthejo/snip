package middleware

type Plugin struct {
	Apply func(*Config) (bool, error)
}
