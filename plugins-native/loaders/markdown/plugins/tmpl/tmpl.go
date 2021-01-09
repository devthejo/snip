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
				tmpfileEnv, err := ioutil.TempFile("", "snip-tmpl*.env")
				if err != nil {
					return err
				}
				tmpfileEnvName := tmpfileEnv.Name()
				defer os.Remove(tmpfileEnvName)

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

				usr, _ := user.Current()
				tmplDir := "tmp/tmpl"
				rootPath := filepath.Join(usr.HomeDir, ".snip", loaderCfg.AppConfig.DeploymentName)
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
