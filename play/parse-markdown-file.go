package play

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
	"gitlab.com/golang-commonmark/markdown"
	"gitlab.com/youtopia.earth/ops/snip/errors"
	"gitlab.com/youtopia.earth/ops/snip/tools"
)

func ParseMarkdownFile(app App, mdPath string, cmd *Cmd) {

	cfg := app.GetConfig()

	var file string
	if len(mdPath) > 0 && mdPath[0:1] != "/" {
		file = cfg.SnippetsDir + "/" + mdPath
	} else {
		file = mdPath
	}

	exists, err := tools.FileExists(file)
	errors.Check(err)
	if !exists {
		logrus.Fatalf("file not found %v", file)
	}

	logrus.Debugf("loading file %v", file)
	source, err := ioutil.ReadFile(file)
	errors.Check(err)

	markdownString := string(source)
	if markdownString[0:4] == "---\n" {
		i := strings.Index(markdownString, "\n---\n")
		markdownString = strings.Trim(markdownString[i+5:], "\n")
	}
	// TODO add descriptions from metas and vars
	cmd.Markdown = markdownString

	md1 := goldmark.New(
		goldmark.WithExtensions(
			meta.New(meta.WithTable()),
		),
		goldmark.WithRendererOptions(
			renderer.WithNodeRenderers(
				util.Prioritized(extension.NewTableHTMLRenderer(), 500),
			),
		),
	)
	var buf bytes.Buffer
	context := parser.NewContext()
	err = md1.Convert(source, &buf, parser.WithContext(context))
	errors.Check(err)

	metaData := meta.Get(context)
	if metaData["key"] == nil {
		metaData["key"] = mdPath
	}
	if metaData["title"] == nil {
		title := mdPath
		title = strings.TrimSuffix(title, filepath.Ext(title))
		title = strings.ReplaceAll(title, "-", " ")
		title = strings.ReplaceAll(title, "/", " ")
		title = "snippet: " + title
		metaData["title"] = title
	}
	cmd.Play.ParseMapAsDefault(metaData)

	md2 := markdown.New(markdown.XHTMLOutput(true), markdown.Nofollow(true))
	tokens := md2.Parse(source)
	for _, t := range tokens {
		switch tok := t.(type) {
		case *markdown.Fence:
			if tok.Content != "" && (tok.Params == "bash" || tok.Params == "sh") {
				codeBlock := &CodeBlock{
					Type:    CodeBlockBash,
					Content: tok.Content,
				}
				cmd.CodeBlocks = append(cmd.CodeBlocks, codeBlock)
			}
		}
	}

}
