package mainNative

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	// "github.com/sirupsen/logrus"

	"github.com/devthejo/snip/errors"
	"github.com/devthejo/snip/plugin/processor"
	loaderMarkdownMod "github.com/devthejo/snip/plugins-native/loaders/markdown/mod"
	"github.com/devthejo/snip/tools"
	"github.com/sirupsen/logrus"
)

var (
	Mod = loaderMarkdownMod.Plugin{
		Mod: func(modCfg *loaderMarkdownMod.Config) error {
			cfg := modCfg.LoaderConfig
			codeBlock := modCfg.CodeBlock
			args := modCfg.Args

			usr, _ := user.Current()
			buildDir := "build/snippets"
			rootPath := filepath.Join(usr.HomeDir, ".snip", cfg.AppConfig.DeploymentName)

			mdpath := cfg.Command[0]
			snippetPath := mdpath + "_scp_" + strconv.Itoa(codeBlock.Index)

			file := filepath.Join(buildDir, snippetPath)

			fileAbs := filepath.Join(rootPath, file)
			dir := filepath.Dir(fileAbs)
			os.MkdirAll(dir, os.ModePerm)
			err := ioutil.WriteFile(fileAbs, []byte(codeBlock.Content), 0644)
			errors.Check(err)

			cfg.RequiredFiles.Set(file, fileAbs)

			cfg.RequiredFilesSrcProcessors[fileAbs] = append(cfg.RequiredFilesSrcProcessors[fileAbs], codeBlock.Processors...)
			codeBlock.Processors = nil

			fileAbsRemote := filepath.Join("${SNIP_SNIPPETS_PATH}", snippetPath)

			targetFile := args[0]
			codeBlock.Lang = "sh"
			codeBlock.Content = "mkdir -p " + path.Dir(targetFile) + "\n"
			codeBlock.Content += "mv " + fileAbsRemote + " " + targetFile + "\n"

			sumProcessor := func(processorCfg *processor.Config, src *string) error {
				b, err := ioutil.ReadFile(*src)
				if err != nil {
					return err
				}
				sum := fmt.Sprintf("%x", sha256.Sum256(b))
				key := "SNIP_SHA256_" + tools.KeyEnv(targetFile)
				processorCfg.RunVars.SetValueString(key, sum)
				return nil
			}

			cfg.RequiredFilesSrcProcessors[fileAbs] = append(cfg.RequiredFilesSrcProcessors[fileAbs], sumProcessor)

			var chmod string
			var chown string
			for _, arg := range args[1:] {
				if strings.TrimSpace(arg) == "" {
					continue
				}
				x := strings.SplitN(arg, "=", 2)
				k := x[0]
				var v string
				if len(x) > 1 {
					v = x[1]
				}
				switch k {
				case "chmod":
					chmod = v
				case "chown":
					chown = v
				default:
					logrus.Errorf("unkown arg %v for scp markdown mod", k)
				}
			}

			if chmod != "" {
				codeBlock.Content += "chmod " + chmod + " " + targetFile + "\n"
			}

			if chown != "" {
				codeBlock.Content += "chown " + chown + " " + targetFile + "\n"
			}

			return nil
		},
	}
)
