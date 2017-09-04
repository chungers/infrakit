package controller

import (
	"github.com/docker/infrakit/pkg/controller"
	"github.com/docker/infrakit/pkg/types"
)

// Controller implements the controller.Controller interface and supports
// testing by letting user assemble behavior dyanmically.
type Controller struct {
	// Plan is a commit without actually making the changes.  The controller returns a proposed object state
	// after commit, with a Plan, or error.
	DoPlan func(operation controller.Operation, spec types.Spec) (types.Object, controller.Plan, error)

	// Commit commits the spec to the controller for management.  The controller's job is to ensure reality
	// matches the specification.  The spec can be composed and references other controllers or plugins.
	// When a spec is committed to a controller, the controller returns the object state corresponding to
	// the spec.  When operation is Destroy, only Metadata portion of the spec is needed to identify
	// the object to be destroyed.
	DoCommit func(operation controller.Operation, spec types.Spec) (types.Object, error)

	// Describe returns a list of objects matching the metadata provided. A list of objects are possible because
	// metadata can be a tags search.  An object has state, and its original spec can be accessed as well.
	// A nil Metadata will instruct the controller to return all objects under management.
	DoDescribe func(metadata *types.Metadata) ([]types.Object, error)

	// DoSpecs returns the objective specifications.  It is in contrast with the output of Describe where is current state.
	// The current state may or may not match the user's specification, which this returns.
	// Note that a list is returned.  This is because Commit can be invoked multiple times with different specs, resulting
	// in a set of objectives each corresponding to the object in Describe.  This does not assume any memory on the part
	// of the implementation.  The specs can be non-durable and the user would have to provide it each time on restart
	// or via a datastore.
	DoSpecs func(metadata *types.Metadata) ([]types.Spec, error)

	// DoPause tells the controller to pause management of objects matching.  To resume, commit again.
	DoPause func(metadata *types.Metadata) ([]types.Object, error)

	// DoTerminate tells the controller to terminate / destroy the objects matching search.
	DoTerminate func(metadata *types.Metadata) ([]types.Object, error)
}

// Plan implements pkg/controller/Controller.Plan
func (t *Controller) Plan(operation controller.Operation, spec types.Spec) (types.Object, controller.Plan, error) {
	return t.DoPlan(operation, spec)
}

// Commit implements pkg/controller/Controller.Commit
func (t *Controller) Commit(operation controller.Operation, spec types.Spec) (types.Object, error) {
	return t.DoCommit(operation, spec)
}

// Describe implements pkg/controller/Controller.Describe
func (t *Controller) Describe(metadata *types.Metadata) ([]types.Object, error) {
	return t.DoDescribe(metadata)
}

// Specs implements pkg/controller/Controller.Specs
func (t *Controller) Specs(metadata *types.Metadata) ([]types.Spec, error) {
	return t.DoSpecs(metadata)
}

// Pause implements pkg/controller/Controller.Pause
func (t *Controller) Pause(metadata *types.Metadata) ([]types.Object, error) {
	return t.DoPause(metadata)
}

// Terminate implements pkg/controller/Controller.Terminate
func (t *Controller) Terminate(metadata *types.Metadata) ([]types.Object, error) {
	return t.DoTerminate(metadata)
}
