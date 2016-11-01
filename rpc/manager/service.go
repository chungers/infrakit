package manager

import (
	"github.com/docker/infrakit/manager"
	group_rpc "github.com/docker/infrakit/rpc/group"
)

// RPCServer returns a RPCService that conforms to the net/rpc rpc call convention.
func RPCServer(m manager.Group) RPCService {
	return &Group{manager: m}
}

// Group the exported type needed to conform to json-rpc call convention
type Group struct {
	manager manager.Group
}

// Commit signals a manager to commit to changes
func (p *Group) Commit(req *CommitRequest, resp *CommitResponse) error {
	err := p.manager.Commit()
	if err != nil {
		return err
	}
	resp.OK = true
	return nil
}

// WatchGroup is the rpc method to watch a group
func (p *Group) WatchGroup(req *group_rpc.WatchGroupRequest, resp *group_rpc.WatchGroupResponse) error {
	err := p.manager.WatchGroup(req.Spec)
	if err != nil {
		return err
	}
	resp.OK = true
	return nil
}

// UnwatchGroup is the rpc method to unwatch a group
func (p *Group) UnwatchGroup(req *group_rpc.UnwatchGroupRequest, resp *group_rpc.UnwatchGroupResponse) error {
	err := p.manager.UnwatchGroup(req.ID)
	if err != nil {
		return err
	}
	resp.OK = true
	return nil
}

// InspectGroup is the rpc method to inspect a group
func (p *Group) InspectGroup(req *group_rpc.InspectGroupRequest, resp *group_rpc.InspectGroupResponse) error {
	desc, err := p.manager.InspectGroup(req.ID)
	if err != nil {
		return err
	}
	resp.Description = desc
	return nil
}

// DescribeUpdate is the rpc method to describe an update without performing it
func (p *Group) DescribeUpdate(req *group_rpc.DescribeUpdateRequest, resp *group_rpc.DescribeUpdateResponse) error {
	plan, err := p.manager.DescribeUpdate(req.Spec)
	if err != nil {
		return err
	}
	resp.Plan = plan
	return nil
}

// UpdateGroup is the rpc method to actually updating a group
func (p *Group) UpdateGroup(req *group_rpc.UpdateGroupRequest, resp *group_rpc.UpdateGroupResponse) error {
	err := p.manager.UpdateGroup(req.Spec)
	if err != nil {
		return err
	}
	resp.OK = true
	return nil
}

// StopUpdate is the rpc method to stop a current update
func (p *Group) StopUpdate(req *group_rpc.StopUpdateRequest, resp *group_rpc.StopUpdateResponse) error {
	err := p.manager.StopUpdate(req.ID)
	if err != nil {
		return err
	}
	resp.OK = true
	return nil
}

// DestroyGroup is the rpc method to destroy a group
func (p *Group) DestroyGroup(req *group_rpc.DestroyGroupRequest, resp *group_rpc.DestroyGroupResponse) error {
	err := p.manager.DestroyGroup(req.ID)
	if err != nil {
		return err
	}
	resp.OK = true
	return nil
}

// DescribeGroups is the rpc method to describe groups
func (p *Group) DescribeGroups(req *group_rpc.DescribeGroupsRequest, resp *group_rpc.DescribeGroupsResponse) error {
	groups, err := p.manager.DescribeGroups()
	if err != nil {
		return err
	}
	resp.Groups = groups
	return nil
}
