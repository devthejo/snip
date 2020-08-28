package mainNative

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
	"gitlab.com/youtopia.earth/ops/snip/errors"
	"gitlab.com/youtopia.earth/ops/snip/plugin/loader"
	"gitlab.com/youtopia.earth/ops/snip/tools"
)

func GetFileContent(cfg *loader.Config) []byte {

	filePath := cfg.Command[0]
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
		logrus.Fatalf("file not found %v", file)
	}

	logrus.Debugf("loading file %v", file)
	source, err := ioutil.ReadFile(file)
	errors.Check(err)

	return source
}
