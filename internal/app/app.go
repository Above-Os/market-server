package app

import (
	"app-store-server/internal/constants"
	"app-store-server/internal/gitapp"
	"app-store-server/internal/helm"
	"app-store-server/internal/mongo"
	"app-store-server/pkg/models"
	"app-store-server/pkg/utils"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/golang/glog"
	"gopkg.in/yaml.v3"
)

func Init() error {
	err := UpdateAppInfosToMongo()
	if err != nil {
		glog.Warningf("%s", err.Error())
	}

	go pullAndUpdateLoop()

	return err
}

func pullAndUpdateLoop() {
	for {
		time.Sleep(time.Duration(1) * time.Minute)
		err := GitPullAndUpdate()
		if err != nil {
			glog.Warningf("%s", err.Error())
		}
	}
}

func GitPullAndUpdate() error {
	err := gitapp.Pull()
	if err == git.NoErrAlreadyUpToDate {
		glog.Infof("info:%s", err.Error())
		return nil
	}
	if err != nil {
		glog.Warningf("%s", err.Error())
		return err
	}

	return UpdateAppInfosToMongo()

	//todo check app infos in mongo if not exist in local, then del it
	//or del by lastCommitHash old
}

func readAppInfo(dir fs.FileInfo) (*models.ApplicationInfo, error) {
	cfgFileName := fmt.Sprintf("%s/%s/%s", constants.AppGitLocalDir, dir.Name(), constants.AppCfgFileName)

	f, err := os.Open(cfgFileName)
	if err != nil {
		glog.Warningf("%s", err.Error())
		return nil, err
	}

	info, err := ioutil.ReadAll(f)
	if err != nil {
		glog.Warningf("%s", err.Error())
		return nil, err
	}

	var appCfg models.AppConfiguration
	if err = yaml.Unmarshal(info, &appCfg); err != nil {
		glog.Warningf("%s", err.Error())
		return nil, err
	}

	// cache app icon data
	// var icon string
	// iconData, found := imageCache.Get(appCfg.Metadata.Icon)
	// if !found {
	// 	icon, err = readImageToBase64(fmt.Sprintf("%s/%s", ChartsPath, dir.Name()), appCfg.Metadata.Icon)
	// 	if err != nil {
	// 		klog.Errorf("get app icon error: %s", err)
	// 	} else {
	// 		imageCache.Set(appCfg.Metadata.Icon, icon, cache.DefaultExpiration)
	// 	}
	// } else {
	// 	icon = iconData.(string)
	// }

	return appCfg.ToAppInfo(), nil
}

func UpdateAppInfosToMongo() error {
	infos, err := GetAppInfosFromGitDir(constants.AppGitLocalDir)
	if err != nil {
		return err
	}

	for _, info := range infos {
		err = mongo.UpsertAppInfoToDb(info)
		if err != nil {
			glog.Warningf("mongo.UpsertAppInfoToDb err:%s", err.Error())
			continue
		}
	}

	return nil
}

func GetAppInfosFromGitDir(dir string) (infos []*models.ApplicationInfo, err error) {
	charts, err := ioutil.ReadDir(dir)
	if err != nil {
		glog.Warningf("read dir %s error: %s", dir, err.Error())
		return nil, err
	}

	for _, c := range charts {
		if !c.IsDir() {
			continue
		}

		if strings.HasPrefix(c.Name(), ".") {
			continue
		}

		// read app info from chart
		appInfo, err := readAppInfo(c)
		if err != nil {
			glog.Warningf("app chart reading error: %s", err.Error())
			continue
		}

		//zip
		//err = zipApp(c.Name(), appInfo.Version)
		appInfo.ChartName, err = helmPackage(c.Name())
		if err != nil {
			glog.Warningf("app chart reading error: %s", err.Error())
			continue
		}

		glog.Infof("name:%s, version:%s, chartName:%s\n", c.Name(), appInfo.Version, appInfo.ChartName)

		//update info to db
		appInfo.LastCommitHash, err = gitapp.GetLastHash()
		if err != nil {
			glog.Warningf("GetLastHash error: %s", err.Error())
		}
		appInfo.CreateTime, err = gitapp.GetCreateTimeSecond(constants.AppGitLocalDir, c.Name())
		if err != nil {
			glog.Warningf("GetCreateTimeSecond error: %s", err.Error())
		}
		appInfo.UpdateTime, err = gitapp.GetLastUpdateTimeSecond(constants.AppGitLocalDir, c.Name())
		if err != nil {
			glog.Warningf("GetLastUpdateTimeSecond error: %s", err.Error())
		}
		infos = append(infos, appInfo)
	}

	err = helm.IndexHelm(constants.RepoName, constants.RepoUrl, constants.AppGitZipLocalDir)
	if err != nil {
		glog.Warningf("IndexHelm error: %s", err.Error())
		return infos, err
	}
	return infos, nil
}

func helmPackage(name string) (string, error) {
	src := fmt.Sprintf("%s/%s", constants.AppGitLocalDir, name)
	return helm.PackageHelm(src, constants.AppGitZipLocalDir)
}

func zipApp(name, version string) error {
	src := fmt.Sprintf("%s/%s", constants.AppGitLocalDir, name)
	dst := fmt.Sprintf("%s/%s-%s.tgz", constants.AppGitZipLocalDir, name, version)

	err := utils.Tar(src, dst)
	return err
}
