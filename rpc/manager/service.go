package manager

import (
	"github.com/docker/infrakit/manager"
	group_rpc "github.com/docker/infrakit/rpc/group"
)

// RPCServer returns a RPCService that conforms to the net/rpc rpc call convention.
func RPCServer(m manager.Service) RPCService {
	return &Manager{
		RPCService: group_rpc.PluginServer(m),
		manager:    m,
	}
}

// Manager the exported type needed to conform to json-rpc call convention
type Manager struct {
	group_rpc.RPCService
	manager manager.Service
}

// Commit signals a manager to commit to changes
func (p *Manager) Commit(req *CommitRequest, resp *CommitResponse) error {
	err := p.manager.Commit()
	if err != nil {
		return err
	}
	resp.OK = true
	return nil
}
