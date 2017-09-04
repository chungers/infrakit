package manager

import (
	"net/http"
	"net/url"

	"github.com/docker/infrakit/pkg/manager"
	"github.com/docker/infrakit/pkg/spi"
	"github.com/docker/infrakit/pkg/types"
)

// PluginServer returns a Manager that conforms to the net/rpc rpc call convention.
func PluginServer(p manager.Manager) *Manager {
	return &Manager{manager: p}
}

// Manager is the exported type for json-rpc
type Manager struct {
	manager manager.Manager
}

// ImplementedInterface returns the interface implemented by this RPC service.
func (p *Manager) ImplementedInterface() spi.InterfaceSpec {
	return manager.InterfaceSpec
}

// Types returns the types exposed by this kind of RPC service
func (p *Manager) Types() []string {
	return []string{"."} // no types
}

// IsLeaderRequest is the rpc request
type IsLeaderRequest struct {
}

// IsLeaderResponse is the rpc response
type IsLeaderResponse struct {
	Leader bool
}

// IsLeader returns information about leadership status for this manager.
func (p *Manager) IsLeader(_ *http.Request, req *IsLeaderRequest, resp *IsLeaderResponse) error {
	is, err := p.manager.IsLeader()
	if err == nil {
		resp.Leader = is
	}
	return err
}

// LeaderLocationRequest is the rpc request
type LeaderLocationRequest struct {
}

// LeaderLocationResponse is the rpc response
type LeaderLocationResponse struct {
	Location *url.URL
}

// LeaderLocation returns the location of the leader
func (p *Manager) LeaderLocation(_ *http.Request, req *LeaderLocationRequest, resp *LeaderLocationResponse) error {
	u, err := p.manager.LeaderLocation()
	if err == nil {
		resp.Location = u
	}
	return err
}

// PlanRequest is the rpc request
type PlanRequest struct {
	Specs []types.Spec
}

// PlanResponse is the rpc response
type PlanResponse struct {
	Changes types.Changes
}

// Plan is the rpc method for Manager.Enforce
func (p *Manager) Plan(_ *http.Request, req *PlanRequest, resp *PlanResponse) error {
	changes, err := p.manager.Plan(req.Specs)
	if err != nil {
		return err
	}
	resp.Changes = changes
	return nil
}

// EnforceRequest is the rpc request
type EnforceRequest struct {
	Specs []types.Spec
}

// EnforceResponse is the rpc response
type EnforceResponse struct {
}

// Enforce is the rpc method for Manager.Enforce
func (p *Manager) Enforce(_ *http.Request, req *EnforceRequest, resp *EnforceResponse) error {
	return p.manager.Enforce(req.Specs)
}

// InspectRequest is the rpc request
type InspectRequest struct {
}

// InspectResponse is the rpc response
type InspectResponse struct {
	Objects []types.Object
}

// Inspect is the rpc method for Manager.Inspect
func (p *Manager) Inspect(_ *http.Request, req *InspectRequest, resp *InspectResponse) error {
	objects, err := p.manager.Inspect()
	if err != nil {
		return err
	}
	resp.Objects = objects
	return nil
}

// SpecsRequest is the rpc request
type SpecsRequest struct {
}

// SpecsResponse is the rpc response
type SpecsResponse struct {
	Specs []types.Spec
}

// Specs is the rpc method for Manager.Specs
func (p *Manager) Specs(_ *http.Request, req *SpecsRequest, resp *SpecsResponse) error {
	specs, err := p.manager.Specs()
	if err != nil {
		return err
	}
	resp.Specs = specs
	return nil
}

// PauseRequest is the rpc request
type PauseRequest struct {
	Specs []types.Spec
}

// PauseResponse is the rpc response
type PauseResponse struct {
}

// Pause is the rpc method for Manager.Pause
func (p *Manager) Pause(_ *http.Request, req *PauseRequest, resp *PauseResponse) error {
	return p.manager.Pause(req.Specs)
}

// TerminateRequest is the rpc request
type TerminateRequest struct {
	Specs []types.Spec
}

// TerminateResponse is the rpc response
type TerminateResponse struct {
}

// Terminate is the rpc method for Manager.Terminate
func (p *Manager) Terminate(_ *http.Request, req *TerminateRequest, resp *TerminateResponse) error {
	return p.manager.Terminate(req.Specs)
}
