package script

import (
	"github.com/docker/infrakit/pkg/controller/internal"
	"github.com/docker/infrakit/pkg/run/scope"
	"github.com/docker/infrakit/pkg/types"
)

// Parser for the targets section.  This parses a string list
func targetsFromStringList(scope scope.Scope, any *types.Any) ([]string, error) {
	list := []string{}
	err := types.Decode(any.Bytes(), &list)
	return list, err
}

// Parser for the targets section.  This parses the spec for loading instances
// from an instance plugin based on select on the tags.
func targetsFromInstanceMatchingSelectTags(scope scope.Scope, any *types.Any) ([]string, error) {

	// Use internal.InstanceObserver as the schema.  It contains select
	// criteria, etc.

	observer := internal.InstanceObserver{}
	if err := any.Decode(&observer); err != nil {
		return nil, err
	}

	if err := observer.Validate(nil); err != nil {
		return nil, err
	}

	// TODO - implement stream/event based batching of calls as new instances
	// become available...

	if err := observer.Init(scope, 0); err != nil {
		return nil, err
	}

	instances, err := observer.Observe()
	if err != nil {
		return nil, err
	}

	targets := []string{}

	for _, instance := range instances {
		if key, err := observer.KeyOf(instance); err != nil {
			targets = append(targets, string(instance.ID))
		} else {
			targets = append(targets, key)
		}
	}
	return targets, nil
}
