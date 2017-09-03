package adapter

import (
	"fmt"

	"github.com/docker/infrakit/pkg/controller"
	"github.com/docker/infrakit/pkg/core"
	"github.com/docker/infrakit/pkg/plugin"
	"github.com/docker/infrakit/pkg/spi/group"
	"github.com/docker/infrakit/pkg/types"
)

// AsController returns a Controller, possibly with a scope of the group ID.
func AsController(addr core.Addressable, g group.Plugin, scope *group.ID) controller.Controller {
	return &gController{
		Addressable: addr, // address of the plugin backend
		scope:       scope,
		plugin:      g,
	}
}

// This controller is used to implement a generic controller *as well as* a named controller
// for a group.  When id is specified, the controller is scoped to the id.  When input is missing
// id, it will be injected.  If input has mismatched id, requests will error.
type gController struct {
	core.Addressable
	scope  *group.ID
	plugin group.Plugin
}

func (c *gController) translateSpec(s types.Spec) (group.Spec, error) {
	spec := s
	gSpec := group.Spec{
		Properties: spec.Properties,
	}
	addressable := core.AsAddressable(&spec)
	if c.scope == nil {
		if addressable.Instance() == "" {
			return gSpec, fmt.Errorf("no group name")
		}
		gSpec.ID = group.ID(addressable.Instance())
		return gSpec, nil
	} else {
		if addressable.Instance() != string(*c.scope) {
			return group.Spec{}, fmt.Errorf("wrong group: %v", *c.scope)
		}
		gSpec.ID = *c.scope
	}

	return gSpec, nil
}

func objectFromSpec(spec types.Spec) types.Object {
	return types.Object{
		Spec: spec,
	}
}

func (c *gController) Plan(operation controller.Operation,
	spec types.Spec) (object types.Object, plan controller.Plan, err error) {

	gSpec, e := c.translateSpec(spec)
	if e != nil {
		err = e
		return
	}

	plan = controller.Plan{}
	object = objectFromSpec(spec)
	if resp, cerr := c.plugin.CommitGroup(gSpec, true); cerr == nil {
		plan.Message = []string{resp}
	} else {
		err = cerr
	}
	return
}

func (c *gController) Commit(operation controller.Operation, spec types.Spec) (object types.Object, err error) {
	gSpec, e := c.translateSpec(spec)
	if e != nil {
		err = e
		return
	}

	object = objectFromSpec(spec)
	switch operation {
	case controller.Enforce:
		_, err = c.plugin.CommitGroup(gSpec, false)
	case controller.Destroy:
		err = c.plugin.DestroyGroup(group.ID(spec.Metadata.Name))
	}
	return
}

func (c *gController) helpFind(search *types.Metadata) (gspecs map[group.ID]group.Spec, err error) {
	gspecs = map[group.ID]group.Spec{}

	all := []group.Spec{}
	all, err = c.plugin.InspectGroups()
	if err != nil {
		return
	}

	for _, gspec := range all {
		gspecs[gspec.ID] = gspec
		if c.scope != nil && *c.scope == gspec.ID {
			break
		}
	}
	return
}

func (c *gController) fromGroupSpec(gspec group.Spec) types.Spec {
	lookup, sub := c.Plugin().GetLookupAndType()
	name := plugin.NameFrom(lookup, string(gspec.ID))
	if sub == "" {
		name = plugin.Name(string(gspec.ID))
	}

	return types.Spec{
		Kind:    c.Kind(),
		Version: group.InterfaceSpec.Encode(),
		Metadata: types.Metadata{
			Identity: &types.Identity{ID: string(gspec.ID)},
			Name:     string(name),
		},
		Properties: gspec.Properties,
		Options:    nil, // TODO(chungers) -- here's a loss of information in the old format
	}
}

func (c *gController) Describe(search *types.Metadata) (objects []types.Object, err error) {
	var gspecs map[group.ID]group.Spec
	gspecs, err = c.helpFind(search)
	if err != nil {
		return
	}

	objects = []types.Object{}

	for gid, gspec := range gspecs {

		match := true
		if search != nil {
			query := core.NewAddressableFromMetadata(c.Kind(), *search)
			match = query.Instance() == string(gid)
		}

		if match {
			var desc group.Description
			desc, err = c.plugin.DescribeGroup(gid)
			if err != nil {
				return
			}
			state := types.Object{
				Spec:  c.fromGroupSpec(gspec),
				State: types.AnyValueMust(desc),
			}
			objects = append(objects, state)
		}
	}
	return
}

func (c *gController) Free(search *types.Metadata) (objects []types.Object, err error) {
	objects, err = c.Describe(search)
	if err != nil {
		return
	}
	for _, object := range objects {
		addr := core.AsAddressable(&object.Spec)
		err = c.plugin.FreeGroup(group.ID(addr.Instance()))
		if err != nil {
			return
		}
	}
	return
}
