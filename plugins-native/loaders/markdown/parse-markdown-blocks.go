package mainNative

import (
	"strings"

	cmap "github.com/orcaman/concurrent-map"
	"gitlab.com/golang-commonmark/markdown"

	"gitlab.com/youtopia.earth/ops/snip/plugin/loader"
)

const annotationPrefix = "<!-- snip:"
const annotationSuffix = "-->"

type ParseMdLoopParams struct {
	ignoreOnce bool

	handleModOnce bool
	handleModArgs []string
}

func ParseMarkdownBlocks(cfg *loader.Config, plugins cmap.ConcurrentMap) []*CodeBlock {

	source := GetMarkdownContent(cfg)

	md := markdown.New(markdown.XHTMLOutput(true), markdown.Nofollow(true))
	tokens := md.Parse(source)
	var codeBlocks []*CodeBlock
	parseMdLoopParams := &ParseMdLoopParams{}
	for _, t := range tokens {
		if parseMdLoopParams.ignoreOnce {
			parseMdLoopParams.ignoreOnce = false
			continue
		}
		switch tok := t.(type) {
		case *markdown.Inline:
			if strings.HasPrefix(tok.Content, annotationPrefix) &&
				strings.HasSuffix(tok.Content, annotationSuffix) {
				content := tok.Content
				content = strings.TrimPrefix(content, annotationPrefix)
				content = strings.TrimSuffix(content, annotationSuffix)
				arr := strings.Split(content, " ")
				if len(arr) > 0 {
					handleInstruction(arr[0], arr[1:], parseMdLoopParams)
				}
			}
		case *markdown.Fence:
			if tok.Content == "" || tok.Params == "" {
				continue
			}

			lang := tok.Params
			langArgs := strings.Split(lang, " ")

			if len(langArgs) > 1 {
				if langArgs[1] == "#" {
					continue
				}
				if langArgs[1][0:1] == "!" {
					parseMdLoopParams.handleModOnce = true
					lang = langArgs[0]
					modArgs := []string{langArgs[1][1:]}
					modArgs = append(modArgs, langArgs[2:]...)
					parseMdLoopParams.handleModArgs = modArgs
				}
			}

			codeBlock := &CodeBlock{
				Lang:    lang,
				Content: tok.Content,
			}

			if parseMdLoopParams.handleModOnce {
				modArgs := parseMdLoopParams.handleModArgs
				parseMdLoopParams.handleModOnce = false
				parseMdLoopParams.handleModArgs = nil
				handleMod(modArgs[0], modArgs[1:], codeBlock, plugins)
				continue
			}

			codeBlocks = append(codeBlocks, codeBlock)
		}
	}

	return codeBlocks

}
