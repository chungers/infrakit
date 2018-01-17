package compute

import (
	"fmt"

	"github.com/docker/infrakit/pkg/api"
	"github.com/docker/infrakit/pkg/api/internal"
	"github.com/docker/infrakit/pkg/run/scope"
	"github.com/docker/infrakit/pkg/types"
)

type ShellScript string

func (s ShellScript) Source() string {
	return string(s)
}

type Machine struct {
	MachineBuilder

	InstanceType string
}

type MachineBuilder interface {
	api.Builder
	internal.Committer
	SetLogicalID(string) MachineBuilder
	AddInit(ShellScript) MachineBuilder
}

type Profile struct {
	internal.Profile
}

func (p Profile) Commit(scope scope.Scope) (types.Metadata, error) {
	// use data here
	return types.Metadata{}, fmt.Errorf("machine commit: not implemented yet")
}

func (p *Profile) AddInit(init ShellScript) MachineBuilder {
	return p
}

func (p *Profile) SetLogicalID(id string) MachineBuilder {
	return p
}

// Customize -- down casts a profile to a machine profile
func Customize(base api.Profile) (MachineBuilder, error) {
	// call Profile.Type() to check here...
	return &Profile{internal.Profile{Profile: base}}, nil
}
