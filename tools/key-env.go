package tools

import "strings"

var replacer = strings.NewReplacer("-", "_", ".", "_", "/", "__")

func KeyEnv(key string) string {
	return strings.ToUpper(replacer.Replace(key))
}
