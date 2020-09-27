package play

type LoadedSnippet struct {
	requiredByDependencies map[string]bool
	requiredByPostInstall  map[string]bool
}

func CreateLoadedSnippet() *LoadedSnippet {
	loadedSnippet := &LoadedSnippet{}
	loadedSnippet.requiredByDependencies = make(map[string]bool)
	loadedSnippet.requiredByPostInstall = make(map[string]bool)
	return loadedSnippet
}
