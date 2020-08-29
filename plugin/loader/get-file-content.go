package loader

import (
	"io/ioutil"
	"path"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gitlab.com/youtopia.earth/ops/snip/errors"
	"gitlab.com/youtopia.earth/ops/snip/tools"
)

func resolveFile(file string) (bool, string, error) {
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

func GetFileContent(cfg *Config, filePath string) []byte {
	appCfg := cfg.AppConfig

	var file string
	if len(filePath) > 0 && filePath[0:1] != "/" {
		dir := appCfg.SnippetsDir
		if filePath[0:2] == "./" || filePath[0:3] == "../" {
			parentBuildFile := cfg.ParentBuildFile
			if parentBuildFile != "" {
				_, parentBuildFile, _ = resolveFile(parentBuildFile)
				ctxDir := path.Dir(parentBuildFile)
				dir = path.Join(dir, ctxDir)
			}
		}
		file = path.Join(dir, filePath)
	} else {
		file = filePath
	}

	var err error
	var exists bool
	exists, file, err = resolveFile(file)
	errors.Check(err)


	if !exists {
		logrus.Fatalf("snippet not found %v", filePath)
	}

	logrus.Debugf("loading file %v", file)
	source, err := ioutil.ReadFile(file)
	errors.Check(err)

	return source
}
