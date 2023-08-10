package helm

import (
	"path"
	"path/filepath"

	"github.com/golang/glog"
	"helm.sh/helm/v3/pkg/action"
)

func PackageHelm(src, dstDir string) (string, error) {
	client := action.NewPackage()
	client.Destination = dstDir
	pathAbs, err := filepath.Abs(src)
	if err != nil {
		return "", err
	}

	var p string
	p, err = client.Run(pathAbs, nil)
	if err != nil {
		return "", err
	}
	glog.Infof("src:%s, dstDir:%s Successfully packaged chart and saved it to: %s\n", src, dstDir, p)

	return path.Base(p), nil
}
