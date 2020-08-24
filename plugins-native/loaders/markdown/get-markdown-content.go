package mainNative

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
	"gitlab.com/youtopia.earth/ops/snip/errors"
	"gitlab.com/youtopia.earth/ops/snip/plugin/loader"
	"gitlab.com/youtopia.earth/ops/snip/tools"
)

func GetMarkdownContent(cfg *loader.Config) []byte {

	mdPath := cfg.Command[0]
	appCfg := cfg.AppConfig

	var file string
	if len(mdPath) > 0 && mdPath[0:1] != "/" {
		file = appCfg.SnippetsDir + "/" + mdPath
	} else {
		file = mdPath
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
