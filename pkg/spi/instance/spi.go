package instance

import (
	"encoding/json"
	"github.com/docker/infrakit/pkg/spi"
)

// InterfaceSpec is the current name and version of the Instance API.
var InterfaceSpec = spi.InterfaceSpec{
	Name:    "Instance",
	Version: "0.1.0",
}

func init() {
	spi.RegisterInterface(InterfaceSpec)
}

// Plugin is a vendor-agnostic API used to create and manage resources with an infrastructure provider.
type Plugin interface {
	// Validate performs local validation on a provision request.
	Validate(req json.RawMessage) error

	// Provision creates a new instance based on the spec.
	Provision(spec Spec) (*ID, error)

	// Destroy terminates an existing instance.
	Destroy(instance ID) error

	// DescribeInstances returns descriptions of all instances matching all of the provided tags.
	DescribeInstances(tags map[string]string) ([]Description, error)
}
