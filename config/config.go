package config

// Config struct
type Config struct {
	LogLevel       string `mapstructure:"LOG_LEVEL" json:"log_level"`
	LogType        string `mapstructure:"LOG_TYPE" json:"log_type"`
	LogForceColors bool   `mapstructure:"LOG_FORCE_COLORS" json:"logs_force_colors"`
	CWD            string `mapstructure:"CWD" json:"cwd"`

	Playbook []interface{} `mapstructure:"PLAYBOOK" json:"playbook"`

	SnippetsDir string `mapstructure:"SNIPPETS_DIR" json:"snippets_dir"`
	BuildDir    string `mapstructure:"BUILD_DIR" json:"build_dir"`

	MarkdownOutput string `mapstructure:"MARKDOWN_OUTPUT" json:"markdown_output"`
}
