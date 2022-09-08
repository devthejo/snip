package loader

import (
	"path"
	"path/filepath"

	"github.com/devthejo/snip/tools"
)

func ResolveSnippetFile(file string) (bool, string, error) {
	exists, err := tools.FileExists(file)
	if err != nil {
		return exists, file, err
	}

	if !exists {
		extension := filepath.Ext(file)
		name := file[0 : len(file)-len(extension)]
		file = name + "/index" + extension
		exists, err = tools.FileExists(file)
		if err != nil {
			return exists, file, err
		}
	}

	return exists, file, err
}

func ResolveSnippetDir(file string) string {
	_, file, _ = ResolveSnippetFile(file)
	return path.Dir(file)
}
