package gitapp

import (
	"app-store-server/internal/constants"
	"app-store-server/internal/mongo"
	"app-store-server/pkg/utils"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/golang/glog"
)

const (
	AppGitHttpsAddr = "https://github.com/Above-Os/terminus-apps.git"
	AppGitBranch    = "main"

	GitAddrEnv   = "GIT_ADDR"
	GitBranchEnv = "GIT_BRANCH"
)

func getGitAddr() string {
	gitAddr := os.Getenv(GitAddrEnv)
	if gitAddr != "" {
		return gitAddr
	}

	return AppGitHttpsAddr
}

func getGitBranch() string {
	gitBranch := os.Getenv(GitBranchEnv)
	if gitBranch != "" {
		return gitBranch
	}
	return AppGitBranch
}

func Init() error {
	return utils.RetryFunction(cloneCode, 3, time.Second)
}

func cloneCode() error {
	//clear local git dir
	err := os.RemoveAll(constants.AppGitLocalDir)
	if err != nil {
		glog.Warningf("os.RemoveAll %s %s", constants.AppGitLocalDir, err.Error())
		return err
	}

	//clear local charts dir
	// err = os.RemoveAll(constants.AppGitZipLocalDir)
	// if err != nil {
	// 	glog.Warningf("os.RemoveAll %s %s", constants.AppGitZipLocalDir, err.Error())
	// 	return err
	// }

	return gitClone(getGitAddr(), getGitBranch(), constants.AppGitLocalDir)
}

func gitClone(url, branch, directory string) error {
	glog.Infof("git clone %s %s %s --recursive", url, branch, directory)

	//clone
	r, err := git.PlainClone(directory, false, &git.CloneOptions{
		URL:           url,
		Progress:      os.Stdout,
		ReferenceName: plumbing.ReferenceName(branch),
	})
	if err != nil {
		glog.Warningf("err:%s", err.Error())
		return err
	}

	ref, err := r.Head()
	if err != nil {
		glog.Warningf("err:%s", err.Error())
		return err
	}
	glog.Infof("ref:%#v", ref)

	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		glog.Warningf("err:%s", err.Error())
		return err
	}
	glog.Infof("commit:%#v", commit)

	updateLastHash(commit.Hash.String())

	return nil
}

func Pull() error {
	return gitPull(constants.AppGitLocalDir)
}

// gitPull attempts to pull using V3 first, falls back to V2 and then V1 if previous attempts fail
func gitPull(directory string) error {
	err := gitPullV3(directory)
	if err != nil {
		glog.Warningf("gitPullV3 failed, fallback to V2: %v", err)
		err = gitPullV2(directory)
		if err != nil {
			glog.Warningf("gitPullV2 failed, fallback to V1: %v", err)
			return gitPullV1(directory)
		}
	}
	return nil
}

// gitPullV1 is the original implementation
func gitPullV1(directory string) error {
	curDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(curDir)

	dir, err := filepath.Abs(directory)
	if err != nil {
		return err
	}

	err = os.Chdir(dir)
	if err != nil {
		return err
	}

	cmd := exec.Command("git", "pull")
	out, err := cmd.CombinedOutput()
	if err != nil {
		glog.Infof("combined out:%s\n", string(out))
		return err
	}

	if strings.Contains(string(out), "Already up to date") {
		return git.NoErrAlreadyUpToDate
	}
	glog.Infof("out:%s\n", string(out))

	return nil
}

// gitPullV2 is the implementation with enhanced error handling
func gitPullV2(directory string) error {
	// Get absolute path
	dir, err := filepath.Abs(directory)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if directory exists and is accessible
	if _, err := os.Stat(dir); err != nil {
		return fmt.Errorf("directory is not accessible: %w", err)
	}

	curDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}
	
	// Ensure we restore the original working directory when function returns
	defer func() {
		if err := os.Chdir(curDir); err != nil {
			glog.Warningf("failed to restore working directory: %v", err)
		}
	}()

	// Change to target directory
	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("failed to change working directory: %w", err)
	}

	cmd := exec.Command("git", "pull")
	out, err := cmd.CombinedOutput()
	if err != nil {
		glog.Infof("git pull output: %s\n", string(out))
		return fmt.Errorf("git pull failed: %w", err)
	}

	if strings.Contains(string(out), "Already up to date") {
		return git.NoErrAlreadyUpToDate
	}
	glog.Infof("git pull output: %s\n", string(out))

	return nil
}

// gitPullV3 implements pull using go-git library
func gitPullV3(directory string) error {
	// Open existing repository
	r, err := git.PlainOpen(directory)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Get the working directory
	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Configure pull options
	opts := &git.PullOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
		Force:      true,
	}

	// Perform pull
	err = w.Pull(opts)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("pull failed: %w", err)
	}

	return nil
}

func AppDirExist(name string) bool {
	filePath := path.Join(constants.AppGitLocalDir, name)
	exist, err := utils.PathExists(filePath)
	if err != nil {
		glog.Warningf("utils.PathExists %s %s", filePath, err.Error())
	}

	return exist
}

func ReadMe(name string) ([]byte, error) {
	exist := AppDirExist(name)
	if !exist {
		return nil, fmt.Errorf("%s not exist", name)
	}

	readFileName := path.Join(constants.AppGitLocalDir, name, constants.ReadmeFileName)

	f, err := os.Open(readFileName)
	if err != nil {
		glog.Warningf("%s", err.Error())
		return nil, err
	}

	content, err := io.ReadAll(f)
	if err != nil {
		glog.Warningf("%s", err.Error())
		return nil, err
	}

	return content, nil
}

func updateLastHash(hash string) error {
	return mongo.SetLastCommitHashToDB(hash)
}

func GetLastHash() (hash string, err error) {
	hash, err = mongo.GetLastCommitHashFromDB()
	if err == nil && hash != "" {
		return hash, nil
	}

	return getGitLastCommitHash(constants.AppGitLocalDir)
}

func GetLastCommitHashAndUpdate() error {
	hash, err := getGitLastCommitHash(constants.AppGitLocalDir)
	if err != nil {
		glog.Warningf("getGitLastCommitHash err:%s", err.Error())
		return err
	}
	glog.Warningf("git hash:%s", hash)

	err = updateLastHash(hash)
	if err != nil {
		glog.Warningf("updateLastHash err:%s", err.Error())
		return err
	}

	return nil
}

func getGitLastCommitHash(directory string) (string, error) {
	r, err := git.PlainOpen(directory)
	if err != nil {
		glog.Warningf("err:%s", err.Error())
		return "", err
	}

	ref, err := r.Head()
	if err != nil {
		glog.Warningf("err:%s", err.Error())
		return "", err
	}
	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		glog.Warningf("err:%s", err.Error())
		return "", err
	}
	// Print the latest commit that was just pulled
	glog.Infof("commit:%#v", commit)

	return commit.Hash.String(), nil
}

func GetCreateTimeSecond(dirPath, subDirPath string) (int64, error) {
	timeStr, err := getCreateTime(dirPath, subDirPath)
	if err != nil {
		return 0, err
	}

	t, err := time.Parse(constants.TimeFormatStr, timeStr)
	if err != nil {
		return 0, err
	}

	return t.Unix(), nil
}

func GetLastUpdateTimeSecond(dirPath, subDirPath string) (int64, error) {
	timeStr, err := getLastUpdateTime(dirPath, subDirPath)
	if err != nil {
		return 0, err
	}

	t, err := time.Parse(constants.TimeFormatStr, timeStr)
	if err != nil {
		return 0, err
	}

	return t.Unix(), nil
}

func getLastUpdateTime(dirPath, subDirPath string) (string, error) {
	curDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	defer os.Chdir(curDir)

	dir, err := filepath.Abs(dirPath)
	if err != nil {
		return "", err
	}

	err = os.Chdir(dir)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("git", "log", "-1", `--pretty=format:%ad`, `--`, subDirPath)
	glog.Infof("cmd:%s", cmd.String())

	out, err := cmd.CombinedOutput()
	if err != nil {
		glog.Warningf("combined out:%s\n", string(out))
		return "", err
	}
	glog.Infof("out:%s", string(out))

	return string(out), nil
}

func getCreateTime(dirPath, subDirPath string) (string, error) {
	curDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	defer os.Chdir(curDir)

	dir, err := filepath.Abs(dirPath)
	if err != nil {
		return "", err
	}
	err = os.Chdir(dir)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("git", "log", `--pretty=format:%ad`, `--`, subDirPath)
	glog.Infof("cmd:%s", cmd.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		glog.Warningf("combined out:%s\n", string(out))
		return "", err
	}

	outStr := string(out)
	glog.Infof("out:%s", outStr)
	outStrSlice := strings.Split(outStr, "\n")
	if len(outStrSlice) <= 0 {
		return "", fmt.Errorf("%s not contain \\n", outStr)
	}

	return outStrSlice[len(outStrSlice)-1], nil
}
