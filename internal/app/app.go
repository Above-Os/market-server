package app

import (
	"app-store-server/internal/constants"
	"app-store-server/internal/es"
	"app-store-server/internal/gitapp"
	"app-store-server/internal/helm"
	"app-store-server/internal/mongo"
	"app-store-server/pkg/models"
	"app-store-server/pkg/utils"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/golang/glog"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	DisableCategoriesEnv = "DISABLE_CATEGORIES"
)

func getDisableCategories() string {
	disableCategories := os.Getenv(DisableCategoriesEnv)
	if disableCategories != "" {
		return disableCategories
	}

	return ""
}

func Init() error {
	err := UpdateAppInfosToDB()
	if err != nil {
		glog.Warningf("%s", err.Error())
		return err
	}

	go pullAndUpdateLoop()

	return nil
}

func UpdateAppInfosToDB() error {
	infos, err := GetAppInfosFromGitDir(constants.AppGitLocalDir)
	if err != nil {
		glog.Warningf("GetAppInfosFromGitDir %s err:%s", constants.AppGitLocalDir, err.Error())
		return err
	}
	var m models.ApplicationInfo
	for _, info := range infos {
		if info.Name == "firefox" {
			m = *info
		}
	}
	glog.Warningf("firefox: %v", m)

	err = UpdateAppInfosToMongo(infos)
	if err != nil {
		glog.Warningf("%s", err.Error())
		return err
	}
	glog.Infof("save to mongo success")

	//sync info from mongodb to es
	go es.SyncInfoFromMongo()

	return nil
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
	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		glog.Infof("info:%s", err.Error())
		return nil
	}
	if err != nil {
		glog.Warningf("git pull err:%s", err.Error())
		return err
	}

	err = gitapp.GetLastCommitHashAndUpdate()
	if err != nil {
		glog.Warningf("GetLastCommitHashAndUpdate err:%s", err.Error())
		return err
	}

	return UpdateAppInfosToDB()

	//todo check app infos in mongo if not exist in local, then del it
	//or del by lastCommitHash old
}

func UpdateAppInfosToMongo(infos []*models.ApplicationInfo) error {
	for _, info := range infos {
		err := mongo.UpsertAppInfoToDb(info)
		if err != nil {
			glog.Warningf("mongo.UpsertAppInfoToDb info:%#v, err:%s", info, err.Error())
		}

		err = mongo.InitCounterByApp(info.Name)
		if err != nil {
			glog.Warningf("mongo.InitCounterByApp info.Name:%s, err:%s", info.Name, err.Error())
		}
	}

	//todo delete expired information

	return nil
}

func ReadAppInfo(dirName string) (*models.ApplicationInfo, error) {
	cfgFileName := path.Join(constants.AppGitLocalDir, dirName, constants.AppCfgFileName)

	f, err := os.Open(cfgFileName)
	if err != nil {
		glog.Warningf("%s", err.Error())
		return nil, err
	}

	cfgContent, err := io.ReadAll(f)
	if err != nil {
		glog.Warningf("%s", err.Error())
		return nil, err
	}

	var appCfg models.AppConfiguration
	if err = yaml.Unmarshal(cfgContent, &appCfg); err != nil {
		glog.Warningf("%s", err.Error())
		return nil, err
	}

	appInfo := appInfoParseQuantity(appCfg.ToAppInfo())

	disableCategories := getDisableCategories()
	for _, categorie := range appInfo.Categories {
		if strings.Contains(disableCategories, categorie) {
			glog.Warningf("%s is disable", categorie)
			return nil, errors.New("disabled")
		}
	}

	// set i18n info
	appDir := path.Join(constants.AppGitLocalDir, dirName)

	glog.Infof("---->start parse i18n<----")
	i18nMap := make(map[string]models.I18n)
	for _, lang := range appInfo.Locale {
		glog.Infof("path:")
		glog.Infof(appDir)
		glog.Infof(lang)
		glog.Infof(constants.AppCfgFileName)

		data, err := ioutil.ReadFile(path.Join(appDir, "i18n", lang, constants.AppCfgFileName))
		if err != nil {
			glog.Warningf("failed to get file %s,err=%v", path.Join("i18n", lang, constants.AppCfgFileName), err)
			continue
		}
		glog.Infof("data:")
		glog.Infof(string(data))

		var i18n models.I18n
		err = yaml.Unmarshal(data, &i18n)
		if err != nil {
			glog.Warningf("unmarshal to I18n failed err=%v", err)
			continue
		}
		fmt.Println(i18n)
		i18nMap[lang] = i18n

	}
	appInfo.I18n = i18nMap
	glog.Infof("---->end parse i18n<----")

	checkAppContainSpecialFile(appInfo, appDir)

	return appInfo, nil
}

func checkAppContainSpecialFile(info *models.ApplicationInfo, appDir string) {
	if isContainRemove(appDir) {
		info.AppLabels = append(info.AppLabels, constants.RemoveLabel)
	}

	if isContainSuspend(appDir) {
		info.AppLabels = append(info.AppLabels, constants.SuspendLabel)
	}

	if isContainNsfw(appDir) {
		info.AppLabels = append(info.AppLabels, constants.NsfwLabel)
	}
}

func isContainSuspend(appDir string) bool {
	return utils.IsDirContainFile(appDir, constants.SuspendFile)
}

func isContainRemove(appDir string) bool {
	return utils.IsDirContainFile(appDir, constants.RemoveFile)
}

func isContainNsfw(appDir string) bool {
	return utils.IsDirContainFile(appDir, constants.NsfwFile)
}

func appInfoParseQuantity(info *models.ApplicationInfo) *models.ApplicationInfo {
	if info == nil {
		return info
	}

	if info.RequiredMemory != "" {
		r, err := resource.ParseQuantity(info.RequiredMemory)
		if err == nil {
			info.RequiredMemory = fmt.Sprintf("%d", int(r.AsApproximateFloat64()))
			//info.RequiredMemory = fmt.Sprintf("%d", int(r.AsApproximateFloat64()/1024/1024))
		}
	}

	if info.RequiredDisk != "" {
		r, err := resource.ParseQuantity(info.RequiredDisk)
		if err == nil {
			info.RequiredDisk = fmt.Sprintf("%d", int(r.AsApproximateFloat64()))
			//info.RequiredDisk = fmt.Sprintf("%d", int(r.AsApproximateFloat64()/1024/1024))
		}
	}

	if info.RequiredGPU != "" {
		r, err := resource.ParseQuantity(info.RequiredGPU)
		if err == nil {
			info.RequiredGPU = fmt.Sprintf("%d", int(r.AsApproximateFloat64()))
			//info.RequiredGPU = fmt.Sprintf("%d", int(r.AsApproximateFloat64()/1024/1024/1024))
		}
	}

	if info.RequiredCPU != "" {
		r, err := resource.ParseQuantity(info.RequiredCPU)
		if err == nil {
			info.RequiredCPU = fmt.Sprintf("%v", r.AsApproximateFloat64())
		}
	}

	return info
}

func GetAppInfosFromGitDir(dir string) (infos []*models.ApplicationInfo, err error) {
	charts, err := os.ReadDir(dir)
	if err != nil {
		glog.Warningf("read dir %s error: %s", dir, err.Error())
		return nil, err
	}

	for _, c := range charts {
		if !c.IsDir() || strings.HasPrefix(c.Name(), ".") {
			continue
		}

		// read app info from chart
		appInfo, err := ReadAppInfo(c.Name())
		if err != nil {
			glog.Warningf("app chart %s reading error: %s", c.Name(), err.Error())
			continue
		}

		//helm package
		appInfo.ChartName, err = helmPackage(c.Name())
		if err != nil {
			glog.Warningf("helm package %s error: %s", c.Name(), err.Error())
			continue
		}

		glog.Infof("name:%s, version:%s, chartName:%s", c.Name(), appInfo.Version, appInfo.ChartName)

		//get git info
		getGitInfosByName(appInfo, c.Name())
		infos = append(infos, appInfo)
	}

	return infos, nil
}

func getGitInfosByName(appInfo *models.ApplicationInfo, name string) {
	var err error
	appInfo.LastCommitHash, err = gitapp.GetLastHash()
	if err != nil {
		glog.Warningf("GetLastHash error: %s", err.Error())
	}

	appInfo.CreateTime, err = gitapp.GetCreateTimeSecond(constants.AppGitLocalDir, name)
	if err != nil {
		glog.Warningf("GetCreateTimeSecond error: %s", err.Error())
	}

	appInfo.UpdateTime, err = gitapp.GetLastUpdateTimeSecond(constants.AppGitLocalDir, name)
	if err != nil {
		glog.Warningf("GetLastUpdateTimeSecond error: %s", err.Error())
	}
}

func helmPackage(name string) (string, error) {
	src := path.Join(constants.AppGitLocalDir, name)
	return helm.PackageHelm(src, constants.AppGitZipLocalDir)
}
