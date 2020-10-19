package mainNative

import (
	"strings"
	"path/filepath"
	"io/ioutil"

	"github.com/sirupsen/logrus"
	"gitlab.com/golang-commonmark/markdown"

	"gitlab.com/ytopia/ops/snip/plugin/loader"
	"gitlab.com/ytopia/ops/snip/plugins-native/loaders/markdown/blocks"
)

func handleInstruction(t string, args []string, parseMdLoopParams *ParseMdLoopParams, cfg *loader.Config, codeBlocks *[]*blocks.Code, snippetPath string) {
	switch t {
	case "ignore-next":
		parseMdLoopParams.ignoreCodeOnce = true
	case "mod":
		if len(args) > 0 {
			parseMdLoopParams.handleModsOnce = true
			parseMdLoopParams.handleModsArgs = append(parseMdLoopParams.handleModsArgs, args)
		}
	case "include":
		argN := 0
		var lg string
		var file string
		for _, arg := range args {
			if strings.TrimSpace(arg) == "" {
				continue
			}
			x := strings.SplitN(arg, "=", 2)
			k := x[0]
			var v string
			if len(x) > 1 {
				v = x[1]
				switch k {
				case "lg":
					lg = v
				case "lang":
					lg = v
				case "ext":
					lg = v
				case "file":
					file = v
				default:
					logrus.Errorf("unkown arg %v for include param", k)
				}
			} else {
				switch argN {
				case 0:
					file = arg
				default:
					logrus.Errorf("unkown arg %v for include param", argN)
				}
				argN++
			}
		}

		if lg == "" {
			if ext := filepath.Ext(file); len(ext) > 1 {
				lg = ext[1:]
			}
		}

		if strings.HasPrefix(file, "./") {
			dirPath := filepath.Dir(snippetPath)
			file = filepath.Join(cfg.AppConfig.SnippetsDir, dirPath, file[2:])
		}

		b, err := ioutil.ReadFile(file)
		if err != nil {
			logrus.Fatalf(`file not found in markdown include: "%v"`, file)
		}
		t := &markdown.Fence{
			Params: lg,
			Content: string(b),
		}
		handleToken(cfg, t, codeBlocks, parseMdLoopParams, snippetPath)
	default:
		logrus.Fatalf("unkown snip instruction %v", t)
	}
}
