package config

import (
	"time"
)

// Config struct
type Config struct {
	LogLevel       string `mapstructure:"LOG_LEVEL" json:"log_level"`
	LogType        string `mapstructure:"LOG_TYPE" json:"log_type"`
	LogForceColors bool   `mapstructure:"LOG_FORCE_COLORS" json:"logs_force_colors"`
	CWD            string `mapstructure:"CWD" json:"cwd"`

	Vars      map[string]interface{}            `mapstructure:"VARS" json:"vars"`
	VarsLoops map[string]map[string]interface{} `mapstructure:"VARS_LOOPS" json:"vars_loops"`
	Play      map[string]interface{}            `mapstructure:"PLAY" json:"play"`

	SnippetsDir string `mapstructure:"SNIPPETS_DIR" json:"snippets_dir"`
	BuildDir    string `mapstructure:"BUILD_DIR" json:"build_dir"`

	MarkdownOutput string `mapstructure:"MARKDOWN_OUTPUT" json:"markdown_output"`

	ShutdownTimeout time.Duration `mapstructure:"SHUTDOWN_TIMEOUT" json:"shutdownTimeout,omitempty"`
}
