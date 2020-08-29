package play

import (
	"regexp"
	"strings"
)

var regNormalizeTreeKeyParts = regexp.MustCompile("[^a-zA-Z0-9-_.]+")

func GetTreeKeyParts(parent interface{}) []string {
	var parts []string
	for {
		var part string
		switch p := parent.(type) {
		case *LoopRow:
			if p == nil {
				parent = nil
				break
			}
			// part = "row." + p.GetKey()
			part = p.GetKey()
			parent = p.ParentPlay
		case *Play:
			if p == nil {
				parent = nil
				break
			}
			// part = "play." + p.GetKey()
			part = p.GetKey()
			parent = p.ParentLoopRow
		case nil:
			parent = nil
		}
		if parent == nil {
			break
		}

		part = strings.ReplaceAll(part, "./", "-")
		part = regNormalizeTreeKeyParts.ReplaceAllString(part, "_")
		parts = append([]string{part}, parts...)
	}
	return parts
}
