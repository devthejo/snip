package mainNative

import (
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"gitlab.com/youtopia.earth/ops/snip/errors"
	"gitlab.com/youtopia.earth/ops/snip/plugin/loader"
)

func BuildScripts(cfg *loader.Config) error {
	mdpath := cfg.Command[0]

	codeBlocks := ParseMarkdownBlocks(cfg)

	now := time.Now()
	nowText := now.Format("2006-01-02 15:04:05")
	usr, _ := user.Current()

	buildDir := "build/snippets"
	rootPath := filepath.Join(usr.HomeDir, ".snip", cfg.AppConfig.DeploymentName)

	buildFile := func(file string, content string) {
		fileAbs := filepath.Join(rootPath, file)

		dir := filepath.Dir(fileAbs)
		os.MkdirAll(dir, os.ModePerm)

		fileP, err := os.OpenFile(fileAbs, os.O_APPEND|os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
		errors.Check(err)
		defer fileP.Close()

		_, err = fileP.WriteString(content)
		errors.Check(err)

		cfg.RequiredFiles[fileAbs] = file

		logrus.Debugf("writed script from md to %v", file)
	}

	mainScriptContent := "#!/bin/sh\n\n"
	mainScriptContent += "# play snippet: " + mdpath + "\n"
	mainScriptContent += "# generated by snip on " + nowText + "\n\n"

	for i, codeBlock := range codeBlocks {
		content := codeBlock.Content
		var header string
		switch codeBlock.Lang {
		case "sh":
			header += "#!/bin/sh\n\n"
		case "bash":
			header = "#!/usr/bin/env bash\n\n"
			header += "set -e\n\n"
		default:
			header = "#!/usr/bin/env " + codeBlock.Lang + "\n\n"
		}
		content = strings.Trim(content, "\n")
		content = header + content
		content += "\n\n# snip vars export \n"
		for _, vr := range cfg.RegisterVars {
			if !vr.Enable {
				continue
			}
			content += `echo "${` + vr.GetSource() + `}">${SNIP_VARS_TREEPATH}/` + vr.GetFrom() + "\n"
		}
		snippetPath := mdpath + "_" + strconv.Itoa(i)
		file := filepath.Join(buildDir, snippetPath)
		buildFile(file, content)
		mainScriptContent += filepath.Join("${SNIP_SNIPPETS_PATH}", snippetPath) + "\n"
	}

	launcher := mdpath + ".sh"
	mainFile := filepath.Join(buildDir, launcher)
	buildFile(mainFile, mainScriptContent)

	bin := filepath.Join("${SNIP_SNIPPETS_PATH}", launcher)
	cfg.Command[0] = bin

	return nil
}
