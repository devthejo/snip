package tools

import "os"

func FileExists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, err
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}
