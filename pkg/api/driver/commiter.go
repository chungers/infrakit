package driver

import (
	logutil "github.com/docker/infrakit/pkg/log"
	"github.com/docker/infrakit/pkg/run/scope"
	"github.com/docker/infrakit/pkg/types"
)

type Committer interface {
	Builder
	Commit(scope.Scope) (types.Metadata, error)
}

var (
	Logger = logutil.New("module", "api/driver")

	connectionFactoriesLock = sync.RWMutex{}
	connectionFactories     = map[string]func(string, Options) (Scope, error){}
)

// Register registers a key / connection type with the constructor function for creating
// the scope.  This is used by different kinds of implementations such as local (unix sockets)
// and remotes. It's typically called from implementations init() functions.
func Register(connTypePrefix string, constructor func(string, Options) (Scope, error)) {
	connectionFactoriesLock.Lock()
	defer connectionFactoriesLock.Unlock()

	connectionFactories[connTypePrefix] = constructor
}
