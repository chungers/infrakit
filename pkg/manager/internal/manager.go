package internal

import (
	"fmt"
	"net/url"

	"github.com/docker/infrakit/pkg/controller"
	"github.com/docker/infrakit/pkg/core"
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

// Supervising returns information about the objects supervised by this backend.
func (b *Backend) Supervising() ([]plugin.Metadata, error) {
	out := []plugin.Metadata{}

	err := b.visitPlugins(
		asController(
			func(n plugin.Name, c controller.Controller) error {

				specs, err := c.Specs(nil)
				if err != nil {
					return err
				}

				kind := ""
				for _, s := range specs {
					addr := core.AsAddressable(s)
					kind = addr.Kind()
					out = append(out, plugin.Metadata{
						Kind:          kind,
						Name:          addr.Plugin(),
						Instance:      addr.Instance(),
						InterfaceSpec: controller.InterfaceSpec,
					})
				}

				addr := core.NewAddressable(kind, n.LookupOnly(), "")
				out = append(out, plugin.Metadata{
					Kind:          addr.Kind(),
					Name:          addr.Plugin(),
					Instance:      addr.Instance(),
					InterfaceSpec: controller.InterfaceSpec,
				})

				return nil
			},
		),
		asGroupPlugin(
			func(n plugin.Name, g group.Plugin) error {

				gspecs, err := g.InspectGroups()
				if err != nil {
					return err
				}

				kind := ""
				for _, gs := range gspecs {
					addr := core.NewAddressable(kind, n.LookupOnly(), string(gs.ID))
					kind = addr.Kind()
					out = append(out, plugin.Metadata{
						Kind:          kind,
						Name:          addr.Plugin(),
						Instance:      addr.Instance(),
						InterfaceSpec: group.InterfaceSpec,
					})
				}

				addr := core.NewAddressable(kind, n.LookupOnly(), "")
				out = append(out, plugin.Metadata{
					Kind:          addr.Kind(),
					Name:          addr.Plugin(),
					Instance:      addr.Instance(),
					InterfaceSpec: group.InterfaceSpec,
				})

				return nil
			},
		))
	return out, err
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
			handler = plugin.Name(s.Metadata.Name)
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
	err := b.visitPlugins(asController(
		func(n plugin.Name, c controller.Controller) error {
			objects, err := c.Describe(nil)
			if err != nil {
				return err
			}
			aggregated = append(aggregated, objects...)
			return nil
		},
	))
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
