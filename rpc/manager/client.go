package manager

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"

	"github.com/docker/infrakit/manager"
)

// NewClient returns a plugin interface implementation connected to a remote plugin
func NewClient(protocol, addr string) (manager.Manager, error) {
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
	return c.rpc.Call("Manager.Commit", req, resp)
}
