package mainNative

import (
	"gitlab.com/golang-commonmark/markdown"

	"gitlab.com/ytopia/ops/snip/plugin/loader"
	"gitlab.com/ytopia/ops/snip/plugins-native/loaders/markdown/blocks"
)

type ParseMdLoopParams struct {
	ignoreCodeOnce bool

	handleModsOnce bool
	handleModsArgs [][]string
}

func ParseMarkdownBlocks(cfg *loader.Config) []*blocks.Code {

	snippetPath := cfg.Command[0]
	source := loader.GetFileContent(cfg, snippetPath)

	md := markdown.New(markdown.XHTMLOutput(true), markdown.Nofollow(true))
	tokens := md.Parse(source)
	var codeBlocks []*blocks.Code
	codeBlocksP := &codeBlocks
	parseMdLoopParams := &ParseMdLoopParams{}
	for index, t := range tokens {
		handleToken(cfg, index, t, codeBlocksP, parseMdLoopParams, snippetPath)
	}

	return codeBlocks

}
