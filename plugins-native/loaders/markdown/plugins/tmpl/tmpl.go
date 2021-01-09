package mainNative

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	gomplate "github.com/hairyhenderson/gomplate/v3"
	"github.com/joho/godotenv"

	"gitlab.com/ytopia/ops/snip/plugin/processor"
	loaderMarkdownMod "gitlab.com/ytopia/ops/snip/plugins-native/loaders/markdown/mod"
)

var (
	Mod = loaderMarkdownMod.Plugin{
		Mod: func(modCfg *loaderMarkdownMod.Config) error {

			loaderCfg := modCfg.LoaderConfig

			codeBlock := modCfg.CodeBlock

			// args := modCfg.Args
			processor := func(processorCfg *processor.Config, src *string) error {
				usr, _ := user.Current()
				rootPath := filepath.Join(usr.HomeDir, ".snip", loaderCfg.AppConfig.DeploymentName)

				tmplDirVars := "tmp/tmplvars"
				tmpDirVars := filepath.Join(rootPath, tmplDirVars)
				if err := os.MkdirAll(tmpDirVars, os.ModePerm); err != nil {
					return err
				}

				tmpfileEnv, err := ioutil.TempFile(tmpDirVars, "snip-tmpl*.env")
				if err != nil {
					return err
				}
				tmpfileEnvName := tmpfileEnv.Name()

				vars := make(map[string]string)
				for k, v := range processorCfg.RunVars.GetAll() {
					if k[0:1] != "@" {
						vars[k] = v
					}
				}
				if err := godotenv.Write(vars, tmpfileEnvName); err != nil {
					return err
				}
				if err := tmpfileEnv.Close(); err != nil {
					return err
				}

				buf := new(bytes.Buffer)

				input, err := ioutil.ReadFile(*src)
				if err != nil {
					return err
				}

				config := &gomplate.Config{
					Input:       string(input),
					Out:         buf,
					OutputFiles: []string{"-"},
					// DataSources: []string{"snipEnv=file://" + tmpfileEnvName},
					Contexts: []string{".=file://" + tmpfileEnvName},
				}

				err = gomplate.RunTemplates(config)
				if err != nil {
					return err
				}

				tmplDir := "tmp/tmpl"
				tmpDir := filepath.Join(rootPath, tmplDir)
				if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
					return err
				}
				tmpfile, err := ioutil.TempFile(tmpDir, "snip-tmpl*.tmplout")
				if err != nil {
					return err
				}
				tmpfileName := tmpfile.Name()

				b := buf.Bytes()
				if err := ioutil.WriteFile(tmpfileName, b, 0755); err != nil {
					return err
				}

				os.Chmod(tmpfileName, 0755)

				// logrus.Error(buf.String())

				*src = tmpfileName

				return nil

			}
			codeBlock.Processors = append(codeBlock.Processors, processor)

			return nil
		},
	}
)
