package api

import (
	api_types "github.com/docker/infrakit/pkg/api/types"
	logutil "github.com/docker/infrakit/pkg/log"
	"github.com/docker/infrakit/pkg/types"
)

type Profile interface {
	api_types.Committer
	api_types.InstanceSpecBuilder
	api_types.Properties
}

// Scope provides a restricted set of services for programmatic access of infrakit.
type Scope interface {

	// Profiles returns all known profiles for creating resource singly or as a whole
	Profiles() (map[string]Profile, error)

	// Provision an instance based on the profile
	Provision(Profile) (types.Metadata, error)

	// Terminate terminates the object
	Terminate(types.Metadata) error
}

var (
	Logger = logutil.New("module", "api")
)

type Options struct {
}

func Connect(connect string, opt Options) (Scope, error) {
	return nil, nil
}
