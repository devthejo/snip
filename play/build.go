package play

func BuildConfig(app App) *CfgPlay {
	buildCtx := CreateBuildCtx()

	cfg := app.GetConfig()

	cfgPlay := CreateCfgPlay(app, cfg.Play, nil, buildCtx)

	return cfgPlay
}

func BuildPlay(config *CfgPlay) *Play {
	return config.BuildRoot()
}
