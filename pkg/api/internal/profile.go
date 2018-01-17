package internal

import (
	"fmt"
	"net/url"
	"path"

	"github.com/docker/infrakit/pkg/api"
	"github.com/docker/infrakit/pkg/plugin"
	"github.com/docker/infrakit/pkg/run/scope"
	"github.com/docker/infrakit/pkg/run/scope/local"
	"github.com/docker/infrakit/pkg/template"
	"github.com/docker/infrakit/pkg/types"
)

type Committer interface {
	Commit(scope scope.Scope) (types.Metadata, error)
}

type NeedsPlugins interface {
	RequiredPlugins() ([]local.StartPlugin, error)
}

func Profiles(paths []string, opts template.Options) (map[string]api.Profile, error) {
	found := map[string]api.Profile{}

	defer log.Debug("profiles", "found", found)

	for _, index := range paths {
		t, err := template.NewTemplate(index, opts)
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

		for key, profile := range profiles {
			path := path.Join(path.Dir(index), profile)

			source, err := url.Parse(path)
			if err != nil {
				return found, err
			}

			// here we load the profile asset and create the Profile object
			profileTemplate, err := template.NewTemplate(path, opts)
			if err != nil {
				return found, err
			}

			kind := "aws"                  // TODO - parse from the template
			backend := "instanceProvision" // TODO - parse from template
			found[key] = Profile{
				source:   source,
				kind:     kind,
				backend:  backend,
				template: profileTemplate,
			}
		}
	}
	return found, nil
}

type Profile struct {
	api.Profile
	kind     string
	backend  string
	source   *url.URL
	template *template.Template
}

func (p Profile) Commit(scope scope.Scope) (types.Metadata, error) {
	// use data here
	return types.Metadata{}, fmt.Errorf("profile commit: not implemented yet")
}

func (p Profile) RequiredPlugins() ([]local.StartPlugin, error) {
	return []local.StartPlugin{}, nil
}

func (p Profile) Source() *url.URL {
	return p.source
}

func (p *Profile) AddTag(key, value string) api.Builder {
	return p
}

func (p *Profile) Set(key string, value interface{}) api.Builder {
	return p
}

func (p *Profile) Get(key string) interface{} {
	return nil
}

func (p *Profile) Properties() []string {
	return nil
}
