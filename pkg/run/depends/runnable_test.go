package depends

import (
	"testing"

	"github.com/docker/infrakit/pkg/plugin"
	"github.com/docker/infrakit/pkg/types"
	"github.com/stretchr/testify/require"
)

func mustSpec(s types.Spec, err error) *types.Spec {
	if err != nil {
		panic(err)
	}
	copy := s
	return &copy
}

func TestRunnable(t *testing.T) {
	v := types.DecodeInterfaceSpec("Test/0.1")
	Register("TestRunnable", v, func(spec types.Spec) (Runnables, error) {
		copy := spec
		return Runnables{AsRunnable(&copy)}, nil
	})

	runnable := AsRunnable(mustSpec(types.SpecFromString(`
kind: group
metadata:
  name: workers
properties:
  max: 100
  min: 10
options:
  poll: 10
`)))

	require.Equal(t, "group", runnable.Kind())
	require.Equal(t, plugin.Name("group/workers"), runnable.Plugin())
	require.Equal(t, "workers", runnable.Instance())
	options := map[string]int{}
	require.NoError(t, runnable.Options().Decode(&options))
	require.Equal(t, map[string]int{"poll": 10}, options)

	deps, err := runnable.Dependents()
	require.NoError(t, err)
	require.Equal(t, Runnables{}, deps)
}

func TestRunnableWithDepends(t *testing.T) {
	v := types.DecodeInterfaceSpec("group/0.1")
	Register("group", v, func(spec types.Spec) (Runnables, error) {
		// This just echos back whatever comes in
		copy := spec
		return Runnables{AsRunnable(&copy)}, nil
	})

	runnable := AsRunnable(mustSpec(types.SpecFromString(`
kind: group
version: group/0.1
metadata:
  name: workers
properties:
  max: 100
  min: 10
options:
  poll: 10
`)))

	require.Equal(t, "group", runnable.Kind())
	require.Equal(t, plugin.Name("group/workers"), runnable.Plugin())
	require.Equal(t, "workers", runnable.Instance())
	options := map[string]int{}
	require.NoError(t, runnable.Options().Decode(&options))
	require.Equal(t, map[string]int{"poll": 10}, options)

	deps, err := runnable.Dependents()
	require.NoError(t, err)
	require.Equal(t, Runnables{runnable}, deps)
}
