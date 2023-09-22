package utils

import (
	"github.com/golang/glog"
	"os"
	"path/filepath"
	"time"
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

func RetryFunction(f func() error, maxAttempts int, delay time.Duration) error {
	var err error
	sleepyTime := delay
	for attempt := 0; attempt < maxAttempts; attempt++ {
		err = f()
		if err == nil {
			return nil
		}
		glog.Infof("retry failed,err:%s, retry times:%d\n", err.Error(), attempt+1)
		time.Sleep(sleepyTime)
		sleepyTime *= 2
	}

	return err
}
