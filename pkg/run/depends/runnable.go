package depends

import (
	"github.com/docker/infrakit/pkg/core"
	"github.com/docker/infrakit/pkg/plugin"
	"github.com/docker/infrakit/pkg/types"
)

// Runnable models an addressable object that can also be started.
type Runnable interface {
	core.Addressable
	// Options returns the options needed to start the plugin
	Options() *types.Any
	// Dependents return all the plugins this runnable depends on
	Dependents() (Runnables, error)
}

// Runnables represent a collection of Runnables
type Runnables []Runnable

// RunnableFrom creates a runnable from input name.  This is a simplification
// for cases where only a plugin name is used to reference another plugin.
func RunnableFrom(name plugin.Name) Runnable {
	lookup, _ := name.GetLookupAndType()
	return specQuery{
		spec: &types.Spec{
			Kind: lookup,
			Metadata: types.Metadata{
				Name: string(name),
			},
		},
	}
}

// AsRunnable returns the Runnable from a spec.
func AsRunnable(spec *types.Spec) Runnable {
	return &specQuery{
		Addressable: core.AsAddressable(spec),
		spec:        spec,
	}
}

type specQuery struct {
	core.Addressable
	spec *types.Spec
}

// Options returns the options
func (ps specQuery) Options() *types.Any {
	return ps.spec.Options
}

// Dependents returns the plugins depended on by this unit
func (ps specQuery) Dependents() (Runnables, error) {

	var interfaceSpec *types.InterfaceSpec
	if ps.spec.Version != "" {
		decoded := types.DecodeInterfaceSpec(ps.spec.Version)
		interfaceSpec = &decoded
	}
	dependentPlugins, err := Resolve(*ps.spec, ps.Kind(), interfaceSpec)
	if err != nil {
		return nil, err
	}
	// join this with the dependencies already in the spec
	out := Runnables{}
	out = append(out, dependentPlugins...)

	for _, d := range ps.spec.Depends {
		out = append(out, AsRunnable(&types.Spec{Kind: d.Kind, Metadata: types.Metadata{Name: d.Name}}))
	}

	log.Debug("dependents", "specQuery", ps, "result", out)
	return out, nil
}

// RunnablesFrom returns the Runnables from given slice of specs
func RunnablesFrom(specs []types.Spec) (Runnables, error) {
	// keyed by kind and the specQuery
	all := map[string]Runnable{}
	for _, s := range specs {
		copy := s
		q := AsRunnable(&copy)
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
