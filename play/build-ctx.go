package play

type BuildCtx struct {
	Parent                          *BuildCtx
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

func CreateNextBuildCtx(prevBuildCtx *BuildCtx) *BuildCtx {
	var parentBuildCtx *BuildCtx
	if prevBuildCtx.Parent != nil {
		parentBuildCtx = prevBuildCtx.Parent
	} else {
		parentBuildCtx = prevBuildCtx
	}

	loadedSnippets := make(map[string]bool)
	for k, v := range parentBuildCtx.LoadedSnippets {
		loadedSnippets[k] = v
	}

	loadedSnippetsDownstream := make(map[string]bool)

	buildCtx := &BuildCtx{
		Parent:                          parentBuildCtx,
		LoadedSnippets:                  parentBuildCtx.LoadedSnippets,
		LoadedSnippetsUpstream:          loadedSnippets,
		LoadedSnippetsDownstream:        loadedSnippetsDownstream,
		LoadedSnippetsDownstreamParents: append(prevBuildCtx.LoadedSnippetsDownstreamParents, loadedSnippetsDownstream),
	}

	return buildCtx
}

func (buildCtx *BuildCtx) RegisterLoadedSnippet(snippet string) {
	buildCtx.LoadedSnippets[snippet] = true
	for _, v := range buildCtx.LoadedSnippetsDownstreamParents {
		v[snippet] = true
	}
}
