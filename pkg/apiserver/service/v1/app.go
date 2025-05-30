package v1

import (
	"app-store-server/internal/app"
	"app-store-server/internal/constants"
	"app-store-server/internal/es"
	"app-store-server/internal/mongo"
	"app-store-server/pkg/models"
	"app-store-server/pkg/utils"
	"fmt"
	"path"

	"github.com/Masterminds/semver/v3"
	"github.com/golang/glog"
)

func getChartPath(appName string, version string) string {
	filePathName := path.Join(constants.AppGitZipLocalDir, appName)
	if exist, _ := utils.PathExists(filePathName); exist {
		return filePathName
	}

	//not a chart name, search chart name
	info, err := getInfoByName(appName)
	app, err := filterVersionForApp(info, version)

	if err == nil && app.ChartName != "" {
		filePathName = path.Join(constants.AppGitZipLocalDir, app.ChartName)
		if exist, _ := utils.PathExists(filePathName); exist {
			return filePathName
		}
	}

	return ""
}

// getChartPath gets the chart file path by fileName only
func getChartPathNoVersion(fileName string) string {
	return path.Join(constants.AppGitZipLocalDir, fileName)
}

func getInfoByName(appName string) (*models.ApplicationInfoFullData, error) {
	info, err := es.SearchByNameAccurate(appName)
	if err == nil && info != nil {
		return info, nil
	}

	info, err = mongo.GetAppInfoByName(appName)
	if err == nil && info != nil {
		return info, nil
	}

	latest, err := app.ReadAppInfo(appName)

	info = &models.ApplicationInfoFullData{}
	info.History["latest"] = *latest
	info.Name = latest.Name
	info.AppLabels = latest.AppLabels

	return info, err
}

func pickVersionForAppsWithMap(apps map[string]*models.ApplicationInfoFullData, version string) (map[string]*models.ApplicationInfoEntry, error) {
	mapInfo := make(map[string]*models.ApplicationInfoEntry)

	// Parse the passed version string
	v, err := semver.NewVersion(version)
	if err != nil {
		return nil, err
	}

	for appName, app := range apps {
		var maxEntry *models.ApplicationInfoEntry
		latestEntry := app.History["latest"]

		// Traversing the history
		for _, entry := range app.History {
			// Traversing dependencies
			for _, dep := range entry.Options.Dependencies {
				if dep.Name == "olares" && dep.Type == "system" {
					// Parsing version strings
					constraint, err := semver.NewConstraint(dep.Version)
					if err != nil {
						return nil, err
					}

					// Check if the version meets the constraints
					if constraint.Check(v) {
						appv, err := semver.NewVersion(entry.Version)
						if err != nil {
							// return nil, err
							continue
						}

						// If the conditions are met, check whether it is the largest version
						if maxEntry == nil || appv.GreaterThan(semver.MustParse(maxEntry.Version)) {
							entryCopy := entry
							maxEntry = &entryCopy

						}
					}
				}
			}
		}

		// If a matching version is found, add it to the map; otherwise, add the latest entry
		if maxEntry != nil {
			mapInfo[appName] = maxEntry
		} else {
			mapInfo[appName] = &latestEntry
		}
	}

	return mapInfo, nil
}

func pickVersionForApps(apps []*models.ApplicationInfoFullData, version string) ([]models.ApplicationInfoEntry, error) {
	var result []models.ApplicationInfoEntry

	// Parse the passed version string
	v, err := semver.NewVersion(version)
	if err != nil {
		return nil, err
	}

	for _, app := range apps {
		var maxEntry *models.ApplicationInfoEntry
		latestEntry := app.History["latest"]

		// Traversing the history
		for _, entry := range app.History {
			// Traversing dependencies
			for _, dep := range entry.Options.Dependencies {
				if dep.Name == "olares" && dep.Type == "system" {
					// Parsing version strings
					constraint, err := semver.NewConstraint(dep.Version)
					if err != nil {
						return nil, err
					}

					// Check if the version meets the constraints
					if constraint.Check(v) {
						appv, err := semver.NewVersion(entry.Version)
						if err != nil {
							// return result, err
							continue
						}

						// If the conditions are met, check whether it is the largest version
						if maxEntry == nil || appv.GreaterThan(semver.MustParse(maxEntry.Version)) {
							entryCopy := entry
							maxEntry = &entryCopy

						}
					}
				}
			}
		}

		// If a matching version is found, add it to the result; otherwise, add the latest entry
		if maxEntry != nil {
			result = append(result, *maxEntry)
		} else {
			result = append(result, latestEntry)
		}
	}

	return result, nil
}

func filterVersionForApps(apps []*models.ApplicationInfoFullData, version string) ([]models.ApplicationInfoEntry, error) {
	var result []models.ApplicationInfoEntry

	// Parse the passed version string
	v, err := semver.NewVersion(version)
	if err != nil {
		glog.Infof("error version:%s", version)
		return nil, err
	}

	for _, app := range apps {
		var maxEntry *models.ApplicationInfoEntry

		// Traversing the history
		for _, entry := range app.History {
			// Traversing dependencies
			for _, dep := range entry.Options.Dependencies {
				if dep.Name == "olares" && dep.Type == "system" {
					// Parsing version strings
					constraint, err := semver.NewConstraint(dep.Version)
					if err != nil {
						glog.Infof("error version:%s", dep.Version)
						return nil, err
					}

					// Check if the version meets the constraints
					if constraint.Check(v) {

						appv, err := semver.NewVersion(entry.Version)
						if err != nil {
							glog.Infof("error version:%s, error app:%s", entry.Version, entry.Name)
							// return result, err
							continue
						}

						if entry.Name == "ollama" {
							glog.Infof("update app:%s, this:%s", entry.Name, entry.Version)
							if maxEntry != nil {
								glog.Infof("update app:%s, max now:%s", entry.Name, maxEntry.Version)
							}
							glog.Infof("update app:%s, dep.Version:%s, v:%s", entry.Name, dep.Version, v)
						}

						// If the conditions are met, check whether it is the largest version
						if maxEntry == nil || appv.GreaterThan(semver.MustParse(maxEntry.Version)) {
							entryCopy := entry
							maxEntry = &entryCopy
							glog.Infof("update app:%s, new:%s", entry.Name, maxEntry.Version)
						}
					}
				}
			}
		}

		// If the largest version matching the criteria is found, add it to the result
		if maxEntry != nil {
			result = append(result, *maxEntry)
		}
	}

	return result, nil
}

func filterVersionForApp(app *models.ApplicationInfoFullData, version string) (models.ApplicationInfoEntry, error) {
	var result models.ApplicationInfoEntry
	var maxEntry *models.ApplicationInfoEntry

	// Parse the passed version string
	v, err := semver.NewVersion(version)
	if err != nil {
		return result, err
	}

	// Traversing the application history
	for _, entry := range app.History {
		// Traversing dependencies
		for _, dep := range entry.Options.Dependencies {
			if dep.Name == "olares" && dep.Type == "system" {
				// Parsing version strings
				constraint, err := semver.NewConstraint(dep.Version)
				if err != nil {
					return result, err
				}

				// Check if the version meets the constraints
				if constraint.Check(v) {
					// Parse the passed version string
					appv, err := semver.NewVersion(entry.Version)
					if err != nil {
						// return result, err
						continue
					}

					// If the condition is met, check whether it is the largest version
					if maxEntry == nil || appv.GreaterThan(semver.MustParse(maxEntry.Version)) {
						entryCopy := entry
						maxEntry = &entryCopy

					}
				}
			}
		}
	}

	// If the largest version matching the criteria is found, it is returned
	if maxEntry != nil {
		return *maxEntry, nil
	}

	return result, fmt.Errorf("no matching version found")
}
