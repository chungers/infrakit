package manager

import (
	"net/url"

	"github.com/docker/infrakit/pkg/plugin"
	"github.com/docker/infrakit/pkg/types"
)

// Plugin implements manager.Manager
type Plugin struct {

	// DoSupervising returns the list of controllers under supervision
	DoSupervising func() ([]plugin.Metadata, error)

	// DoLeaderLocation returns the location
	DoLeaderLocation func() (*url.URL, error)

	// DoIsLeader returns true if manager is leader
	DoIsLeader func() (bool, error)

	// DoPlan explains the changes
	DoPlan func(specs []types.Spec) (types.Changes, error)

	// DoEnforce enforces infrastructure state to match that of the specs
	DoEnforce func(specs []types.Spec) error

	// DoSpecs returns the list of specs the manager is tasked with enforcing
	DoSpecs func() ([]types.Spec, error)

	// DoInspect returns the current state of the infrastructure
	DoInspect func() ([]types.Object, error)

	// DoPause pauses all controllers in the specs
	DoPause func(specs []types.Spec) error

	// DoTerminate destroys all resources associated with the specs
	DoTerminate func(specs []types.Spec) error
}

// IsLeader returns true if manager is leader
func (t *Plugin) IsLeader() (bool, error) {
	return t.DoIsLeader()
}

// Supervising returns list of controllers under supervision
func (t *Plugin) Supervising() ([]plugin.Metadata, error) {
	return t.DoSupervising()
}

// LeaderLocation returns the location of the leader
func (t *Plugin) LeaderLocation() (*url.URL, error) {
	return t.DoLeaderLocation()
}

// Plan explains the changes
func (t *Plugin) Plan(specs []types.Spec) (types.Changes, error) {
	return t.DoPlan(specs)
}

// Enforce enforces infrastructure state to match that of the specs
func (t *Plugin) Enforce(specs []types.Spec) error {
	return t.DoEnforce(specs)
}

// Specs returns the specs this manager is tasked with enforcing
func (t *Plugin) Specs() ([]types.Spec, error) {
	return t.DoSpecs()
}

// Inspect returns the current state of the infrastructure
func (t *Plugin) Inspect() ([]types.Object, error) {
	return t.DoInspect()
}

// Pause pauses all controllers in the specs
func (t *Plugin) Pause(specs []types.Spec) error {
	return t.DoPause(specs)
}

// Terminate destroys all resources associated with the specs
func (t *Plugin) Terminate(specs []types.Spec) error {
	return t.DoTerminate(specs)
}
