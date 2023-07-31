package gitapp

import (
	"app-store-server/internal/mongo"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/golang/glog"

	"app-store-server/internal/constants"
)

const (
	AppGitHttpsAddr = "https://github.com/Above-Os/terminus-apps.git"
	AppGitBranch    = "dev"
)

func Init() error {
	//todo add retry
	return cloneCode()
}

func GetLastHash() (string, error) {
	return mongo.GetLastCommitHashFromDB()
}

func updateLastHash(hash string) error {
	return mongo.SetLastCommitHashToDB(hash)
}

func cloneCode() error {
	//clear local dir
	err := os.RemoveAll(constants.AppGitLocalDir)
	if err != nil {
		glog.Warningf("os.RemoveAll %s %s", constants.AppGitLocalDir, err.Error())
		return err
	}

	err = os.RemoveAll(constants.AppGitZipLocalDir)
	if err != nil {
		glog.Warningf("os.RemoveAll %s %s", constants.AppGitLocalDir, err.Error())
		return err
	}

	return gitClone(AppGitHttpsAddr, AppGitBranch, constants.AppGitLocalDir)
}

func Pull() error {
	return gitPull(constants.AppGitLocalDir)
}

func gitClone(url, branch, directory string) error {
	glog.Infof("git clone %s %s %s--recursive", url, branch, directory)

	//git.DefaultSubmoduleRecursionDepth
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

func gitPull(directory string) error {
	r, err := git.PlainOpen(directory)
	if err != nil {
		glog.Warningf("err:%s", err.Error())
		return err
	}

	// Get the working directory for the repository
	w, err := r.Worktree()
	if err != nil {
		glog.Warningf("err:%s", err.Error())
		return err
	}

	// Pull the latest changes from the origin remote and merge into the current branch
	glog.Infof("git pull origin")
	err = w.Pull(&git.PullOptions{RemoteName: "origin"})
	if err != nil {
		glog.Warningf("err:%s", err.Error())
		return err
	}

	// Print the latest commit that was just pulled
	ref, err := r.Head()
	if err != nil {
		glog.Warningf("err:%s", err.Error())
		return err
	}
	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		glog.Warningf("err:%s", err.Error())
		return err
	}
	glog.Infof("commit:%#v", commit)

	updateLastHash(commit.Hash.String())

	return err
}

func GetCreateTimeSecond(dirPath, subDirPath string) (int64, error) {
	timeStr, err := getCreateTime(dirPath, subDirPath)
	if err != nil {
		return 0, err
	}

	t, err := time.Parse("Mon Jan 02 15:04:05 2006 -0700", timeStr)
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

	t, err := time.Parse("Mon Jan 02 15:04:05 2006 -0700", timeStr)
	if err != nil {
		return 0, err
	}

	return t.Unix(), nil
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
