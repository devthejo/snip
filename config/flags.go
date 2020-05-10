package config

const (
	FlagConfigDesc         = "config file (default is ./snip.yml or /etc/snip.yml)"
	FlagLogLevelDesc       = "log level panic|fatal|error|warning|info|debug|trace"
	FlagLogTypeDesc        = "log type json|text"
	FlagLogForceColorsDesc = "force log colors for text log when no tty"
	FlagCWDDesc            = "current working directory"

	FlagSnippetsDirDesc    = "snippets directory"
	FlagBuildDirDesc       = "bash build directory"
	FlagMarkdownOutputDesc = "markdown output file"
)

var (
	FlagLogForceColorsDefault = false
	FlagLogTypeDefault        = "text"
	FlagLogLevelDefault       = "info"

	FlagSnippetsDirDefault    = "snippets"
	FlagBuildDirDefault       = "build"
	FlagMarkdownOutputDefault = "snip.md"
)
