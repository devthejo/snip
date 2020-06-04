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

	FlagShutdownTimeoutDesc = "shutdown timeout"

	FlagSSHHostDesc     = "SSH host address"
	FlagSSHPortDesc     = "SSH port number"
	FlagSSHUserDesc     = "SSH user name"
	FlagSSHFileDesc     = "SSH key file path"
	FlagSSHPassDesc     = "SSH password (to use with user or to unlock key file if provided)"
	FlagSSHRetryMaxDesc = "SSH maximum retry attempts"
)

var (
	FlagLogForceColorsDefault = false
	FlagLogTypeDefault        = "text"
	FlagLogLevelDefault       = "info"

	FlagSnippetsDirDefault    = "snippets"
	FlagBuildDirDefault       = "build"
	FlagMarkdownOutputDefault = "snip.md"

	FlagShutdownTimeoutDefault = "5"

	FlagSSHHostDefault     = "localhost"
	FlagSSHPortDefault     = 22
	FlagSSHUserDefault     = ""
	FlagSSHFileDefault     = ""
	FlagSSHPassDefault     = ""
	FlagSSHRetryMaxDefault = 3
)
