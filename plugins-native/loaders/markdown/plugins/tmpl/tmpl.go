package mainNative

import (
	"bytes"
	"io/ioutil"
	"os"

	"github.com/hairyhenderson/gomplate/v3"
	"github.com/joho/godotenv"

	"gitlab.com/youtopia.earth/ops/snip/plugin/runner"
	loaderMardownMod "gitlab.com/youtopia.earth/ops/snip/plugins-native/loaders/markdown/mod"
)

var (
	Mod = loaderMardownMod.Plugin{
		Mod: func(modCfg *loaderMardownMod.Config) error {

			// loaderCfg := modCfg.LoaderConfig
			// vars := loaderCfg.LoaderVars

			codeBlock := modCfg.CodeBlock

			// args := modCfg.Args
			processor := func(runnerCfg *runner.Config, src *string) (func(), error) {
				tmpfileEnv, err := ioutil.TempFile("", "snip-tmpl*.env")
				if err != nil {
					return nil, err
				}
				tmpfileEnvName := tmpfileEnv.Name()
				defer os.Remove(tmpfileEnvName)

				vars := make(map[string]string)
				for k, v := range runnerCfg.Vars {
					if k[0:1] != "@" {
						vars[k] = v
					}
				}
				if err := godotenv.Write(vars, tmpfileEnvName); err != nil {
					return nil, err
				}
				if err := tmpfileEnv.Close(); err != nil {
					return nil, err
				}

				buf := new(bytes.Buffer)

				input, err := ioutil.ReadFile(*src)
				if err != nil {
					return nil, err
				}

				config := &gomplate.Config{
					Input:       string(input),
					Out:         buf,
					OutputFiles: []string{"-"},
					// DataSources: []string{"snipEnv=file://" + tmpfileEnvName},
					Contexts: []string{".=file://" + tmpfileEnvName},
				}

				gomplate.RunTemplates(config)

				tmpfile, err := ioutil.TempFile("", "snip-tmpl*.tmplout")
				if err != nil {
					return nil, err
				}
				tmpfileName := tmpfile.Name()
				clean := func(){
					os.Remove(tmpfileName)
				}

				if err := ioutil.WriteFile(tmpfileName, buf.Bytes(), 0644); err != nil {
					return clean, err
				}

				*src = tmpfileName

				return clean, nil

			}
			codeBlock.Processors = append(codeBlock.Processors, processor)


			return nil
		},
	}
)
