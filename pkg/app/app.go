package app

import (
	"app-store-server/pkg/constants"
	"app-store-server/pkg/gitapp"
	"app-store-server/pkg/models"
	"app-store-server/pkg/mongo"
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
		glog.Warningf("name:%s, version:%s\n", c.Name(), appInfo.Version)

		//zip
		err = zipApp(c.Name(), appInfo.Version)
		if err != nil {
			glog.Warningf("app chart reading error: %s", err.Error())
			continue
		}
		//update info to db
		appInfo.LastCommitHash = gitapp.GetLastHash()
		infos = append(infos, appInfo)
	}

	return infos, nil
}

func zipApp(name, version string) error {
	src := fmt.Sprintf("%s/%s", constants.AppGitLocalDir, name)
	dst := fmt.Sprintf("%s/%s-%s.tgz", constants.AppGitZipLocalDir, name, version)

	err := utils.Tar(src, dst)
	return err
}
