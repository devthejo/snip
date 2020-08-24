package mainNative

import (
	"strings"

	"github.com/sirupsen/logrus"
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

func handleInstruction(t string, args []string, parseMdLoopParams *ParseMdLoopParams) {
	switch t {
	case "ignore-next":
		parseMdLoopParams.ignoreOnce = true
	case "mod":
		if len(args) > 0 {
			parseMdLoopParams.handleModOnce = true
			parseMdLoopParams.handleModArgs = args
		}
	default:
		logrus.Fatalf("unkown snip instruction %v", t)
	}
}

var mods map[string]interface{}

func handleMod(mod string, args []string, codeBlock *CodeBlock) {
	logrus.Debugf("mod: %v, args: %v", mod, args)
	if _, ok := mods[mod]; !ok {
		logrus.Fatalf("unkown markdown mod: %v", mod)
	}
}

func ParseMarkdownBlock(cfg *loader.Config) []*CodeBlock {

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
				handleMod(modArgs[0], modArgs[1:], codeBlock)
				continue
			}

			codeBlocks = append(codeBlocks, codeBlock)
		}
	}

	return codeBlocks

}
