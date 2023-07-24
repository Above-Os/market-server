package gitapp

import (
	"app-store-server/pkg/constants"
	"fmt"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/golang/glog"
)

var (
	appGit *AppGit
)

const (
	AppGitHttpsAddr = "https://github.com/Above-Os/terminus-apps.git"
	AppGitBranch    = "dev"
)

type AppGit struct {
	//todo save it to db
	lastCommitHash string
	lastCommitTime int64
}

func newAppGit() *AppGit {
	return &AppGit{}
}

func Init() error {
	appGit = newAppGit()
	//todo add retry
	return appGit.cloneCode()
}

func GetLastHash() string {
	if appGit == nil {
		return "not inited"
	}
	return appGit.lastCommitHash
}

func (ag *AppGit) updateLastHash(hash string) {
	ag.lastCommitHash = hash
	ag.lastCommitTime = time.Now().UnixNano()
	glog.Infof("lastCommitHash:%s", ag.lastCommitHash)
}

func (ag *AppGit) cloneCode() error {
	//clear local dir
	err := os.RemoveAll(constants.AppGitLocalDir)
	if err != nil {
		glog.Warningf("os.RemoveAll %s %s", constants.AppGitLocalDir, err.Error())
		return err
	}

	return ag.gitClone(AppGitHttpsAddr, AppGitBranch, constants.AppGitLocalDir)
}

func Pull() error {
	if appGit == nil {
		return fmt.Errorf("not inited")
	}
	return appGit.gitPull(constants.AppGitLocalDir)
}

func (ag *AppGit) gitClone(url, branch, directory string) error {
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

	ag.updateLastHash(commit.Hash.String())

	return nil
}

func (ag *AppGit) gitPull(directory string) error {
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

	ag.updateLastHash(commit.Hash.String())

	return err
}
