package instance

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/docker/infrakit/pkg/plugin"
	"github.com/docker/infrakit/pkg/spi/instance"
)

// PluginServer returns a RPCService that conforms to the net/rpc rpc call convention.
func PluginServer(p instance.Plugin) *Instance {
	return &Instance{plugin: p}
}

// Instance is the JSON RPC service representing the Instance Plugin.  It must be exported in order to be
// registered by the rpc server package.
type Instance struct {
	plugin instance.Plugin
}

// Info returns a metadata object about the plugin, if the plugin implements it.  See plugin.Informer
func (p *Instance) Info() plugin.Info {
	if m, is := p.plugin.(plugin.Vendor); is {
		return m.Info()
	}
	return plugin.NoInfo
}

// ExampleProperties returns an example properties used by the plugin
func (p *Instance) ExampleProperties() *json.RawMessage {
	if i, is := p.plugin.(plugin.InputExample); is {
		return i.ExampleProperties()
	}
	return nil
}

// Validate performs local validation on a provision request.
func (p *Instance) Validate(_ *http.Request, req *ValidateRequest, resp *ValidateResponse) error {
	if req.Properties == nil {
		return errors.New("Request Properties must be set")
	}

	err := p.plugin.Validate(*req.Properties)
	if err != nil {
		return err
	}
	resp.OK = true
	return nil
}

// Provision creates a new instance based on the spec.
func (p *Instance) Provision(_ *http.Request, req *ProvisionRequest, resp *ProvisionResponse) error {
	id, err := p.plugin.Provision(req.Spec)
	if err != nil {
		return err
	}
	resp.ID = id
	return nil
}

// Destroy terminates an existing instance.
func (p *Instance) Destroy(_ *http.Request, req *DestroyRequest, resp *DestroyResponse) error {
	err := p.plugin.Destroy(req.Instance)
	if err != nil {
		return err
	}
	resp.OK = true
	return nil
}

// DescribeInstances returns descriptions of all instances matching all of the provided tags.
func (p *Instance) DescribeInstances(_ *http.Request, req *DescribeInstancesRequest, resp *DescribeInstancesResponse) error {
	desc, err := p.plugin.DescribeInstances(req.Tags)
	if err != nil {
		return err
	}
	resp.Descriptions = desc
	return nil
}