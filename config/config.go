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

	Playbooks []string `mapstructure:"PLAYBOOKS" json:"key,omitempty"`

	Vars      map[string]interface{}            `mapstructure:"VARS" json:"vars"`
	VarsLoops map[string]map[string]interface{} `mapstructure:"VARS_LOOPS" json:"vars_loops"`
	Play      map[string]interface{}            `mapstructure:"PLAY" json:"play"`

	SnippetsDir string `mapstructure:"SNIPPETS_DIR" json:"snippets_dir"`

	MarkdownOutput string `mapstructure:"MARKDOWN_OUTPUT" json:"markdown_output"`

	ShutdownTimeout time.Duration `mapstructure:"SHUTDOWN_TIMEOUT" json:"shutdownTimeout,omitempty"`

	Runner  string   `mapstructure:"RUNNER" json:"runner,omitempty"`
	Loaders []string `mapstructure:"LOADERS" json:"loaders,omitempty"`

	DeploymentName string `mapstructure:"DEPLOYMENT_NAME" json:"deployment_name,omitempty"`

	PlayKey      []string `mapstructure:"KEY" json:"key,omitempty"`
	PlayKeyDeps  bool     `mapstructure:"KEY_DEPS" json:"key_deps,omitempty"`
	PlayKeyPost  bool     `mapstructure:"KEY_POST" json:"key_post,omitempty"`
	PlayKeyStart string   `mapstructure:"KEY_START" json:"key_start,omitempty"`
	PlayKeyEnd   string   `mapstructure:"KEY_END" json:"key_end,omitempty"`
	PlayResume   bool     `mapstructure:"RESUME" json:"resume,omitempty"`

	PlayNoClean bool `mapstructure:"NO_CLEAN" json:"no_clean,omitempty"`
}
