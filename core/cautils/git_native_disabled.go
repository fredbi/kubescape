//go:build !gitenabled

package cautils

import (
	"errors"
	"fmt"
	"io"
	"log"

	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	gitv5 "github.com/go-git/go-git/v5"
	configv5 "github.com/go-git/go-git/v5/config"
	plumbingv5 "github.com/go-git/go-git/v5/plumbing"
	objectv5 "github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/ioutil"
	"github.com/kubescape/go-git-url/apis"
)

var ErrFatalNotSupportedByBuild = errors.New(`git scan not supported by this build. Build with tag "gitenabled" to enable the git scan feature`)

type gitRepository struct {
	goGitRepo        *gitv5.Repository
	head             *plumbingv5.Reference
	config           *configv5.Config
	fileToLastCommit map[string]*objectv5.Commit
}

func newGitRepository(root string) (*gitRepository, error) {
	return &gitRepository{}, nil
}

func (g *gitRepository) Init(repo *gitv5.Repository, head *plumbingv5.Reference, config *configv5.Config) {
	g.goGitRepo = repo
	g.head = head
	g.config = config
}

func (g *gitRepository) GetFileLastCommit(filePath string) (*apis.Commit, error) {
	if len(g.fileToLastCommit) == 0 {
		filePathToCommitTime := map[string]time.Time{}
		filePathToCommit := map[string]*objectv5.Commit{}
		allCommits, _ := g.getAllCommits()

		// builds a map of all files to their last commit
		for _, commit := range allCommits {
			// Ignore merge commits (2+ parents)
			if commit.NumParents() <= 1 {
				tree, err := commit.Tree()
				if err != nil {
					continue
				}

				// ParentCount can be either 1 or 0 (initial commit)
				// In case it's the initial commit, prevTree is nil
				var prevTree *objectv5.Tree
				if commit.NumParents() == 1 {
					prevCommit, _ := commit.Parent(0)
					prevTree, err = prevCommit.Tree()
					if err != nil {
						continue
					}
				}

				changes, err := prevTree.Diff(tree)
				if err != nil {
					continue
				}

				for _, change := range changes {
					deltaFilePath := change.To.Name
					commitTime := commit.Author.When

					// In case we have the commit information for the file which is not the latest - we override it
					if currentCommitTime, exists := filePathToCommitTime[deltaFilePath]; exists {
						if currentCommitTime.Before(commitTime) {
							filePathToCommitTime[deltaFilePath] = commitTime
							filePathToCommit[deltaFilePath] = commit
						}
					} else {
						filePathToCommitTime[deltaFilePath] = commitTime
						filePathToCommit[deltaFilePath] = commit
					}
				}
			}
		}
		g.fileToLastCommit = filePathToCommit
	}

	if relevantCommit, exists := g.fileToLastCommit[filePath]; exists {
		return g.getCommit(relevantCommit), nil
	}

	return nil, fmt.Errorf("failed to get commit information for file: %s", filePath)
}

func (g *gitRepository) getAllCommits() ([]*objectv5.Commit, error) {
	ref, err := g.goGitRepo.Head()
	if err != nil {
		return nil, err
	}
	logItr, err := g.goGitRepo.Log(&gitv5.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, err
	}

	var allCommits []*objectv5.Commit
	err = logItr.ForEach(func(commit *objectv5.Commit) error {
		if commit != nil {
			allCommits = append(allCommits, commit)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return allCommits, nil
}

func (g *gitRepository) getCommit(commit *objectv5.Commit) *apis.Commit {
	return &apis.Commit{
		SHA: commit.Hash.String(),
		Author: apis.Committer{
			Name:  commit.Author.Name,
			Email: commit.Author.Email,
			Date:  commit.Author.When,
		},
		Message:   commit.Message,
		Committer: apis.Committer{},
		Files:     []apis.Files{},
	}
}

const gitDirName = ".git"

func dotGitToOSFilesystems(path string, detect bool) (dot, wt billy.Filesystem, err error) {
	if path, err = filepath.Abs(path); err != nil {
		return nil, nil, err
	}

	var fs billy.Filesystem
	var fi os.FileInfo
	for {
		fs = osfs.New(path)

		pathinfo, err := fs.Stat("/")
		if !os.IsNotExist(err) {
			if pathinfo == nil {
				return nil, nil, err
			}
			if !pathinfo.IsDir() && detect {
				fs = osfs.New(filepath.Dir(path))
			}
		}

		fi, err = fs.Stat(gitDirName)
		if err == nil {
			// no error; stop
			break
		}
		if !os.IsNotExist(err) {
			// unknown error; stop
			return nil, nil, err
		}
		if detect {
			// try its parent as long as we haven't reached
			// the root dir
			if dir := filepath.Dir(path); dir != path {
				path = dir
				continue
			}
		}
		// not detecting via parent dirs and the dir does not exist;
		// stop
		return fs, nil, nil
	}

	if fi.IsDir() {
		dot, err = fs.Chroot(gitDirName)
		return dot, fs, err
	}

	dot, err = dotGitFileToOSFilesystem(path, fs)
	if err != nil {
		return nil, nil, err
	}

	return dot, fs, nil
}

func dotGitFileToOSFilesystem(path string, fs billy.Filesystem) (bfs billy.Filesystem, err error) {
	f, err := fs.Open(gitDirName)
	if err != nil {
		return nil, err
	}
	defer ioutil.CheckClose(f, &err)

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	line := string(b)
	const prefix = "gitdir: "
	if !strings.HasPrefix(line, prefix) {
		return nil, fmt.Errorf(".git file has no %s prefix", prefix)
	}

	gitdir := strings.Split(line[len(prefix):], "\n")[0]
	gitdir = strings.TrimSpace(gitdir)
	if filepath.IsAbs(gitdir) {
		log.Printf("DEBUG(1): dotGit: %s", gitdir)
		return osfs.New(gitdir), nil
	}

	log.Printf("DEBUG(2): dotGit: %s/%s", path, gitdir)
	return osfs.New(fs.Join(path, gitdir)), nil
}
