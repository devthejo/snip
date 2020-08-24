package mainNative

import "github.com/sirupsen/logrus"

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
