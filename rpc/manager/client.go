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
func NewClient(protocol, addr string) (manager.Service, error) {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		return nil, err
	}
	gp, err := group_rpc.NewClientFromConn(conn)
	if err != nil {
		return nil, err
	}
	return &client{Plugin: gp, rpc: jsonrpc.NewClient(conn)}, nil
}

type client struct {
	group.Plugin
	rpc *rpc.Client
}

// Commit signals the manager to commit the changes
func (c *client) Commit() error {
	req := &CommitRequest{}
	resp := &CommitResponse{}
	return c.rpc.Call("Group.Commit", req, resp)
}
