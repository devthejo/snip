package tools

import "strings"

func IndentBash(text, indent string) string {
	if len(text) > 0 && text[len(text)-1:] == "\n" {
		result := ""
		for _, j := range strings.Split(text[:len(text)-1], "\n") {
			if j != "EOF" {
				result += indent
			}
			result += j + "\n"
		}
		return result
	}
	result := ""
	for _, j := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		if j != "EOF" {
			result += indent
		}
		result += j + "\n"
	}
	return result[:len(result)-1]
}
