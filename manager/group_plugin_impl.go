package manager

import (
	"encoding/json"

	log "github.com/Sirupsen/logrus"
	rpc "github.com/docker/infrakit/rpc/group"
	"github.com/docker/infrakit/spi/group"
)

// proxyForGroupPlugin registers a group plugin that this manager will proxy for.
func (m *manager) proxyForGroupPlugin(name string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	endpoint, err := m.plugins.Find(name)
	if err != nil {
		return err
	}

	client, err := rpc.NewClient(endpoint.Protocol, endpoint.Address)
	if err != nil {
		return err
	}

	m.backendName = name
	m.backend = client
	return nil
}

// This implements the Group Plugin interface to support single group-only operations
// This is contrast with the declarative semantics of commit.  It offers an imperative
// operation by operation interface to the user.

func (m *manager) updateConfig(spec group.Spec) error {
	log.Debugln("Updating config", spec)
	m.lock.Lock()
	defer m.lock.Unlock()

	// Always read and then update with the current value.  Assumes the user's input
	// is always authoritative.
	stored := GlobalSpec{}

	err := m.snapshot.Load(&stored)
	if err != nil && err.Error() != "not-found" {
		return err
	}
	// if not-found ok to continue...

	if stored.Groups == nil {
		stored.Groups = map[group.ID]PluginSpec{}
	}

	buff, err := json.MarshalIndent(spec, "  ", "  ")
	if err != nil {
		return err
	}
	raw := json.RawMessage(buff)
	stored.Groups[spec.ID] = PluginSpec{
		Plugin:     m.backendName,
		Properties: &raw,
	}
	log.Debugln("Saving updated config", stored)

	return m.snapshot.Save(stored)
}

func (m *manager) WatchGroup(grp group.Spec) error {
	log.Debugln("Proxy WatchGroup:", grp)
	if err := m.updateConfig(grp); err != nil {
		return err
	}
	return m.backend.WatchGroup(grp)
}

func (m *manager) UnwatchGroup(id group.ID) error {
	log.Debugln("Proxy UnwatchGroup:", id)
	return m.backend.UnwatchGroup(id)
}

func (m *manager) InspectGroup(id group.ID) (group.Description, error) {
	log.Debugln("Proxy InspectGroup:", id)
	return m.backend.InspectGroup(id)
}

func (m *manager) DescribeUpdate(updated group.Spec) (string, error) {
	log.Debugln("Proxy DescribeUpdate:", updated)
	return m.backend.DescribeUpdate(updated)
}

func (m *manager) UpdateGroup(updated group.Spec) error {
	log.Debugln("Proxy UpdateGroup:", updated)
	if err := m.updateConfig(updated); err != nil {
		return err
	}
	return m.backend.UpdateGroup(updated)
}

func (m *manager) StopUpdate(id group.ID) error {
	log.Debugln("Proxy StopUpdate:", id)
	return m.backend.StopUpdate(id)
}

func (m *manager) DestroyGroup(id group.ID) error {
	log.Debugln("Proxy DestroyGroup:", id)
	return m.backend.DestroyGroup(id)
}

func (m *manager) DescribeGroups() ([]group.Spec, error) {
	log.Debugln("Proxy DescribeGroups")
	return m.backend.DescribeGroups()
}
