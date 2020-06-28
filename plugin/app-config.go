package plugin

import (
	"path/filepath"
)

type AppConfig struct {
	DeploymentName string
	SnippetsDir    string
	Runner         string
}

func (appCfg *AppConfig) TreeDirVars(kp []string) string {
	return appCfg.TreeDir(kp, -2)
}
func (appCfg *AppConfig) TreeDirLauncher(kp []string) string {
	return appCfg.TreeDir(kp, 0)
}
func (appCfg *AppConfig) TreeDir(kp []string, index int) string {
	dp := kp[0 : len(kp)+index]
	dirParts := make([]string, len(dp))
	for i := 0; i < len(dp); i++ {
		var p string
		if i%2 == 0 {
			p = "key"
		} else {
			p = "row"
		}
		p += "."
		p += dp[i]
		dirParts[i] = p
	}
	return filepath.Join(dirParts...)
}
