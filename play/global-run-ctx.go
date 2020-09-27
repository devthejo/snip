package play

type GlobalRunCtx struct {
	RunReport        *RunReport
	SkippingState *bool
}

func CreateGlobalRunCtx() *GlobalRunCtx {
	ctx := &GlobalRunCtx{}
	ctx.RunReport = &RunReport{}
	return ctx
}
