package types

import (
	"github.com/docker/infrakit/pkg/run/scope"
	"github.com/docker/infrakit/pkg/types"
)

type ShellScript string

func (s ShellScript) Source() string {
	return string(s)
}

type InstanceInit interface {
	Source() string
}

type Properties interface {
	Get(key string) interface{}
}

type InstanceSpecBuilder interface {
	AddTag(key, value string) InstanceSpecBuilder
	AddInit(InstanceInit) InstanceSpecBuilder
	SetLogicalID(string) InstanceSpecBuilder
	Set(key string, value interface{}) InstanceSpecBuilder
}

type Committer interface {
	Commit(scope.Scope) (types.Metadata, error)
}
