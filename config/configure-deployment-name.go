package config

import (
	"os"
	"path"
)

func (cl *ConfigLoader) ConfigureDeploymentName() {
	cfg := cl.Config

	if cfg.DeploymentName == "" {
		wd, _ := os.Getwd()
		cfg.DeploymentName = path.Base(wd)
	}
}
