package config

import (
	"os"
)

func (cl *ConfigLoader) ConfigureCWD() {
	cfg := cl.Config

	cwd, _ := cl.RootCmd.PersistentFlags().GetString("cwd")
	if cwd != "" {
		cfg.CWD = cwd
	} else if CWD := os.Getenv(cl.PrefixEnv("CWD")); CWD != "" {
		cfg.CWD = CWD
	}

	if cfg.CWD != "" {
		err := os.Chdir(cfg.CWD)
		if err != nil {
			panic(err)
		}
	}
}
