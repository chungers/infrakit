package script

import (
	"testing"

	"github.com/docker/infrakit/pkg/controller/script/types"
	"github.com/stretchr/testify/require"
)

func intPtr(v int) *int {
	vv := v
	return &vv
}

func TestShards(t *testing.T) {

	require.Equal(t, shardsT{}, computeShards(Step{}, []string{}))

	// If no parallelism is specified, there are no shards.
	require.Equal(t, shardsT{}, computeShards(Step{}, []string{
		"h1", "h2", "h3",
	}))

	require.Equal(t, shardsT{}, computeShards(Step{
		Step: types.Step{
			Parallelism: intPtr(2),
		},
	}, []string{}))

	require.Equal(t, shardsT{
		{
			Step: types.Step{
				Parallelism: intPtr(2),
			},
			targets: []string{"h1", "h2"},
		},
		{
			Step: types.Step{
				Parallelism: intPtr(2),
			},
			targets: []string{"h3", "h4"},
		},
	}, computeShards(Step{
		Step: types.Step{
			Parallelism: intPtr(2),
		},
	}, []string{
		"h1", "h2", "h3", "h4",
	}))

	require.Equal(t, shardsT{
		{
			Step: types.Step{
				Parallelism: intPtr(2),
			},
			targets: []string{"h1", "h2"},
		},
		{
			Step: types.Step{
				Parallelism: intPtr(2),
			},
			targets: []string{"h3", "h4"},
		},
		{
			Step: types.Step{
				Parallelism: intPtr(2),
			},
			targets: []string{"h5"},
		},
	}, computeShards(Step{
		Step: types.Step{
			Parallelism: intPtr(2),
		},
	}, []string{
		"h1", "h2", "h3", "h4", "h5",
	}))

}
