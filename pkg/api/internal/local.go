package internal

import (
	"fmt"

	"github.com/docker/infrakit/pkg/api"
	"github.com/docker/infrakit/pkg/api/internal/registry"
	"github.com/docker/infrakit/pkg/discovery"
	"github.com/docker/infrakit/pkg/discovery/local"
	logutil "github.com/docker/infrakit/pkg/log"
	"github.com/docker/infrakit/pkg/run/scope"
	"github.com/docker/infrakit/pkg/template"
	"github.com/docker/infrakit/pkg/types"
)

var (
	log    = logutil.New("module", "api/internal/local")
	debugV = logutil.V(300)
)

func init() {
	registry.Register("local://", Connect)
}

// Connect connects to the infrakit running locally on the same host where the controller's
// unix sockets are accessible to the process making this call.
func Connect(connect string, opt api.Options) (api.Scope, error) {
	if err := local.Setup(); err != nil {
		return nil, err
	}
	if err := template.Setup(); err != nil {
		return nil, err
	}
	discover, err := local.NewPluginDiscovery()
	if err != nil {
		return nil, err
	}
	return &localScope{
		Scope:   scope.DefaultScope(func() discovery.Plugins { return discover }),
		options: opt,
	}, nil
}

type localScope struct {
	scope.Scope
	options api.Options
}

func (l *localScope) Profiles() (map[string]api.Profile, error) {
	return Profiles(l.options.ProfilePaths, l.options.TemplateOptions)
}

func (l *localScope) Enforce(profile api.Profile) (types.Metadata, error) {
	result := types.Metadata{}

	fmt.Println("profile=", profile)

	required, is := profile.(NeedsPlugins)
	if is {
		plugins, err := required.RequiredPlugins()
		if err != nil {
			return result, err
		}
		for _, plugin := range plugins {
			// start each plugin if not running
		}
	}

	internal, is := profile.(Committer)
	if is {
		return internal.Commit(l)
	}
	return result, fmt.Errorf("not implementation")
}

func (l *localScope) Terminate(obj types.Metadata) error {
	return nil
}
