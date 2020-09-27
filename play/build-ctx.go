package play

import (
	"gitlab.com/ytopia/ops/snip/plugin/runner"
	"gitlab.com/ytopia/ops/snip/variable"
)

type LoadedSnippet struct {
	requiredBy []string
}

type BuildCtx struct {
	Parent                          *BuildCtx
	LoadedSnippets                  map[string]*LoadedSnippet
	LoadedSnippetsUpstream          map[string]*LoadedSnippet
	LoadedSnippetsDownstream        map[string]*LoadedSnippet
	LoadedSnippetsDownstreamParents []map[string]*LoadedSnippet
	DefaultRunner                   *runner.Runner
	DefaultVars                     map[string]*variable.Var
}

func CreateBuildCtx() *BuildCtx {
	buildCtx := &BuildCtx{
		LoadedSnippets: make(map[string]*LoadedSnippet),
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

	loadedSnippets := make(map[string]*LoadedSnippet)
	for k, v := range parentBuildCtx.LoadedSnippets {
		loadedSnippets[k] = v
	}

	loadedSnippetsDownstream := make(map[string]*LoadedSnippet)

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
	loadedSnippet := &LoadedSnippet{}
	buildCtx.LoadedSnippets[snippet] = loadedSnippet
	for _, v := range buildCtx.LoadedSnippetsDownstreamParents {
		v[snippet] = loadedSnippet
	}
}
