package xjsonnet

import (
	"os"
)

func RenderToFile(src string, target string) error {
	var output string
	if res, err := Render(src); err != nil {
		return err
	} else {
		output = res
	}

	f, err := os.Create(target)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(output)
	if err == nil {
		err = f.Sync()
	}
	if err != nil {
		return err
	}

	return nil

}
