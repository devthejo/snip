package middleware

type Func func(*Config, func() error) error
