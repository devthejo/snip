package loader

import (
	"io/ioutil"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gitlab.com/youtopia.earth/ops/snip/errors"
	"gitlab.com/youtopia.earth/ops/snip/tools"
)

func GetFileContent(cfg *Config, filePath string) []byte {
	appCfg := cfg.AppConfig

	var file string
	if len(filePath) > 0 && filePath[0:1] != "/" {
		file = appCfg.SnippetsDir + "/" + filePath
	} else {
		file = filePath
	}

	exists, err := tools.FileExists(file)
	errors.Check(err)

	if !exists {
		extension := filepath.Ext(file)
		name := file[0 : len(file)-len(extension)]
		file = name + "/index" + extension
		exists, err = tools.FileExists(file)
		errors.Check(err)
	}

	if !exists {
		logrus.Fatalf("snippet not found %v", filePath)
	}

	logrus.Debugf("loading file %v", file)
	source, err := ioutil.ReadFile(file)
	errors.Check(err)

	return source
}
