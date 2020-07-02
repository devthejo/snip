package play

type BuildCtx struct {
	LoadedSnippets                  map[string]bool
	LoadedSnippetsUpstream          map[string]bool
	LoadedSnippetsDownstream        map[string]bool
	LoadedSnippetsDownstreamParents []map[string]bool
}

func CreateBuildCtx() *BuildCtx {
	buildCtx := &BuildCtx{
		LoadedSnippets: make(map[string]bool),
	}
	return buildCtx
}

func (buildCtx *BuildCtx) RegisterLoadedSnippet(snippet string) {
	buildCtx.LoadedSnippets[snippet] = true
	buildCtx.LoadedSnippetsDownstream[snippet] = true
	for _, v := range buildCtx.LoadedSnippetsDownstreamParents {
		v[snippet] = true
	}
}
