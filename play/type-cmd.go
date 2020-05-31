package play

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"gitlab.com/youtopia.earth/ops/snip/errors"
)

type Cmd struct {
	Play *Play

	Command string
	Args    []string
	Vars    map[string]string

	IsMD       bool
	MDPath     string
	Markdown   string
	CodeBlocks []*CodeBlock

	RunStack []*Run
}

type CodeBlockType int

const (
	CodeBlockBash CodeBlockType = iota
)

type CodeBlock struct {
	Type    CodeBlockType
	Content string
}

func (cmd *Cmd) Parse(c []string) {
	if len(c) < 1 {
		return
	}
	command := c[0]
	cmd.IsMD = strings.HasSuffix(command, ".md")
	if cmd.IsMD {
		cmd.MDPath = command
		cmd.BuildBashFromMD()
	} else {
		cmd.Command = command
	}
	if len(c) > 1 {
		cmd.Args = c[1:]
	}
}

func (cmd *Cmd) BuildBashFromMD() {
	app := cmd.Play.App

	ParseMarkdownFile(app, cmd.MDPath, cmd)

	cfg := app.GetConfig()
	now := app.GetNow()
	nowText := now.Format("2006-01-02 15:04:05")

	file := cfg.BuildDir + "/snippets/" + cmd.MDPath + ".bash"
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
	outputAppend("# play snippet: " + cmd.MDPath + "\n")
	outputAppend("# generated by snip on " + nowText + "\n\n")
	outputAppend("set -e\n\n")

	for _, codeBlock := range cmd.CodeBlocks {
		content := codeBlock.Content
		content = strings.Trim(content, "\n")
		outputAppend(content + "\n")
	}
	logrus.Debugf("writed bash from md to %v", file)

	cmd.Command = file
}

func (cmd *Cmd) Run(ctx *RunCtx) {

	vars := make(map[string]string)
	for k, v := range ctx.VarsDefault.Items() {
		vars[k] = v.(string)
	}
	for k, v := range ctx.Vars.Items() {
		vars[k] = v.(string)
	}

	// logrus.Debugf(strings.Repeat("  ", cmd.Play.Depth+2)+" vars: %v", tools.JsonEncode(vars))

	if !ctx.ReadyToRun {
		return
	}

	r := CreateRun(cmd)
	r.Vars = vars

	cmd.RunStack = append(cmd.RunStack, r)

	r.Exec()

}

func unexpectedTypeCmd(m map[string]interface{}, key string) {
	errors.UnexpectedType(m, key, "cmd")
}
