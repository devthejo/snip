package mainNative

import (
	"crypto/sha256"
	"fmt"

	"gitlab.com/golang-commonmark/markdown"

	"gitlab.com/youtopia.earth/ops/snip/plugin/loader"
)

func ParseMarkdownBlocksChecksum(cfg *loader.Config) []string {

	source := loader.GetFileContent(cfg, cfg.Command[0])

	md := markdown.New(markdown.XHTMLOutput(true), markdown.Nofollow(true))
	tokens := md.Parse(source)
	var checksums []string
	for _, t := range tokens {
		switch tok := t.(type) {
		case *markdown.Fence:
			if tok.Content == "" || tok.Params == "" {
				continue
			}
			sum := fmt.Sprintf("%x", sha256.Sum256([]byte(tok.Content)))
			checksums = append(checksums, sum)
		}
	}

	return checksums

}
