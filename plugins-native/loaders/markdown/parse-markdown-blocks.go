package mainNative

import (
	"strings"

	"gitlab.com/golang-commonmark/markdown"

	"gitlab.com/youtopia.earth/ops/snip/plugin/loader"
	"gitlab.com/youtopia.earth/ops/snip/plugins-native/loaders/markdown/blocks"
)

const annotationPrefix = "<!-- snip:"
const annotationSuffix = "-->"

type ParseMdLoopParams struct {
	ignoreCodeOnce bool

	handleModOnce bool
	handleModArgs []string
}

func ParseMarkdownBlocks(cfg *loader.Config) []*blocks.Code {

	source := loader.GetFileContent(cfg, cfg.Command[0])

	md := markdown.New(markdown.XHTMLOutput(true), markdown.Nofollow(true))
	tokens := md.Parse(source)
	var codeBlocks []*blocks.Code
	parseMdLoopParams := &ParseMdLoopParams{}
	for blockIndex, t := range tokens {
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
			if parseMdLoopParams.ignoreCodeOnce {
				parseMdLoopParams.ignoreCodeOnce = false
				continue
			}

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

			codeBlock := &blocks.Code{
				Lang:    lang,
				Content: tok.Content,
				Index:   blockIndex,
			}

			if parseMdLoopParams.handleModOnce {
				modArgs := parseMdLoopParams.handleModArgs
				parseMdLoopParams.handleModOnce = false
				parseMdLoopParams.handleModArgs = nil
				loopContinue := handleMod(modArgs[0], modArgs[1:], codeBlock, cfg)
				if loopContinue {
					continue
				}
			}

			codeBlocks = append(codeBlocks, codeBlock)
		}
	}

	return codeBlocks

}
