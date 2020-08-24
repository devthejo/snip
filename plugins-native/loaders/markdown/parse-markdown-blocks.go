package mainNative

import (
	"gitlab.com/golang-commonmark/markdown"
	"gitlab.com/youtopia.earth/ops/snip/plugin/loader"
)

func ParseMarkdownBlock(cfg *loader.Config) []*CodeBlock {

	source := GetMarkdownContent(cfg)

	md := markdown.New(markdown.XHTMLOutput(true), markdown.Nofollow(true))
	tokens := md.Parse(source)
	var codeBlocks []*CodeBlock
	for _, t := range tokens {
		switch tok := t.(type) {
		case *markdown.Fence:
			lang := tok.Params
			if tok.Content == "" || lang == "" || lang[len(lang)-1:] == "#" {
				continue
			}
			codeBlock := &CodeBlock{
				Lang:    lang,
				Content: tok.Content,
			}
			codeBlocks = append(codeBlocks, codeBlock)
		}
	}

	return codeBlocks

}
