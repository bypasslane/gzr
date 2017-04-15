package comms

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// GitManager is an interface to retrieve data from a git repo
type GitManager interface {
	CommitHash() (string, error)
	Remote() (string, error)
	Tags() ([]string, []string, error)
	SetPath(string)
	GetPath() string
}

// LocalGitManager implements GitManager
type LocalGitManager struct {
	path string
}

const (
	failedToGetwdMsg           = "failed to get working directory from OS"
	failedToGetCommitHashMsg   = "failed to get commit hash"
	failedToGetRemoteMsg       = "failed to get remote of git repo"
	failedToChangeDirectoryMsg = "failed to change directory to %q."
)

// NewImageMetadata returns a populated ImageMetadata based on a LocalGitManager
func NewImageMetadata() (ImageMetadata, error) {
	meta := ImageMetadata{}
	path, err := os.Getwd()
	if err != nil {
		return meta, errors.Wrap(err, failedToGetwdMsg)
	}
	gm := NewLocalGitManager(path)
	tags, annotations, err := gm.Tags()
	if err != nil {
		return meta, errors.Wrap(err, "Failed to get tags and annotations")
	}
	meta.GitTag = tags
	meta.GitAnnotation = annotations

	remote, err := gm.Remote()
	if err != nil {
		return meta, errors.Wrap(err, failedToGetRemoteMsg)
	}
	meta.GitOrigin = remote

	hash, err := gm.CommitHash()
	if err != nil {
		return meta, errors.Wrap(err, failedToGetCommitHashMsg)
	}
	meta.GitCommit = hash

	meta.CreatedAt = time.Now().Format(time.RFC3339)

	return meta, nil
}

// NewLocalGitManager returns a pointer to an initialized LocalGitManager and takes a `path`
func NewLocalGitManager(path ...string) *LocalGitManager {
	var thePath string
	if path != nil {
		thePath = path[0]
	}
	return &LocalGitManager{path: thePath}
}

// CommitHash returns the commit hash of a git repo at either the set path or current
// working directory
func (gm *LocalGitManager) CommitHash() (string, error) {
	if gm.path != "" {
		oldPath, err := os.Getwd()
		if err != nil {
			return "", errors.Wrap(err, failedToGetwdMsg)
		}
		err = os.Chdir(gm.path)
		if err != nil {
			return "", errors.Wrapf(err, failedToChangeDirectoryMsg, gm.path)
		}
		defer os.Chdir(oldPath)
	}
	hashCmd := exec.Command("git", "rev-parse", "--short", "HEAD")

	hash, err := hashCmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrap(err, "Failed to retrive git commit hash from git")
	}

	stripped := strings.TrimSpace(string(hash))
	return stripped, nil
}

// Remote returns the remote of a git repo at either the set path or current
// working directory
func (gm *LocalGitManager) Remote() (string, error) {
	if gm.path != "" {
		oldPath, err := os.Getwd()
		if err != nil {
			return "", errors.Wrap(err, failedToGetwdMsg)
		}
		err = os.Chdir(gm.path)
		if err != nil {
			return "", errors.Wrapf(err, failedToChangeDirectoryMsg, gm.path)
		}
		defer os.Chdir(oldPath)
	}
	remoteCmd := exec.Command("git", "remote", "-v")
	remote, err := remoteCmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrap(err, "Failed to retrive git remote value")
	}

	remotes := strings.Fields(string(remote))
	if len(remotes) == 0 {
		return "", nil
	}
	return remotes[1], nil
}

// RepoName returns the name of the repository extracted from the remote
// TODO: support multiple remotes (by ignoring all but origin)
// TODO: treat "origin" as a standard special case of remote that we will treat as the canonical name
func (gm *LocalGitManager) RepoName() (string, error) {
	origin, err := gm.Remote()

	if err != nil {
		return "", errors.Wrap(err, failedToGetRemoteMsg)
	}

	httpsRegex := regexp.MustCompile("^https://")
	gitRegex := regexp.MustCompile("^git@")

	if httpsRegex.MatchString(origin) {
		return processHTTPRepoName(origin), nil
	}

	if gitRegex.MatchString(origin) {
		return processGitRepoName(origin), nil
	}
	return "", errors.Errorf("Unknown git scheme for remote address %q", origin)
}

// Tags returns the tags and accompanying annotations of a git repo at either
// the set path or current working directory
func (gm *LocalGitManager) Tags() ([]string, []string, error) {
	if gm.path != "" {
		oldPath, err := os.Getwd()
		if err != nil {
			return nil, nil, errors.Wrap(err, failedToGetwdMsg)
		}
		err = os.Chdir(gm.path)
		if err != nil {
			return nil, nil, errors.Wrapf(err, failedToChangeDirectoryMsg, gm.path)
		}
		defer os.Chdir(oldPath)
	}
	currentTags := exec.Command("git", "tag", "-n1", "--points-at", "HEAD")

	rawTags, err := currentTags.CombinedOutput()
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to retrieve tags and annotations from git")
	}
	tags, annotations := processTags(string(rawTags))
	return tags, annotations, nil
}

// processTags returns two slices of strings, the first indicating the tags
// and the second indicating the annotations related to the current git
// commit. The incoming rawString is expected to be formatted as each tag/annotation
// pair being on its own line with the annotation beginning after the first
// whitespace on the line. For example, "v1.0.0     annotation here"
// (This is the equivalent to the output of `git tag -n1 --points-at HEAD`)
func processTags(rawString string) ([]string, []string) {
	var tags []string
	var annotations []string
	splitTagInfo := strings.Split(rawString, "\n")
	for _, v := range splitTagInfo {
		if v == "" {
			continue
		}
		r := regexp.MustCompile("[^\\s]+")
		split := r.FindAllString(v, -1)
		if len(split) > 1 {
			tag := split[0]
			annotation := strings.Join(split[1:], " ")
			tags = append(tags, tag)
			annotations = append(annotations, annotation)
		}
		if len(split) == 1 {
			tag := split[0]
			tags = append(tags, tag)
		}
	}
	return tags, annotations
}

// processHTTPRepoName returns the name of the repo if it uses a form like https://github.com/foobar/bargaz.git
func processHTTPRepoName(remote string) string {
	pieces := strings.Split(remote, "/")
	fullRepoName := pieces[len(pieces)-1]
	golfed := strings.Replace(fullRepoName, ".git", "", -1)
	return golfed
}

// processGitRepoName returns the name of the repo if it uses a form like git@github.com:foobar/bargaz.git
func processGitRepoName(remote string) string {
	fullRepoName := strings.Split(remote, "/")[1]
	golfed := strings.Replace(fullRepoName, ".git", "", -1)
	return golfed
}
