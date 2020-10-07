package play

type GlobalRunCtx struct {
	RunReport      *RunReport
	SkippingState  *bool
	CurrentTreeKey string
}

func CreateGlobalRunCtx() *GlobalRunCtx {
	ctx := &GlobalRunCtx{}
	ctx.RunReport = &RunReport{}
	return ctx
}
