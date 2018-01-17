package registry

import (
	"fmt"
	"strings"
	"sync"

	"github.com/docker/infrakit/pkg/api"
)

var (
	connectionFactoriesLock = sync.RWMutex{}
	connectionFactories     = map[string]func(string, api.Options) (api.Scope, error){}
)

// Register registers a key / connection type with the constructor function for creating
// the scope.  This is used by different kinds of implementations such as local (unix sockets)
// and remotes. It's typically called from implementations init() functions.
func Register(connTypePrefix string, constructor func(string, api.Options) (api.Scope, error)) {
	connectionFactoriesLock.Lock()
	defer connectionFactoriesLock.Unlock()

	connectionFactories[connTypePrefix] = constructor
}

func Find(connect string, opt api.Options) (api.Scope, error) {

	connectionFactoriesLock.RLock()
	defer connectionFactoriesLock.RUnlock()

	for prefix, constructor := range connectionFactories {

		// e.g. local:// or tcp://
		if strings.Index(connect, prefix) == 0 {
			return constructor(connect, opt)
		}
	}

	return nil, fmt.Errorf("not found: %v", connect)
}
