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

	SSHHost     string `mapstructure:"SSH_HOST" json:"ssh_host"`
	SSHPort     int    `mapstructure:"SSH_PORT" json:"ssh_port"`
	SSHUser     string `mapstructure:"SSH_USER" json:"ssh_user"`
	SSHFile     string `mapstructure:"SSH_FILE" json:"ssh_file"`
	SSHPass     string `mapstructure:"SSH_PASS" json:"ssh_pass"`
	SSHRetryMax int    `mapstructure:"SSH_RETRY_MAX" json:"ssh_retry_max"`
}
