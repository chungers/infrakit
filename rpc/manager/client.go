package manager

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"

	"github.com/docker/infrakit/manager"
	group_rpc "github.com/docker/infrakit/rpc/group"
	"github.com/docker/infrakit/spi/group"
)

// NewClient returns a plugin interface implementation connected to a remote plugin
func NewClient(protocol, addr string) (manager.Group, error) {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		return nil, err
	}
	return &client{rpc: jsonrpc.NewClient(conn)}, nil
}

type client struct {
	rpc *rpc.Client
}

// Commit signals the manager to commit the changes
func (c *client) Commit() error {
	req := &CommitRequest{}
	resp := &CommitResponse{}
	return c.rpc.Call("Group.Commit", req, resp)
}

func (c *client) WatchGroup(grp group.Spec) error {
	req := &group_rpc.WatchGroupRequest{Spec: grp}
	resp := &group_rpc.WatchGroupResponse{}
	err := c.rpc.Call("Group.WatchGroup", req, resp)
	if err != nil {
		return err
	}
	resp.OK = true
	return nil
}

func (c *client) UnwatchGroup(id group.ID) error {
	req := &group_rpc.UnwatchGroupRequest{ID: id}
	resp := &group_rpc.UnwatchGroupResponse{}
	err := c.rpc.Call("Group.UnwatchGroup", req, resp)
	if err != nil {
		return err
	}
	resp.OK = true
	return nil
}

func (c *client) InspectGroup(id group.ID) (group.Description, error) {
	req := &group_rpc.InspectGroupRequest{ID: id}
	resp := &group_rpc.InspectGroupResponse{}
	err := c.rpc.Call("Group.InspectGroup", req, resp)
	return resp.Description, err
}

func (c *client) DescribeUpdate(updated group.Spec) (string, error) {
	req := &group_rpc.DescribeUpdateRequest{Spec: updated}
	resp := &group_rpc.DescribeUpdateResponse{}
	err := c.rpc.Call("Group.DescribeUpdate", req, resp)
	return resp.Plan, err
}

func (c *client) UpdateGroup(updated group.Spec) error {
	req := &group_rpc.UpdateGroupRequest{Spec: updated}
	resp := &group_rpc.UpdateGroupResponse{}
	err := c.rpc.Call("Group.UpdateGroup", req, resp)
	if err != nil {
		return err
	}
	resp.OK = true
	return nil
}

func (c *client) StopUpdate(id group.ID) error {
	req := &group_rpc.StopUpdateRequest{ID: id}
	resp := &group_rpc.StopUpdateResponse{}
	err := c.rpc.Call("Group.StopUpdate", req, resp)
	if err != nil {
		return err
	}
	resp.OK = true
	return nil
}

func (c *client) DestroyGroup(id group.ID) error {
	req := &group_rpc.DestroyGroupRequest{ID: id}
	resp := &group_rpc.DestroyGroupResponse{}
	err := c.rpc.Call("Group.DestroyGroup", req, resp)
	if err != nil {
		return err
	}
	resp.OK = true
	return nil
}

func (c *client) DescribeGroups() ([]group.Spec, error) {
	req := &group_rpc.DescribeGroupsRequest{}
	resp := &group_rpc.DescribeGroupsResponse{}
	err := c.rpc.Call("Group.DescribeGroups", req, resp)
	return resp.Groups, err
}
