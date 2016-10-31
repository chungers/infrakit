package manager

import (
	"github.com/docker/infrakit/manager"
)

// RPCServer returns a RPCService that conforms to the net/rpc rpc call convention.
func RPCServer(m manager.Manager) RPCService {
	return &Manager{manager: m}
}

// Manager the exported type needed to conform to json-rpc call convention
type Manager struct {
	manager manager.Manager
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
