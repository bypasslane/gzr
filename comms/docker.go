package comms

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// ImageManager is an interface for building an application's image
type ImageManager interface {
	Build(...string) error
	Push(string) error
}

// DockerManager implements ImageManager in order to manage images for Docker
type DockerManager struct{}

// NewDockerManager returns a pointer to an initialized DockerManager
func NewDockerManager() *DockerManager {
	return &DockerManager{}
}

// Build takes a series of arguments to be sent to Docker and builds an image
func (docker *DockerManager) Build(args ...string) error {
	tag, err := GetDockerTag()
	if err != nil {
		return err
	}
	args = append([]string{"build", "-t", tag}, args...)
	build := exec.Command("docker", args...)
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	err = build.Run()
	if err != nil {
		return err
	}
	return nil
}

// Push takes a name and pushes this to Docker
func (docker *DockerManager) Push(name string) error {
	push := exec.Command("docker", "push", name)
	push.Stdout = os.Stdout
	push.Stderr = os.Stderr

	err := push.Run()
	if err != nil {
		return err
	}
	return nil
}

// GetDockerTag combines a configured Docker repository name, the current working directory,
// the current time, and a git hash to create a Docker tag appropriate to gzr
// Output format: `repository/$CWD:YYYYMMDD.SHORT_HASH`
func GetDockerTag() (string, error) {
	// The cwd call isn't needed. The commit hash function already uses the current working directory
	// and less work is done if the path isn't set
	gm := NewLocalGitManager()
	hash, err := gm.CommitHash()
	if err != nil {
		return "", errors.Wrap(err, "Failed to get git commit hash")
	}

	name, err := gm.RepoName()

	if err != nil {
		return "", errors.Wrap(err, "Failed to get Repo Name from git")
	}

	return fmt.Sprintf("%s/%s:%s.%s", viper.GetString("repository"), name, time.Now().Format("20060102"), hash), nil
}
