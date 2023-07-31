package helm

import (
	"helm.sh/helm/v3/pkg/repo"
	"path/filepath"

	"github.com/golang/glog"
	"helm.sh/helm/v3/pkg/action"
)

func PackageHelm(src, dstDir string) error {
	client := action.NewPackage()
	client.Destination = dstDir
	pathAbs, err := filepath.Abs(src)
	if err != nil {
		return err
	}

	p, err := client.Run(pathAbs, nil)
	if err != nil {
		return err
	}
	glog.Infof("src:%s, dstDir:%s Successfully packaged chart and saved it to: %s\n", src, dstDir, p)

	return nil
}

func IndexHelm(name, url, dir string) error {
	path, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	out := filepath.Join(path, "index.yaml")

	i, err := repo.IndexDirectory(path, url)
	if err != nil {
		return err
	}
	//merge to not implement

	i.SortEntries()
	return i.WriteFile(out, 0644)
}
