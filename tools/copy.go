package tools

import (
	"fmt"
	"io"
	"os"

	"gitlab.com/ytopia/ops/snip/tools/copydir"
)

func Copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	mode := sourceFileStat.Mode()
	if !mode.IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)

	if err := os.Chmod(dst, mode); err != nil {
		return nBytes, err
	}

	return nBytes, err
}

func CopyDir(src, dst string) error {
	return copydir.CopyDirectory(src, dst)
}
