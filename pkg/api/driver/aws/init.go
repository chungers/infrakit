package aws

import (
	api "github.com/docker/infrakit/pkg/api/types"
	"github.com/docker/infrakit/pkg/run/scope"
	"github.com/docker/infrakit/pkg/types"

	_ "github.com/docker/infrakit/pkg/run/v0/aws"
)

func init() {

}

type profile struct {
}

func (p *profile) Commit(scope scope.Scope) (types.Metadata, error) {
	return types.Metadata{}, nil
}

func (p *profile) AddTag(key, value string) *profile {
	return p
}

func (p *profile) AddInit(init api.InstanceInit) *profile {
	return p
}

func (p *profile) SetLogicalID(id string) *profile {
	return p
}

func (p *profile) Set(key string, value interface{}) *profile {
	return p
}

func (p *profile) Get(key string) interface{} {
	return nil
}
