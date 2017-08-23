package internal

import (
	"fmt"

	logutil "github.com/docker/infrakit/pkg/log"
	"github.com/docker/infrakit/pkg/plugin"
)

var log = logutil.New("module", "rpc/internal")

const debugV = logutil.V(500)

// ServeKeyed returns a map containing keyed rpc objects
func ServeKeyed(listFunc func() (map[string]interface{}, error)) *Keyed {
	return &Keyed{listFunc: listFunc}
}

// ServeSingle returns a keyed that conforms to the net/rpc rpc call convention.
func ServeSingle(c interface{}) *Keyed {
	return ServeKeyed(func() (map[string]interface{}, error) {
		return map[string]interface{}{
			".": c,
		}, nil
	})
}

// Addressable is for RPC requests to implement so that the rpc handler can extract key from the RPC request.
type Addressable interface {
	Plugin() (plugin.Name, error)
}

// Keyed is a helper that manages multiple keyed rpc objects in a common namespace
type Keyed struct {
	listFunc func() (map[string]interface{}, error)
}

// Types returns the types exposed by this kind of RPC service
func (k *Keyed) Types() []string {
	m, err := k.listFunc()
	if err != nil {
		return nil
	}
	log.Debug("Types", "map", m, "V", debugV)
	types := []string{}
	for key := range m {
		types = append(types, fmt.Sprintf("%v", key))
	}
	return types
}

// Do performs work calling the work function once the request resolves to an object
func (k *Keyed) Do(request Addressable, work func(resolved interface{}) error) error {
	resolved, err := k.Resolve(request)
	log.Debug("Do", "resolved", resolved, "err", err, "req", request, "V", debugV)
	if err != nil {
		return err
	}
	return work(resolved)
}

// Resolve resolves input (a request object for example) that implements the Addressable interface into a plugin
func (k *Keyed) Resolve(request Addressable) (interface{}, error) {
	to, err := request.Plugin()
	log.Debug("Resolve", "to", to, "err", err, "req", request, "V", debugV)
	if err != nil {
		return nil, err
	}
	log.Debug("dispatching request", "to", to, "V", debugV)
	return k.Keyed(to)
}

// Keyed performs a lookup of the object by plugin name
func (k *Keyed) Keyed(name plugin.Name) (interface{}, error) {
	m, err := k.listFunc()
	if err != nil {
		return nil, err
	}
	l, subtype := name.GetLookupAndType()

	// Special case of single, unkeyed plugin object
	if l == "." || l == "" {
		if len(m) == 1 {
			for _, v := range m {
				return v, nil // first value
			}
		}
	}

	if subtype == "." || subtype == "" {
		if len(m) == 1 {
			for _, v := range m {
				return v, nil // first value
			}
		}
	}

	if p, has := m[subtype]; has {
		return p, nil
	}
	return nil, fmt.Errorf("not found: %v (key=%v)", name, subtype)
}
