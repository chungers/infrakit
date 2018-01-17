package api

import (
	"net/url"

	logutil "github.com/docker/infrakit/pkg/log"
	"github.com/docker/infrakit/pkg/template"
	"github.com/docker/infrakit/pkg/types"
)

var Logger = logutil.New("module", "api")

type Builder interface {
	Profile
	Properties
	AddTag(key, value string) Builder
	Set(key string, value interface{}) Builder
}

type Properties interface {
	Properties() []string
	Get(key string) interface{}
}

// Profiles are yml or json files that define parameters and the input of plugin that's
// responsible for managing it.  A profile yml is the same as a playbook command.
// The flags defined in the yml are automatically built as parameters that can be set via
// the Set function on the profile.
type Profile interface {
	Source() *url.URL
}

// Scope provides a restricted set of services for programmatic access of infrakit.
type Scope interface {

	// Profiles returns all known profiles. Profiles are loaded from the Options.ProfilePaths
	// and a profile can be a template for either a single instance or an entire cluster.
	Profiles() (map[string]Profile, error)

	// Enforce enforces the desired state
	Enforce(Profile) (types.Metadata, error)

	// Terminate terminates the object
	Terminate(types.Metadata) error
}

// Options contains configuration setting for the scope
type Options struct {

	// ProfilePath is the search paths for the profiles that are available for access
	ProfilePaths []string

	TemplateOptions template.Options
}
