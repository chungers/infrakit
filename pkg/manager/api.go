package manager

import (
	"net/url"

	"github.com/docker/infrakit/pkg/plugin"
	"github.com/docker/infrakit/pkg/spi"
	"github.com/docker/infrakit/pkg/types"
)

var (
	// InterfaceSpec is the current name and version of the Instance API.
	InterfaceSpec = spi.InterfaceSpec{
		Name:    "Manager",
		Version: "0.1.0",
	}
)

// Leadership is the interface for getting information about the current leader node
type Leadership interface {
	// IsLeader returns true only if for certain this is a leader. False if not or unknown.
	IsLeader() (bool, error)
}

// Manager is the interface for interacting locally or remotely with the manager
type Manager interface {
	Leadership

	// Supervising returns a list of plugins this manager is supervising.  By supervising,
	// the manager provides persistence and leadership checking in HA mode.
	Supervising() ([]plugin.Metadata, error)

	// LeaderLocation returns the location of the leader
	LeaderLocation() (*url.URL, error)

	// Plan returns the changes to be made
	Plan(specs []types.Spec) (types.Changes, error)

	// Enforce enforces infrastructure state to match that of the specs
	Enforce(specs []types.Spec) error

	// Specs returns the specs this manager tasked with enforcing
	Specs() ([]types.Spec, error)

	// Inspect returns the current state of the infrastructure
	Inspect() ([]types.Object, error)

	// Pause pauses all the controllers in the given set of specs. If the input is emtpy and there's a
	// locally stored version, that version will be used.
	Pause([]types.Spec) error

	// Terminate destroys all resources associated with the specs. If the specs is empty and there's a
	// locally stored, version, that version will be used.
	Terminate(specs []types.Spec) error
}
