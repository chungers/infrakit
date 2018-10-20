package local // import "github.com/docker/infrakit/pkg/run/local"

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/docker/infrakit/pkg/types"
)

const (
	// EnvInfrakitHome is the environment variable for defining the top level working directory
	// for infrakit.
	EnvInfrakitHome = "INFRAKIT_HOME"

	// EnvInfrakitHost is the environment variable for setting the infrakit host to connect to
	EnvInfrakitHost = "INFRAKIT_HOST"

	// EnvPlaybooks is the environment variable for storing the playbooks file
	EnvPlaybooks = "INFRAKIT_PLAYBOOKS_FILE"

	// EnvClientTimeout is the timeout used by the rpc client
	EnvClientTimeout = "INFRAKIT_CLIENT_TIMEOUT"
)

func init() {

	EnsureDir(defaultInfrakitHomeDir())

}

// ClientTimeout returns the client timeout
func ClientTimeout() time.Duration {
	return types.MustParseDuration(Getenv(EnvClientTimeout, "15s")).Duration()
}

// defaultInfrakitHomeDir returns the default
func defaultInfrakitHomeDir() string {
	top := ""
	if usr, err := user.Current(); err == nil {
		top = usr.HomeDir
	} else if dir := os.Getenv("HOME"); dir != "" {
		top = dir
	} else if dir, err := os.Getwd(); err == nil {
		top = dir
	} else {
		top = os.TempDir()
	}
	return filepath.Join(top, ".infrakit")
}

// InfrakitHome returns the directory of INFRAKIT_HOME if specified. Otherwise, it will return
// the user's home directory.  If that cannot be determined, then it returns the current working
// directory.  If that still cannot be determined, a temporary directory is returned.
func InfrakitHome() string {
	return Getenv(EnvInfrakitHome, defaultInfrakitHomeDir())
}

// InfrakitHost returns the value of the INFRAKIT_HOST environment
func InfrakitHost() string {
	return Getenv(EnvInfrakitHost, "local")
}

// Playbooks returns the path to the playbooks
func Playbooks() string {
	return Getenv(EnvPlaybooks, filepath.Join(InfrakitHome(), "playbooks.yml"))
}

// Getenv returns the value at the environment variable 'env'.  If the value is not found
// then default value is returned
func Getenv(env string, defaultValue string) string {
	v := os.Getenv(env)
	if v != "" {
		return v
	}
	return defaultValue
}

// EnsureDir ensures the directory exists
func EnsureDir(dir string) error {
	stat, err := os.Stat(dir)
	if err == nil {
		if !stat.IsDir() {
			return fmt.Errorf("not a directory %v", dir)
		}
		return nil
	}
	if os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return fmt.Errorf("error access dir %s: %s", dir, err)
}
