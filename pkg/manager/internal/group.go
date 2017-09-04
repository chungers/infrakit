package internal

import (
	"fmt"

	group_types "github.com/docker/infrakit/pkg/plugin/group/types"
	"github.com/docker/infrakit/pkg/spi/group"
	"github.com/docker/infrakit/pkg/spi/instance"
	"github.com/docker/infrakit/pkg/types"
)

// This file contains the manager implementation / override of the group.Plugin interface.
// The manager exposes itself as a Group managing the state for the backend stateless group plugin.
// TODO(chunger) - allow the manager to represent multiple group plugin backends.
//
// Note - the group spec operations here deal with a group within a single Plugin data model.
// Support for multiple groups need to account for *naming* of multiple connected Group plugins.
//
// Support for multiple group plugins isn't a high-priority, because it's still possible to have
// a single group abstraction for managing multiple groups of different instance/flavor pairs
// (e.g. us-east-1a/workers vs. us-west-2a/workers)

// QueuedGroupPlugin returns a controller.Controller that has a backing storage for specs and
// where all operations are serialized onto a work queue.
func QueuedGroupPlugin(g group.Plugin, ch chan<- backendOp,
	allGroupSpecsFunc func() ([]group.Spec, error),
	findGroupSpecFunc func(group.ID) (group.Spec, error),
	updateGroupSpecFunc func(group.Spec) error,
	removeGroupSpecFunc func(group.ID) error) group.Plugin {

	return &queuedGroupPlugin{
		Queued:              queue(ch),
		Plugin:              g,
		allGroupSpecsFunc:   allGroupSpecsFunc,
		findGroupSpecFunc:   findGroupSpecFunc,
		updateGroupSpecFunc: updateGroupSpecFunc,
		removeGroupSpecFunc: removeGroupSpecFunc,
	}
}

type queuedGroupPlugin struct {
	Queued

	group.Plugin // the backend that does the real work

	allGroupSpecsFunc   func() ([]group.Spec, error)
	findGroupSpecFunc   func(group.ID) (group.Spec, error)
	updateGroupSpecFunc func(group.Spec) error
	removeGroupSpecFunc func(group.ID) error
}

// This implements/ overrides the Group Plugin interface to support single group-only operations
func (q *queuedGroupPlugin) CommitGroup(grp group.Spec, pretend bool) (resp string, err error) {
	result := q.Run(group.Plugin.CommitGroup,
		func() []interface{} {

			// We first update the user's desired state first
			if !pretend {
				if updateErr := q.updateGroupSpecFunc(grp); updateErr != nil {
					log.Warn("Error updating", "err", updateErr)
					return []interface{}{"cannot update", updateErr}
				}
			}

			r1, r2 := q.Plugin.CommitGroup(grp, pretend)
			return []interface{}{r1, r2}
		})

	if v, has := result[0].(string); has {
		resp = v
	}
	if v, has := result[1].(error); has && v != nil {
		err = v
	}
	return
}

// InspectGroups returns all the desired specs.  This is intercepted with the stored version
func (q *queuedGroupPlugin) InspectGroups() (specs []group.Spec, err error) {
	return q.allGroupSpecsFunc()
}

// Serialized describe group
func (q *queuedGroupPlugin) DescribeGroup(id group.ID) (desc group.Description, err error) {
	result := q.Run(group.Plugin.DescribeGroup,
		func() []interface{} {
			r1, r2 := q.Plugin.DescribeGroup(id)
			return []interface{}{r1, r2}
		})

	if v, is := result[0].(group.Description); is {
		desc = v
	}
	if v, is := result[1].(error); is {
		err = v
	}
	return
}

// This implements/ overrides the Group Plugin interface to support single group-only operations
func (q *queuedGroupPlugin) DestroyGroup(id group.ID) (err error) {
	result := q.Run(group.Plugin.DestroyGroup,
		func() []interface{} {

			// We first update the user's desired state first

			// At least make sure we know about this record
			// If we'd let this proceed is a matter of policy
			if _, err := q.findGroupSpecFunc(id); err == nil {
				if removeErr := q.removeGroupSpecFunc(id); removeErr != nil {
					log.Warn("Error updating/ remove spec. Continue.", "err", removeErr)
				}
			}

			return []interface{}{q.Plugin.DestroyGroup(id)}
		})

	if v, is := result[0].(error); is {
		err = v
	}
	return
}

// This implements/ overrides the Group Plugin interface to support single group-only operations
func (q *queuedGroupPlugin) FreeGroup(id group.ID) (err error) {
	result := q.Run(group.Plugin.FreeGroup,
		func() []interface{} {

			// We first update the user's desired state first

			// At least make sure we know about this record
			// If we'd let this proceed is a matter of policy
			if _, err := q.findGroupSpecFunc(id); err == nil {
				if removeErr := q.removeGroupSpecFunc(id); removeErr != nil {
					log.Warn("Error updating/ remove spec. Continue.", "err", removeErr)
				}
			}

			return []interface{}{q.Plugin.FreeGroup(id)}
		})
	if v, is := result[0].(error); is {
		err = v
	}
	return
}

// This implements/ overrides the Group Plugin interface to support single group-only operations
func (q *queuedGroupPlugin) DestroyInstances(id group.ID, instances []instance.ID) (err error) {
	result := q.Run(group.Plugin.DestroyInstances,
		func() []interface{} {
			return []interface{}{q.Plugin.DestroyInstances(id, instances)}
		})
	if v, is := result[0].(error); is {
		err = v
	}
	return
}

// This implements/ overrides the Group Plugin interface to support single group-only operations
func (q *queuedGroupPlugin) SetSize(id group.ID, size int) error {
	spec, err := q.findGroupSpecFunc(id)
	if err != nil {
		return err
	}
	parsed, err := group_types.ParseProperties(spec)
	if err != nil {
		return err
	}
	if s := len(parsed.Allocation.LogicalIDs); s > 0 {
		return fmt.Errorf("cannot set size when logical ids are explicitly set")
	}
	parsed.Allocation.Size = uint(size)
	spec.Properties = types.AnyValueMust(parsed)
	_, err = q.CommitGroup(spec, false)
	return err
}

// This implements/ overrides the Group Plugin interface to support single group-only operations
func (q *queuedGroupPlugin) Size(id group.ID) (size int, err error) {
	spec, err := q.findGroupSpecFunc(id)
	if err != nil {
		return 0, err
	}
	parsed, err := group_types.ParseProperties(spec)
	if err != nil {
		return 0, err
	}
	if s := len(parsed.Allocation.LogicalIDs); s > 0 {
		size = s
		return size, nil
	}
	return int(parsed.Allocation.Size), nil
}
