package mainNative

import (
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"crypto/sha256"
	"fmt"

	// "github.com/sirupsen/logrus"

	"gitlab.com/youtopia.earth/ops/snip/errors"
	"gitlab.com/youtopia.earth/ops/snip/tools"
	"gitlab.com/youtopia.earth/ops/snip/plugin/processor"
	loaderMardownMod "gitlab.com/youtopia.earth/ops/snip/plugins-native/loaders/markdown/mod"
)

var (
	Mod = loaderMardownMod.Plugin{
		Mod: func(modCfg *loaderMardownMod.Config) error {
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

			cfg.RequiredFiles[file] = fileAbs

			cfg.RequiredFilesSrcProcessors[fileAbs] = append(cfg.RequiredFilesSrcProcessors[fileAbs], codeBlock.Processors...)
			codeBlock.Processors = nil

			fileAbsRemote := filepath.Join("${SNIP_SNIPPETS_PATH}", snippetPath)

			targetFile := args[0]
			codeBlock.Lang = "sh"
			codeBlock.Content = "mkdir -p " + path.Dir(targetFile) + "\n"
			codeBlock.Content += "mv " + fileAbsRemote + " " + targetFile

			sumProcessor := func(processorCfg *processor.Config, src *string) ( error) {
				b, err := ioutil.ReadFile(*src)
				if err != nil {
					return err
				}
				sum := fmt.Sprintf("%x", sha256.Sum256(b))
				key := "SNIP_SHA256_" + tools.KeyEnv(targetFile)
				processorCfg.Vars[key] = sum
				return nil
			}

			cfg.RequiredFilesSrcProcessors[fileAbs] = append(cfg.RequiredFilesSrcProcessors[fileAbs], sumProcessor)

			return nil
		},
	}
)
