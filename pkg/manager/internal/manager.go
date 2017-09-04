package internal

import (
	"fmt"
	"net/url"

	"github.com/docker/infrakit/pkg/controller"
	"github.com/docker/infrakit/pkg/plugin"
	"github.com/docker/infrakit/pkg/spi/group"
	"github.com/docker/infrakit/pkg/types"
)

// This file contains the backends implementation of the pkg/manager/Manager interface

// IsLeader returns leader status.  False if not or unknown.
func (b *Backend) IsLeader() (bool, error) {
	return b.isLeader, nil
}

// LeaderLocation returns the location of the leader
func (b *Backend) LeaderLocation() (*url.URL, error) {
	if b.leaderStore == nil {
		return nil, fmt.Errorf("cannot locate leader")
	}

	return b.leaderStore.GetLocation()
}

// Plan returns the changes needed given the new input
func (b *Backend) Plan(specs []types.Spec) (types.Changes, error) {
	current, err := b.Specs()
	if err != nil {
		return types.Changes{}, err
	}

	currentSpecs := types.Specs(current)
	updatedSpecs := types.Specs(specs)
	return currentSpecs.Changes(updatedSpecs), nil
}

// Enforce enforces infrastructure state to match that of the specs
func (b *Backend) Enforce(specs []types.Spec) error {

	// TODO
	requested := globalSpec{}
	for _, s := range specs {
		handler := plugin.NameFrom(s.Kind, s.Metadata.Name)
		if s.Kind == "group" {
			// TODO(chungers) -- this really needs to be cleaned up
			handler = plugin.Name(b.backendName)
			gspec := group.Spec{
				ID:         group.ID(s.Metadata.Name),
				Properties: s.Properties,
			}
			requested.updateGroupSpec(gspec, handler)
		} else {
			requested.updateSpec(s, handler)
		}
	}

	// Note we also have a version that's in the persistent store.
	// Should we do some delta calculations?

	return requested.store(b.snapshot)
}

// Specs returns the specs this manager is tasked with enforcing
func (b *Backend) Specs() ([]types.Spec, error) {
	global := globalSpec{}
	err := global.load(b.snapshot)
	if err != nil {
		return nil, err
	}
	saved := []types.Spec{}
	err = global.visit(func(k key, r record) error {
		saved = append(saved, r.Spec)
		return nil
	})
	return saved, err
}

// Inspect returns the current state of the infrastructure.  It performs an 'all-shard' query across
// all plugins of the type 'group' and then aggregate the results.
func (b *Backend) Inspect() ([]types.Object, error) {
	aggregated := []types.Object{}
	err := b.allControllers(func(c controller.Controller) error {
		objects, err := c.Describe(nil)
		if err != nil {
			return err
		}
		aggregated = append(aggregated, objects...)
		return nil
	})
	return aggregated, err
}

// Pause puts all the controllers referenced in the specs on pause.
// If specs is empty, locally stored version will be used.
func (b *Backend) Pause(specs []types.Spec) error {
	return fmt.Errorf("not implemented")
}

// Terminate destroys all resources associated with the specs
// If specs is empty, locally stored version will be used.
func (b *Backend) Terminate(specs []types.Spec) error {
	return fmt.Errorf("not implemented")
}
