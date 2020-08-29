package loader

import (
	"path"
	// "path/filepath"
	// "strings"
)

func SnippetDefaultTitle(key string, cfg *Config) string {
	title := key

	if len(title) > 0 && (title[0:2] == "./" || title[0:3] == "../") {
		if cfg.ParentBuildFile != "" {
			d := ResolveSnippetDir(cfg.ParentBuildFile)
			title = path.Join(d, title)
		}
	}

	// title = strings.TrimSuffix(title, filepath.Ext(title))
	// title = strings.ReplaceAll(title, "-", " ")
	// title = strings.ReplaceAll(title, "/", " ")

	title = "snippet: " + title
	return title
}
