package config

const (
	FlagConfigDesc         = "config file (default is ./snip.yml or /etc/snip.yml)"
	FlagLogLevelDesc       = "log level panic|fatal|error|warning|info|debug|trace"
	FlagLogTypeDesc        = "log type json|text"
	FlagLogForceColorsDesc = "force log colors for text log when no tty"
	FlagCWDDesc            = "current working directory"

	FlagSnippetsDirDesc = "snippets directory"
)

var (
	FlagLogForceColorsDefault = false
	FlagLogTypeDefault        = "json"
	FlagLogLevelDefault       = "info"

	FlagSnippetsDirDefault = "snippets"
)
