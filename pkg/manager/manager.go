package manager

import (
	"github.com/docker/infrakit/pkg/controller"
	"github.com/docker/infrakit/pkg/discovery"
	"github.com/docker/infrakit/pkg/leader"
	"github.com/docker/infrakit/pkg/manager/internal"
	"github.com/docker/infrakit/pkg/plugin"
	"github.com/docker/infrakit/pkg/spi/group"
	"github.com/docker/infrakit/pkg/store"
)

// Backend is the admin / server-side interface
type Backend interface {
	Manager

	group.Plugin

	Controllers() (map[string]controller.Controller, error)
	Groups() (map[group.ID]group.Plugin, error)

	Start() (<-chan struct{}, error)
	Stop()
}

// NewManager returns a manager implementation
func NewManager(name plugin.Name,
	plugins discovery.Plugins,
	leader leader.Detector,
	leaderStore leader.Store,
	snapshot store.Snapshot,
	backendName string) Backend {
	return internal.NewBackend(name, plugins, leader, leaderStore, snapshot, backendName)
}
