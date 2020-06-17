package middleware

type Middleware struct {
	Apply func(*Config) (bool, error)
}
