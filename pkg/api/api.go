package api

import (
	"fmt"
	"strings"
	"sync"

	api_types "github.com/docker/infrakit/pkg/api/types"
	logutil "github.com/docker/infrakit/pkg/log"
	"github.com/docker/infrakit/pkg/types"
)

// Profiles are yml or json files that define parameters and the input of plugin that's
// responsible for managing it.  A profile yml is the same as a playbook command.
// The flags defined in the yml are automatically built as parameters that can be set via
// the Set function on the profile.
type Profile interface {
	api_types.Committer
	api_types.InstanceSpecBuilder
	api_types.Properties
}

// Scope provides a restricted set of services for programmatic access of infrakit.
type Scope interface {

	// Profiles returns all known profiles. Profiles are loaded from the Options.ProfilePaths
	// and a profile can be a template for either a single instance or an entire cluster.
	Profiles() (map[string]Profile, error)

	// Provision an instance based on the profile
	Provision(Profile) (types.Metadata, error)

	// Terminate terminates the object
	Terminate(types.Metadata) error
}

var (
	Logger = logutil.New("module", "api")

	connectionFactoriesLock = sync.RWMutex{}
	connectionFactories     = map[string]func(string, Options) (Scope, error){}
)

// Register registers a key / connection type with the constructor function for creating
// the scope.  This is used by different kinds of implementations such as local (unix sockets)
// and remotes. It's typically called from implementations init() functions.
func Register(connTypePrefix string, constructor func(string, Options) (Scope, error)) {
	connectionFactoriesLock.Lock()
	defer connectionFactoriesLock.Unlock()

	connectionFactories[connTypePrefix] = constructor
}

// Options contains configuration setting for the scope
type Options struct {

	// ProfilePath is the search paths for the profiles that are available for access
	ProfilePaths []string
}

// Connect
func Connect(connect string, opt Options) (Scope, error) {
	connectionFactoriesLock.RLock()
	defer connectionFactoriesLock.RUnlock()

	for prefix, constructor := range connectionFactories {

		// e.g. local:// or tcp://
		if strings.Index(connect, prefix) == 0 {
			return constructor(connect, opt)
		}
	}

	return nil, fmt.Errorf("not found: %v", connect)
}
