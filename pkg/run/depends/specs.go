package depends

import (
	"strings"

	"github.com/docker/infrakit/pkg/plugin"
	"github.com/docker/infrakit/pkg/types"
)

/*
Examples:
{ kind:ingress,          name:lb1 }             => { kind:ingress,   plugin:ingress/lb1 }
{ kind:ingress,          name:us-east/lb1 }     => { kind:ingress,   plugin:us-east/lb1 }
{ kind:group,            name:workers }         => { kind:group,     plugin:group/workers }
{ kind:group,            name:us-east/workers } => { kind:group,     plugin:us-east/workers }
{ kind:resource,         name:vpc1 }            => { kind:resource,  plugin:resource/vpc1 }
{ kind:resource,         name:us-east/vpc1 }    => { kind:resource,  plugin:us-east/vpc1 }
{ kind:simulator/disk,   name:disk1 }           => { kind:simulator, plugin:simulator/disk        but query with disk1 }
{ kind:simulator/disk,   name:us-east/disk1 }   => { kind:simulator, plugin:us-east1/disk         but query with disk1 }
{ kind:aws/ec2-instance, name:host1 }           => { kind:aws,       plugin:aws/ec2-instance      but query with host1 }
{ kind:aws/ec2-instance, name:us-east/host1 }   => { kind:aws,       plugin:us-east1/ec2-instance but query with host1 }
*/

// Runnable captures all the information necessary to start a plugin
type Runnable interface {
	// Kind corresponds to the packages under pkg/run/v0
	Kind() string
	// Plugin returns the address of the rpc (endpoint)
	Plugin() plugin.Name
	// Options returns the options needed to start the plugin
	Options() *types.Any
	// Dependents return all the plugins this runnable depends on
	Dependents() (Runnables, error)
}

// RunnableFrom creates a runnable from input name.  This is a simplification
// for cases where only a plugin name is used to reference another plugin.
func RunnableFrom(name plugin.Name) Runnable {
	lookup, _ := name.GetLookupAndType()
	return specQuery{
		Spec: types.Spec{
			Kind: lookup,
			Metadata: types.Metadata{
				Name: string(name),
			},
		},
	}
}

// Runnables represent a collection of Runnables
type Runnables []Runnable

type specQuery struct {
	types.Spec
}

// Kind returns the kind to use for launching.  It's assumed these map to something in the launch Rules.
func (ps specQuery) Kind() string {
	// kind can be qualified, like aws/ec2-instance, but the kind is always the base.
	return strings.Split(ps.Spec.Kind, "/")[0]
}

// Plugin derives a plugin name from the record
func (ps specQuery) Plugin() plugin.Name {
	typeName := ""
	kind := strings.Split(ps.Spec.Kind, "/")
	if len(kind) > 1 {
		typeName = kind[1]
	}
	parts := strings.Split(ps.Spec.Metadata.Name, "/")
	if len(parts) > 1 {
		if typeName != "" {
			return plugin.NameFrom(parts[0], typeName)
		}
		return plugin.NameFrom(parts[0], parts[1])
	}
	if typeName != "" {
		return plugin.NameFrom(ps.Kind(), typeName)
	}
	return plugin.NameFrom(ps.Kind(), parts[0])
}

// Options returns the options
func (ps specQuery) Options() *types.Any {
	return ps.Spec.Options
}

// Dependents returns the plugins depended on by this unit
func (ps specQuery) Dependents() (Runnables, error) {

	var interfaceSpec *types.InterfaceSpec
	if ps.Spec.Version != "" {
		decoded := types.DecodeInterfaceSpec(ps.Spec.Version)
		interfaceSpec = &decoded
	}
	dependentPlugins, err := Resolve(ps.Spec, ps.Kind(), interfaceSpec)
	if err != nil {
		return nil, err
	}
	// join this with the dependencies already in the spec
	out := Runnables{}
	out = append(out, dependentPlugins...)

	for _, d := range ps.Depends {
		out = append(out, specQuery{types.Spec{Kind: d.Kind, Metadata: types.Metadata{Name: d.Name}}})
	}

	log.Debug("dependents", "specQuery", ps, "result", out)
	return out, nil
}

// RunnablesFrom returns the Runnables from given slice of specs
func RunnablesFrom(specs []types.Spec) (Runnables, error) {
	// keyed by kind and the specQuery
	all := map[string]Runnable{}
	for _, s := range specs {
		q := specQuery{s}
		all[q.Kind()] = q

		deps, err := q.Dependents()
		if err != nil {
			return nil, err
		}

		for _, d := range deps {
			// last win -- check for configs?  atm just focus on referenced objects
			all[d.Kind()] = d
		}
	}
	out := Runnables{}
	for _, s := range all {
		out = append(out, s)
	}
	return out, nil
}
