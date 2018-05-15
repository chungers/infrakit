package types

import (
	"github.com/docker/infrakit/pkg/run/depends"
	"github.com/docker/infrakit/pkg/types"
)

// ResolveDependencies returns a list of dependencies by parsing the opaque Properties blob.
func ResolveDependencies(spec types.Spec) (depends.Runnables, error) {
	if spec.Properties == nil {
		return nil, nil
	}

	properties := Properties{}
	err := spec.Properties.Decode(&properties)
	if err != nil {
		return nil, err
	}

	return depends.Runnables{}, nil
}
