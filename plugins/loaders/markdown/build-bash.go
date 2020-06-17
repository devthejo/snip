package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"gitlab.com/youtopia.earth/ops/snip/errors"
	"gitlab.com/youtopia.earth/ops/snip/plugin/loader"
)

type CodeBlockType int

const (
	CodeBlockBash CodeBlockType = iota
)

type CodeBlock struct {
	Type    CodeBlockType
	Content string
}

func BuildBash(cfg *loader.Config) error {
	mdpath := cfg.Command[0]

	codeBlocks, defaultsPlayProps := ParseMarkdownFile(cfg)
	cfg.DefaultsPlayProps = defaultsPlayProps

	appCfg := cfg.AppConfig
	now := time.Now()
	nowText := now.Format("2006-01-02 15:04:05")

	file := appCfg.BuildDir + "/snippets/" + mdpath + ".bash"
	dir := filepath.Dir(file)
	os.MkdirAll(dir, os.ModePerm)

	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	errors.Check(err)
	defer f.Close()
	outputAppend := func(str string) {
		_, err := f.WriteString(str)
		errors.Check(err)
	}

	outputAppend("#!/usr/bin/env bash\n\n")
	outputAppend("# play snippet: " + mdpath + "\n")
	outputAppend("# generated by snip on " + nowText + "\n\n")
	outputAppend("set -e\n\n")

	for _, codeBlock := range codeBlocks {
		content := codeBlock.Content
		content = strings.Trim(content, "\n")
		outputAppend(content + "\n")
	}
	logrus.Debugf("writed bash from md to %v", file)

	cfg.Command[0] = "~/.snip/" + file

	cfg.RequiredFiles[file] = file

	return nil
}
