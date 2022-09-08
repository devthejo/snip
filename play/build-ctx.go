package play

import (
	"github.com/devthejo/snip/plugin/runner"
	"github.com/devthejo/snip/variable"
)

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

func (buildCtx *BuildCtx) LoadedSnippetKey(scope string, snippet string) string {
	return scope + ":" + snippet
}
func (buildCtx *BuildCtx) RegisterLoadedSnippet(scope string, snippet string) {
	key := buildCtx.LoadedSnippetKey(scope, snippet)
	if _, hasKey := buildCtx.LoadedSnippets[key]; !hasKey {
		buildCtx.LoadedSnippets[key] = CreateLoadedSnippet()
	}
	loadedSnippet := buildCtx.LoadedSnippets[key]
	for _, v := range buildCtx.LoadedSnippetsDownstreamParents {
		if _, hasKey := v[key]; !hasKey {
			v[key] = loadedSnippet
		}
	}
}
