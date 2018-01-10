package internal

import (
	"fmt"
	"path"

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

func init() {
	api.Register("local://", Connect)
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

	found := map[string]api.Profile{}

	for _, index := range l.options.ProfilePaths {

		t, err := template.NewTemplate(index, template.Options{})
		if err != nil {
			log.Error("cannot parse profile index", "err", err)
			continue
		}

		buff, err := t.Render(nil)
		if err != nil {
			log.Error("cannot process profile index", "err", err)
			continue
		}

		profiles := map[string]string{}
		if err := types.Decode([]byte(buff), &profiles); err != nil {
			log.Error("cannot load profiles", "err", err)
			continue
		}

		for _, profile := range profiles {
			path := path.Join(path.Dir(index), profile)

			fmt.Println(">>>>>", path)
			// here we load the profile asset and create the Profile object
		}

	}

	return found, nil
}

func (l *localScope) Provision(profile api.Profile) (types.Metadata, error) {
	return profile.Commit(l)
}

func (l *localScope) Terminate(obj types.Metadata) error {
	return nil
}
