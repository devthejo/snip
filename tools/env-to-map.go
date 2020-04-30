package tools

import (
	"strings"
)

func EnvToMap(data []string) map[string]string {
	items := make(map[string]string)
	for _, item := range data {
		splits := strings.Split(item, "=")
		key := splits[0]
		val := strings.Join(splits[1:], "=")
		items[key] = val
	}
	return items
}
