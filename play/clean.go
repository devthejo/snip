package play

import (
	"fmt"
	"os"
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
	"tmp",
}

func cleanRootPath(app App) error {
	rootPath := GetRootPath(app)
	for _, d := range cleanDirs {
		dAbs := filepath.Join(rootPath, d)
		if err := os.RemoveAll(dAbs); err != nil {
			return fmt.Errorf("unable to clean namespace path %s %v", rootPath, err)
		}
	}
	return nil
}
