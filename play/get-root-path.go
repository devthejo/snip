package play

import (
	"os/user"
	"path/filepath"
)

func GetRootPath(app App) string {
	usr, _ := user.Current()
	rootPath := filepath.Join(usr.HomeDir, ".snip", app.GetConfig().DeploymentName)
	return rootPath
}
