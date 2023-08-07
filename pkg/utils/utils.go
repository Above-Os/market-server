package utils

import (
	"os"
	"path/filepath"
)

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CheckDir(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	dir := filepath.Dir(absPath)

	exist, err := PathExists(dir)
	if exist {
		return nil
	}

	err = os.MkdirAll(dir, 0755)
	return err
}
