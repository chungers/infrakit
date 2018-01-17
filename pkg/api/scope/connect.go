package scope

import (
	"github.com/docker/infrakit/pkg/api"
	"github.com/docker/infrakit/pkg/api/internal/registry"
)

// Connect
func Connect(connect string, opt api.Options) (api.Scope, error) {
	return registry.Find(connect, opt)
}
