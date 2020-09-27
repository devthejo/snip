package loader

import (
	"io/ioutil"
	"path"

	"github.com/sirupsen/logrus"
	"gitlab.com/ytopia/ops/snip/errors"
)

func GetFileContent(cfg *Config, filePath string) []byte {
	appCfg := cfg.AppConfig

	var file string
	if len(filePath) > 0 && filePath[0:1] != "/" {
		dir := appCfg.SnippetsDir
		if filePath[0:2] == "./" || filePath[0:3] == "../" {
			parentBuildFile := cfg.ParentBuildFile
			if parentBuildFile != "" {
				sDir := ResolveSnippetDir(parentBuildFile)
				dir = path.Join(dir, sDir)
			}
		}
		file = path.Join(dir, filePath)
	} else {
		file = filePath
	}

	var err error
	var exists bool
	exists, file, err = ResolveSnippetFile(file)
	errors.Check(err)

	if !exists {
		logrus.Fatalf("snippet not found %v", filePath)
	}

	logrus.Debugf("loading file %v", file)
	source, err := ioutil.ReadFile(file)
	errors.Check(err)

	return source
}
