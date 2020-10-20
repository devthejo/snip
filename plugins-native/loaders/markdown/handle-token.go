package mainNative

import (
	"strings"

	"gitlab.com/golang-commonmark/markdown"
	// "github.com/sirupsen/logrus"

	"gitlab.com/ytopia/ops/snip/plugin/loader"
	"gitlab.com/ytopia/ops/snip/plugins-native/loaders/markdown/blocks"
)

const annotationPrefix = "<!-- snip:"
const annotationSuffix = "-->"

func handleToken(cfg *loader.Config, tIndex int, t interface{}, codeBlocks *[]*blocks.Code, parseMdLoopParams *ParseMdLoopParams, snippetPath string) {
	switch tok := t.(type) {
	case *markdown.Inline:
		if strings.HasPrefix(tok.Content, annotationPrefix) &&
			strings.HasSuffix(tok.Content, annotationSuffix) {
			content := tok.Content
			content = strings.TrimPrefix(content, annotationPrefix)
			content = strings.TrimSuffix(content, annotationSuffix)
			arr := strings.Split(content, " ")
			if len(arr) > 0 {
				handleInstruction(tIndex, arr[0], arr[1:], parseMdLoopParams, cfg, codeBlocks, snippetPath)
			}
		}
	case *markdown.Fence:
		if parseMdLoopParams.ignoreCodeOnce {
			parseMdLoopParams.ignoreCodeOnce = false
			return
		}

		if tok.Content == "" || tok.Params == "" {
			return
		}

		lang := tok.Params
		langArgs := strings.Split(lang, " ")

		if len(langArgs) > 1 {

			if langArgs[1] == "#" {
				return
			}
			lang = langArgs[0]
			parseMdLoopParams.handleModsOnce = true

			var handleModsArgs [][]string
			var modArgs []string
			for i, langArg := range langArgs {
				if langArg[0:1] == "!" {
					if len(modArgs) > 0 {
						handleModsArgs = append(handleModsArgs, modArgs)
					}
					modArgs = []string{langArg[1:]}
				} else if i != 0 {
					modArgs = append(modArgs, langArg)
				}
			}
			if len(modArgs) > 0 {
				handleModsArgs = append(handleModsArgs, modArgs)
			}

			parseMdLoopParams.handleModsArgs = append(handleModsArgs, parseMdLoopParams.handleModsArgs...)

		}

		codeBlock := &blocks.Code{
			Lang:    lang,
			Content: tok.Content,
			Index:   tIndex,
		}

		if parseMdLoopParams.handleModsOnce {
			modsArgs := parseMdLoopParams.handleModsArgs
			parseMdLoopParams.handleModsOnce = false
			parseMdLoopParams.handleModsArgs = nil
			var loopContinue bool
			for _, modArgs := range modsArgs {
				loopContinue = handleMod(modArgs[0], modArgs[1:], codeBlock, cfg)
			}
			if loopContinue {
				return
			}
		}

		*codeBlocks = append(*codeBlocks, codeBlock)
	}
}
