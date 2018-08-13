package internal

import (
	"github.com/docker/infrakit/pkg/spi/controller"
	"github.com/docker/infrakit/pkg/spi/instance"
	"github.com/docker/infrakit/pkg/types"
)

type collectionInstance struct {
	controller controller.Controller
}

func newCollectionInstancePlugin(controller controller.Controller) *collectionInstance {
	return &collectionInstance{
		controller: controller,
	}
}

// Validate performs local validation on a provision request.
func (c *collectionInstance) Validate(req *types.Any) error {
	return nil
}

// Provision creates a new instance based on the spec.
func (c *collectionInstance) Provision(spec instance.Spec) (*instance.ID, error) {

	controllerSpec := types.Spec{}

	if err := spec.Properties.Decode(&controllerSpec); err != nil {
		return nil, err
	}

	if spec.LogicalID != nil {
		if controllerSpec.Metadata.Name == "" {
			controllerSpec.Metadata.Name = string(*spec.LogicalID)
		}

		if controllerSpec.Metadata.Identity == nil {
			controllerSpec.Metadata.Identity = &types.Identity{ID: string(*spec.LogicalID)}
		}
	}

	if spec.Init != "" {
		// warn - ignored
	}

	if len(spec.Attachments) > 0 {
		// warn - ignored
	}

	// merge tags -- the inner spec's tags overrides the outer (the instance spec's)
	if controllerSpec.Metadata.Tags == nil {
		controllerSpec.Metadata.Tags = map[string]string{}
	}

	for k, v := range spec.Tags {
		if _, has := controllerSpec.Metadata.Tags[k]; !has {
			controllerSpec.Metadata.Tags[k] = v
		}
	}

	return nil, nil
}

// Label labels the instance
func (c *collectionInstance) Label(instance instance.ID, labels map[string]string) error {
	return nil
}

// Destroy terminates an existing instance.
func (c *collectionInstance) Destroy(instance instance.ID, context instance.Context) error {
	return nil
}

// DescribeInstances returns descriptions of all instances matching all of the provided tags.
// The properties flag indicates the client is interested in receiving details about each instance.
func (c *collectionInstance) DescribeInstances(labels map[string]string,
	properties bool) ([]instance.Description, error) {

	return nil, nil
}
