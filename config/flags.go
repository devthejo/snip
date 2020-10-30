package config

const (
	FlagConfigDesc         = "config file (default is ./snip.yml or /etc/snip.yml)"
	FlagLogLevelDesc       = "log level panic|fatal|error|warning|info|debug|trace"
	FlagLogTypeDesc        = "log type json|text"
	FlagLogForceColorsDesc = "force log colors for text log when no tty"
	FlagCWDDesc            = "current working directory"

	FlagSnippetsDirDesc    = "snippets directory"
	FlagMarkdownOutputDesc = "markdown output file"

	FlagShutdownTimeoutDesc = "shutdown timeout"

	FlagRunnerDesc  = "default runner"
	FlagLoadersDesc = "default loaders"

	FlagDeploymentNameDesc = "deployment name, must be unique to avoid collisions"

	FlagPlayKeyDesc      = "comma separated list of keys (or tree keys) to run skipping all others"
	FlagPlayKeyDepsDesc  = "load out of scope dependencies of provided keys"
	FlagPlayKeyPostDesc  = "load out of scope post_install of provided keys"
	FlagPlayKeyStartDesc = "skip all before key"
	FlagPlayKeyEndDesc   = "skip all after key"
	FlagPlayResumeDesc   = "resume feeding --key-start with previous interrupted playing"
)

var (
	FlagLogForceColorsDefault = false
	FlagLogTypeDefault        = "text"
	FlagLogLevelDefault       = "info"

	FlagSnippetsDirDefault    = "snippets"
	FlagMarkdownOutputDefault = "snip.md"

	FlagShutdownTimeoutDefault = "5"

	FlagRunnerDefault  = "sh"
	FlagLoadersDefault = `["markdown"]`

	FlagDeploymentNameDefault = "default"

	FlagPlayKeyDefault      = []string{}
	FlagPlayKeyDepsDefault  = false
	FlagPlayKeyPostDefault  = false
	FlagPlayKeyStartDefault = ""
	FlagPlayKeyEndDefault   = ""
	FlagPlayResumeDefault   = false
)
