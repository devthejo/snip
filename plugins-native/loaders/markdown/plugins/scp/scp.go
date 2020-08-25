package mainNative

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"gitlab.com/youtopia.earth/ops/snip/errors"
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

			cfg.RequiredFiles[fileAbs] = file

			fileAbsRemote := filepath.Join("${SNIP_SNIPPETS_PATH}", snippetPath)

			targetFile := args[0]
			codeBlock.Lang = "sh"
			codeBlock.Content = "mv " + fileAbsRemote + " " + targetFile

			return nil
		},
	}
)
