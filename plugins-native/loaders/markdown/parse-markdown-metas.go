package mainNative

import (
	"bytes"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
	"gitlab.com/youtopia.earth/ops/snip/errors"
	"gitlab.com/youtopia.earth/ops/snip/plugin/loader"
)

func ParseMarkdownMetas(cfg *loader.Config) map[string]interface{} {

	mdPath := cfg.Command[0]

	source := loader.GetFileContent(cfg, cfg.Command[0])

	// markdownString := string(source)
	// if markdownString[0:4] == "---\n" {
	// 	i := strings.Index(markdownString, "\n---\n")
	// 	markdownString = strings.Trim(markdownString[i+5:], "\n")
	// }

	md := goldmark.New(
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
	err := md.Convert(source, &buf, parser.WithContext(context))
	errors.Check(err)

	defaultsPlayProps := meta.Get(context)
	if defaultsPlayProps == nil {
		defaultsPlayProps = make(map[string]interface{})
	}
	if defaultsPlayProps["key"] == nil {
		defaultsPlayProps["key"] = mdPath
	}
	if defaultsPlayProps["title"] == nil {
		defaultsPlayProps["title"] = loader.SnippetDefaultTitle(mdPath, cfg)
	}

	return defaultsPlayProps
}
