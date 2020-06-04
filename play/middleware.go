package play

type Middleware struct {
	Bin string
	Key string
}

func BuildMiddlewaresMap(app App) map[string]*Middleware {
	cfg := app.GetConfig()
	middlewaresMap := make(map[string]*Middleware)

	for key, val := range cfg.Middlewares {
		middlewaresMap[key] = &Middleware{
			Key: key,
			Bin: val,
		}
	}

	return middlewaresMap
}
