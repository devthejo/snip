package play

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

func Clean(app App) error {
	if err := cleanRootPath(app); err != nil {
		return err
	}
	return nil
}

var cleanDirs = []string{
	"build",
	"vars",
}

func cleanRootPath(app App) error {
	usr, _ := user.Current()
	rootPath := filepath.Join(usr.HomeDir, ".snip", app.GetConfig().DeploymentName)
	for _, d := range cleanDirs {
		dAbs := filepath.Join(rootPath, d)
		if err := os.RemoveAll(dAbs); err != nil {
			return fmt.Errorf("unable to clean namespace path %s %v", rootPath, err)
		}
	}
	return nil
}
