package app

import (
	"app-store-server/internal/constants"
	"app-store-server/internal/es"
	"app-store-server/internal/gitapp"
	"app-store-server/internal/helm"
	"app-store-server/internal/mongo"
	"app-store-server/pkg/models"
	"app-store-server/pkg/utils"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/golang/glog"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/api/resource"
	"text/template"
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

func packApp(app *models.ApplicationInfoEntry) *models.ApplicationInfoFullData {

	fullData := &models.ApplicationInfoFullData{
		Id:        app.Id,
		Name:      app.Name,
		History:   make(map[string]models.ApplicationInfoEntry),
		AppLabels: app.AppLabels,
	}

	fullData.History["latest"] = *app
	fullData.History[app.Version] = *app

	return fullData
}

func packApps(apps []*models.ApplicationInfoEntry) []*models.ApplicationInfoFullData {
	var fullDataList []*models.ApplicationInfoFullData

	for _, app := range apps {
		fullData := packApp(app)
		fullDataList = append(fullDataList, fullData)
	}

	return fullDataList
}

func UpdateAppInfosToDB() error {
	infos, err := GetAppInfosFromGitDir(constants.AppGitLocalDir)
	if err != nil {
		glog.Warningf("GetAppInfosFromGitDir %s err:%s", constants.AppGitLocalDir, err.Error())
		return err
	}

	glog.Infof("app infos size:%d", len(infos))

	apps := packApps(infos)

	glog.Infof("apps size:%d", len(apps))
	// just print one cell
	// var m models.ApplicationInfo
	// for _, info := range infos {
	// 	if info.Name == "firefox" {
	// 		m = *info
	// 	}
	// }
	// glog.Warningf("firefox: %v", m)

	err = UpdateAppInfosToMongo(apps)
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
		time.Sleep(time.Duration(5) * time.Minute)
		err := GitPullAndUpdate(false)
		if err != nil {
			glog.Warningf("%s", err.Error())
		}
	}
}

func GitPullAndUpdate(force bool) error {
	err := gitapp.Pull()
	if err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) && force {
			glog.Infof("info:%s", err.Error())
			glog.Infof("force update")
		} else {
			glog.Warningf("git pull err:%s", err.Error())
			return err
		}
	}

	glog.Infof("git repo update")

	err = gitapp.GetLastCommitHashAndUpdate()
	if err != nil {
		glog.Warningf("GetLastCommitHashAndUpdate err:%s", err.Error())
		return err
	}

	return UpdateAppInfosToDB()

	//todo check app infos in mongo if not exist in local, then del it
	//or del by lastCommitHash old
}

func UpdateAppInfosToMongo(infos []*models.ApplicationInfoFullData) error {
outerLoop:
	for _, info := range infos {

		for _, label := range info.AppLabels {
			if label == constants.DisableLabel {

				err := mongo.DisableAppInfoToDb(info)
				if err != nil {
					glog.Warningf("mongo.DisableAppInfoToDb info:%#v, err:%s", info, err.Error())
				}

				continue outerLoop
			}
		}

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

func ReadAppInfo(dirName string) (*models.ApplicationInfoEntry, error) {
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

	// First, attempt to parse the original configuration file
	var appCfg models.AppConfiguration
	if err = yaml.Unmarshal(cfgContent, &appCfg); err != nil {
		// Check if the file contains template syntax
		if strings.Contains(string(cfgContent), "{{") {
			// Need to render templates to generate two different configurations for administrators and regular users
			adminAppInfo, err := renderAppConfigWithTemplate(string(cfgContent), true)
			if err != nil {
				glog.Warningf("Failed to render admin application configuration: %s", err.Error())
				return nil, err
			}
			
			userAppInfo, err := renderAppConfigWithTemplate(string(cfgContent), false)
			if err != nil {
				glog.Warningf("Failed to render user application configuration: %s", err.Error())
				return nil, err
			}
			
			// Merge the two configurations to create an application information that contains both views
			mergedAppInfo := mergeAppInfos(adminAppInfo, userAppInfo)
			
			// Continue processing the merged application information
			disableCategories := getDisableCategories()
			for _, categorie := range mergedAppInfo.Categories {
				if strings.Contains(disableCategories, categorie) {
					glog.Warningf("%s %s is disable", categorie, mergedAppInfo.AppID)
					mergedAppInfo.AppLabels = append(mergedAppInfo.AppLabels, constants.DisableLabel)
				}
			}
			
			// Set i18n information
			setI18nInfo(mergedAppInfo, path.Join(constants.AppGitLocalDir, dirName))
			
			// Check for special files
			checkAppContainSpecialFile(mergedAppInfo, path.Join(constants.AppGitLocalDir, dirName))
			
			return mergedAppInfo, nil
		}
		
		glog.Warningf("Failed to parse application configuration: %s", err.Error())
		return nil, err
	}

	// Normal non-template processing flow
	appInfo := appInfoParseQuantity(appCfg.ToAppInfo())

	disableCategories := getDisableCategories()
	for _, categorie := range appInfo.Categories {
		if strings.Contains(disableCategories, categorie) {
			glog.Warningf("%s %s is disable", categorie, appInfo.AppID)
			appInfo.AppLabels = append(appInfo.AppLabels, constants.DisableLabel)
		}
	}

	// Set i18n information
	setI18nInfo(appInfo, path.Join(constants.AppGitLocalDir, dirName))

	checkAppContainSpecialFile(appInfo, path.Join(constants.AppGitLocalDir, dirName))

	return appInfo, nil
}

// Render application configuration with templates
func renderAppConfigWithTemplate(templateContent string, isAdmin bool) (*models.ApplicationInfoEntry, error) {
	// Create the values for template rendering
	values := map[string]interface{}{
		"Values": map[string]interface{}{
			"admin": "admin", // Default admin username
			"bfl": map[string]interface{}{
				"username": func() string {
					if isAdmin {
						return "admin"
					}
					return "user"
				}(), // Set different usernames based on role
			},
		},
	}
	
	// Create and render the template
	tmpl, err := template.New("appconfig").Parse(templateContent)
	if err != nil {
		return nil, err
	}
	
	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, values); err != nil {
		return nil, err
	}
	
	// Parse the rendered YAML
	var appCfg models.AppConfiguration
	if err := yaml.Unmarshal(rendered.Bytes(), &appCfg); err != nil {
		return nil, err
	}
	
	// Convert to application information entry
	appInfo := appInfoParseQuantity(appCfg.ToAppInfo())
	
	return appInfo, nil
}

// Merge two application information, connecting different parts with special markers
func mergeAppInfos(adminInfo, userInfo *models.ApplicationInfoEntry) *models.ApplicationInfoEntry {
	// Create a new application information based on the admin information
	mergedInfo := &models.ApplicationInfoEntry{}
	*mergedInfo = *adminInfo
	
	// Use special markers to concatenate different resource requirements
	if adminInfo.RequiredMemory != userInfo.RequiredMemory {
		mergedInfo.RequiredMemory = adminInfo.RequiredMemory + "||" + userInfo.RequiredMemory
	}
	
	if adminInfo.RequiredDisk != userInfo.RequiredDisk {
		mergedInfo.RequiredDisk = adminInfo.RequiredDisk + "||" + userInfo.RequiredDisk
	}
	
	if adminInfo.RequiredCPU != userInfo.RequiredCPU {
		mergedInfo.RequiredCPU = adminInfo.RequiredCPU + "||" + userInfo.RequiredCPU
	}
	
	if adminInfo.RequiredGPU != userInfo.RequiredGPU {
		mergedInfo.RequiredGPU = adminInfo.RequiredGPU + "||" + userInfo.RequiredGPU
	}

	// Handle other potentially different fields
	// For example, middleware, dependencies, etc.
	if !reflect.DeepEqual(adminInfo.Middleware, userInfo.Middleware) {
		// Custom logic is needed to merge the Middleware structure
		// Consider using JSON serialization and then connecting with special markers
		adminMiddleware, _ := json.Marshal(adminInfo.Middleware)
		userMiddleware, _ := json.Marshal(userInfo.Middleware)
		if len(adminMiddleware) > 0 && len(userMiddleware) > 0 {
			// Since we cannot directly create a new tapr.Middleware type, we will keep the middleware from adminInfo for now.
			// Record this difference for later processing.
			glog.Infof("The middleware configuration in the admin view and user view of the application %s is different", adminInfo.Name)
			// Save the differences in a separate comment
			// Cannot directly modify Middleware here due to type incompatibility
			
			// Optional: Add a custom field in the subsequent process to store these differences
			// For example: add a MiddlewareJSON field in the model and then set it:
			// mergedInfo.MiddlewareJSON = string(adminMiddleware) + "||" + string(userMiddleware)
		}
	}
	
	// Handle appScope and other options
	// Additional merging logic for more fields should be added based on actual requirements
	
	// Handle differences in the Options field
	// Check Dependencies in Options
	if !reflect.DeepEqual(adminInfo.Options.Dependencies, userInfo.Options.Dependencies) {
		glog.Infof("The dependency configuration in the admin view and user view of the application %s is different", adminInfo.Name)
		// Here you can consider merging the dependency lists or selecting a more comprehensive one.
		if len(adminInfo.Options.Dependencies) < len(userInfo.Options.Dependencies) {
			mergedInfo.Options.Dependencies = userInfo.Options.Dependencies
		}
	}
	
	// Check for differences in AppScope
	if !reflect.DeepEqual(adminInfo.Options.AppScope, userInfo.Options.AppScope) {
		glog.Infof("The AppScope configuration in the admin view and user view of the application %s is different", adminInfo.Name)
		// Typically choose the configuration with higher permissions.
		if adminInfo.Options.AppScope != nil && adminInfo.Options.AppScope.ClusterScoped {
			mergedInfo.Options.AppScope = adminInfo.Options.AppScope
		} else if userInfo.Options.AppScope != nil && userInfo.Options.AppScope.ClusterScoped {
			mergedInfo.Options.AppScope = userInfo.Options.AppScope
		}
	}
	
	// Check for differences in Policies
	if !reflect.DeepEqual(adminInfo.Options.Policies, userInfo.Options.Policies) {
		glog.Infof("The policy configuration in the admin view and user view of the application %s is different", adminInfo.Name)
		// Merge policies, keeping all unique policies from both sides
		policyMap := make(map[string]models.Policy)
		for _, policy := range adminInfo.Options.Policies {
			policyMap[policy.EntranceName] = policy
		}
		for _, policy := range userInfo.Options.Policies {
			if _, exists := policyMap[policy.EntranceName]; !exists {
				mergedInfo.Options.Policies = append(mergedInfo.Options.Policies, policy)
			}
		}
	}
	
	// Check for differences in Permissions
	if !reflect.DeepEqual(adminInfo.Permission, userInfo.Permission) {
		glog.Infof("The permission configuration in the admin view and user view of the application %s is different", adminInfo.Name)
		// Typically, choose the configuration with higher permissions.
		// Here we simply take adminInfo as the standard, but in practice, more complex merging logic may be needed.
	}
	
	// Check for differences in Entrances
	if !reflect.DeepEqual(adminInfo.Entrances, userInfo.Entrances) {
		glog.Infof("The entrance configuration in the admin view and user view of the application %s is different", adminInfo.Name)
		// Merge entrances, keeping all unique entrances from both sides
		entranceMap := make(map[string]bool)
		for _, entrance := range adminInfo.Entrances {
			entranceMap[entrance.Name] = true
		}
		for _, entrance := range userInfo.Entrances {
			if !entranceMap[entrance.Name] {
				mergedInfo.Entrances = append(mergedInfo.Entrances, entrance)
			}
		}
	}
	
	return mergedInfo
}

// Helper function to set i18n information
func setI18nInfo(appInfo *models.ApplicationInfoEntry, appDir string) {
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
}

func checkAppContainSpecialFile(info *models.ApplicationInfoEntry, appDir string) {
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

func appInfoParseQuantity(info *models.ApplicationInfoEntry) *models.ApplicationInfoEntry {
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

func GetAppInfosFromGitDir(dir string) (infos []*models.ApplicationInfoEntry, err error) {
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

func getGitInfosByName(appInfo *models.ApplicationInfoEntry, name string) {
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
