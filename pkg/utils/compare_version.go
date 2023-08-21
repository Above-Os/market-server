package utils

import (
	"github.com/Masterminds/semver/v3"
	"github.com/golang/glog"
)

func NeedUpdate(curVersion, latestVersion string) bool {
	vCur, err := semver.NewVersion(curVersion)
	if err != nil {
		glog.Warningf("invalid curVersion:%s %s", curVersion, err.Error())
		return false
	}

	vLatest, err := semver.NewVersion(latestVersion)
	if err != nil {
		glog.Warningf("invalid latestVersion:%s %s", curVersion, err.Error())
		return false
	}

	//compare version
	if vCur.LessThan(vLatest) {
		return true
	}
	return false
}
