package manager

import (
	"fmt"

	"github.com/docker/infrakit/pkg/cli"
	"github.com/docker/infrakit/pkg/discovery"
	logutil "github.com/docker/infrakit/pkg/log"
	"github.com/docker/infrakit/pkg/manager"
	manager_rpc "github.com/docker/infrakit/pkg/rpc/manager"
)

var log = logutil.New("module", "cli/v0/manager")

func init() {
	cli.Register(manager.InterfaceSpec,
		[]cli.CmdBuilder{
			Leader,
			Specs,
			Enforce,
			Describe,
			Terminate,
		})
}

// Load loads the typed object
func Load(plugins discovery.Plugins) (manager.Manager, error) {
	discovered, err := plugins.List()
	if err != nil {
		return nil, err
	}
	for _, endpoint := range discovered {
		manager, err := manager_rpc.NewClient(endpoint.Address)
		if err == nil {
			return manager, nil
		}
	}
	return nil, fmt.Errorf("no manager found")
}
