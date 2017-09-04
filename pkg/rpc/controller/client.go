package controller

import (
	"github.com/docker/infrakit/pkg/controller"
	"github.com/docker/infrakit/pkg/plugin"
	rpc_client "github.com/docker/infrakit/pkg/rpc/client"
	"github.com/docker/infrakit/pkg/types"
)

// NewClient returns a plugin interface implementation connected to a remote plugin
func NewClient(name plugin.Name, socketPath string) (controller.Controller, error) {
	rpcClient, err := rpc_client.New(socketPath, controller.InterfaceSpec)
	if err != nil {
		return nil, err
	}
	return Adapt(name, rpcClient), nil
}

// Adapt converts a rpc client to a Controller object
func Adapt(name plugin.Name, rpcClient rpc_client.Client) controller.Controller {
	return &client{name: name, client: rpcClient}
}

type client struct {
	name   plugin.Name
	client rpc_client.Client
}

// Plan is a commit without actually making the changes.  The controller returns a proposed object state
// after commit, with a Plan, or error.
func (c client) Plan(operation controller.Operation, spec types.Spec) (types.Object, controller.Plan, error) {
	req := ChangeRequest{
		Name:      c.name,
		Operation: operation,
		Spec:      spec,
	}
	resp := ChangeResponse{}
	err := c.client.Call("Controller.Plan", req, &resp)
	return resp.Object, resp.Plan, err
}

// Commit commits the spec to the controller for management.  The controller's job is to ensure reality
// matches the specification.  The spec can be composed and references other controllers or plugins.
// When a spec is committed to a controller, the controller returns the object state corresponding to
// the spec.  When operation is Destroy, only Metadata portion of the spec is needed to identify
// the object to be destroyed.
func (c client) Commit(operation controller.Operation, spec types.Spec) (types.Object, error) {
	req := ChangeRequest{
		Name:      c.name,
		Operation: operation,
		Spec:      spec,
	}
	resp := ChangeResponse{}
	err := c.client.Call("Controller.Commit", req, &resp)
	return resp.Object, err
}

// Describe returns a list of objects matching the metadata provided. A list of objects are possible because
// metadata can be a tags search.  An object has state, and its original spec can be accessed as well.
// A nil Metadata will instruct the controller to return all objects under management.
func (c client) Describe(metadata *types.Metadata) ([]types.Object, error) {
	req := FindRequest{
		Name:     c.name,
		Metadata: metadata,
	}
	resp := FindResponse{}
	err := c.client.Call("Controller.Describe", req, &resp)
	return resp.Objects, err
}

// Specs returns the objective specifications.  It is in contrast with the output of Describe where is current state.
// The current state may or may not match the user's specification, which this returns.
// Note that a list is returned.  This is because Commit can be invoked multiple times with different specs, resulting
// in a set of objectives each corresponding to the object in Describe.  This does not assume any memory on the part
// of the implementation.  The specs can be non-durable and the user would have to provide it each time on restart
// or via a datastore.
func (c client) Specs(metadata *types.Metadata) ([]types.Spec, error) {
	req := FindRequest{
		Name:     c.name,
		Metadata: metadata,
	}
	resp := FindResponse{}
	err := c.client.Call("Controller.Specs", req, &resp)
	return resp.Specs, err
}

// Pause tells the controller to pause management of objects matching.  To resume, commit again.
func (c client) Pause(metadata *types.Metadata) ([]types.Object, error) {
	req := FindRequest{
		Name:     c.name,
		Metadata: metadata,
	}
	resp := FindResponse{}
	err := c.client.Call("Controller.Pause", req, &resp)
	return resp.Objects, err
}

// Terminate tells the controller to terminate management of objects matching.
func (c client) Terminate(metadata *types.Metadata) ([]types.Object, error) {
	req := FindRequest{
		Name:     c.name,
		Metadata: metadata,
	}
	resp := FindResponse{}
	err := c.client.Call("Controller.Terminate", req, &resp)
	return resp.Objects, err
}
