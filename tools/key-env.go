package tools

import "strings"

var replacer = strings.NewReplacer("-", "_", ".", "_")

func KeyEnv(key string) string {
	return strings.ToUpper(replacer.Replace(key))
}
