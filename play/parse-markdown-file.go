package play

import (
	"bytes"
	"io/ioutil"
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

func ParseMarkdownFile(play *Play, app App) {

	cfg := app.GetConfig()

	// playMap := make(map[interface{}]interface{})

	file := cfg.SnippetsDir + "/" + play.Name + ".md"

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
	// TODO add descriptions from metas
	play.Markdown = markdownString

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
	ParseMap(play, metaData)

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
				play.SourceCodeBlocks = append(play.SourceCodeBlocks, codeBlock)
				play.CodeBlocks = append(play.CodeBlocks, generateBashVarsCodeBlock(play))
				play.CodeBlocks = append(play.CodeBlocks, codeBlock)
			}
		}
	}

}

func generateBashVarsCodeBlock(play *Play) *CodeBlock {
	var content string

	// for _, v := range play.Vars {
	//
	// }

	codeBlock := &CodeBlock{
		Type:    CodeBlockBash,
		Content: content,
	}

	return codeBlock
}
