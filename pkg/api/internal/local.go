package internal

import (
	"github.com/docker/infrakit/pkg/api"
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

// Connect connects to the infrakit running locally on the same host where the controller's
// unix sockets are accessible to the process making this call.
func Connect(opt api.Options) (api.Scope, error) {
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
	return &localScope{scope.DefaultScope(func() discovery.Plugins { return discover })}, nil
}

type localScope struct {
	scope.Scope
}

func (l *localScope) Profiles() (map[string]api.Profile, error) {
	return nil, nil
}

func (l *localScope) Provision(profile api.Profile) (types.Metadata, error) {
	return types.Metadata{}, nil
}

func (l *localScope) Terminate(obj types.Metadata) error {
	return nil
}
