package play

type GlobalRunCtx struct {
	RunReport      *RunReport
	SkippingState  *bool
	CurrentTreeKey string
	NoSkipTreeKeys map[string]bool
}

func CreateGlobalRunCtx() *GlobalRunCtx {
	ctx := &GlobalRunCtx{}
	ctx.RunReport = &RunReport{}
	ctx.NoSkipTreeKeys = make(map[string]bool)
	return ctx
}
