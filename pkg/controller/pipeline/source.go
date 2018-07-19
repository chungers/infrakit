package pipeline

import (
	//	"github.com/docker/infrakit/pkg/controller/internal"
	pipeline_types "github.com/docker/infrakit/pkg/controller/pipeline/types"
	"github.com/docker/infrakit/pkg/run/scope"
	"github.com/docker/infrakit/pkg/types"
)

type sourceCancel func()

// targetParsers is a list of parsers that takes a blob *types.Any to a list
// of hosts/ targets
type targetParsers []func(scope.Scope, *types.Any) ([]string, error)

func targetsFrom(spec pipeline_types.Targets) (<-chan string, sourceCancel, error) {

	source := make(chan string)

	return source, func() {
		close(source)
	}, nil
}

func (l targetParsers) targets(scope scope.Scope, properties pipeline_types.Properties, blob *types.Any) []string {
	// Go through all the defined parsers and return whatever on first success.
	for _, parser := range l {
		if t, err := parser(scope, blob); err == nil {
			return t
		}
	}
	return nil
}
