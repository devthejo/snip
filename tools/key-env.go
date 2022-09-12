package tools

import (
	"regexp"
	"strings"
)

var regNormalizeEnvVarName = regexp.MustCompile("[^A-Z0-9_]+")
var regSingleUnderscore = regexp.MustCompile(`_+`)

func KeyEnv(key string) string {
	key = strings.ToUpper(key)
	key = regNormalizeEnvVarName.ReplaceAllString(key, "_")
	key = regSingleUnderscore.ReplaceAllString(key, "_")
	key = strings.Trim(key, "_")
	return key
}
