package config

import (
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/devthejo/snip/tools"
	"github.com/devthejo/snip/xjsonnet"
)

func ConfigJsonnetRender(dirpaths []string, name string) (bool, error) {
	srcFilename := name + ".jsonnet"

	var allreadyExists bool
	var efile string
	exts := []string{"toml", "yaml", "yml", "hcl"}

	for _, dpath := range dirpaths {
		var err error
		dpath, err = filepath.Abs(dpath)
		if err != nil {
			continue
		}
		for _, ext := range exts {
			efile = dpath + "/" + name + "." + ext
			if ok, _ := tools.FileExists(efile); ok {
				allreadyExists = true
				break
			}
		}
		if allreadyExists {
			break
		}
	}
	if allreadyExists {
		logrus.Debugf(`jsonnet config allready exists for "%v" as "%v"`, name, efile)
		return false, nil
	}

	var dirpath string
	var found bool
	for _, dpath := range dirpaths {
		var err error
		dpath, err = filepath.Abs(dpath)
		if err != nil {
			continue
		}
		fpath := dpath + "/" + srcFilename
		if ok, err := tools.FileExists(fpath); ok {
			found = true
			dirpath = dpath
		} else if err != nil {
			return false, err
		}
	}
	if !found {
		logrus.Debugf(`%v not found in "%v"`, srcFilename, dirpaths)
		return false, nil
	}

	src := dirpath + "/" + srcFilename
	target := dirpath + "/" + name + ".json"

	if err := xjsonnet.RenderToFile(src, target); err != nil {
		return false, err
	}

	logrus.Debugf(`config writed to %v`, target)

	return true, nil
}
