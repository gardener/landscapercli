package util

import (
	"io"
	"os"
)

func IsDirectoryEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}

	defer f.Close()

	_, err = f.Readdirnames(1)
	if err != nil {
		if err == io.EOF {
			return true, nil
		}

		return false, err
	}

	return false, nil
}
