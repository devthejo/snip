package middleware

type Func func(mutableCmd *MutableCmd, next func() error) error
