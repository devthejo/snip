package plugin

import (
	"path/filepath"
)

type AppConfig struct {
	DeploymentName string
	SnippetsDir    string
	Runner         string
}

func (appCfg *AppConfig) TreepathVarsDir(kp []string) string {
	dp := kp[0 : len(kp)-2]
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
