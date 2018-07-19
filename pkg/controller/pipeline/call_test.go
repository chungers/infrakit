package pipeline

import (
	"testing"
	//	pipeline_types "github.com/docker/infrakit/pkg/controller/pipeline/types"
	//	"github.com/stretchr/testify/require"
)

func intPtr(v int) *int {
	vv := v
	return &vv
}

func TestShards(t *testing.T) {

	// 	require.Equal(t, shardsT{}, computeShards(pipeline_types.Properties{}, Step{}, []string{}))

	// 	// If no parallelism is specified, there are no shards.
	// 	require.Equal(t, shardsT{}, computeShards(pipeline_types.Properties{}, Step{}, []string{
	// 		"h1", "h2", "h3",
	// 	}))

	// 	require.Equal(t, shardsT{}, computeShards(pipeline_types.Properties{
	// 		Source: pipeline_types.Targets{
	// 			Parallelism: intPtr(2),
	// 		},
	// 	}, Step{
	// 		Step: pipeline_types.Step{},
	// 	}, []string{}))

	// 	require.Equal(t, shardsT{
	// 		{
	// 			Step:    pipeline_types.Step{},
	// 			targets: []string{"h1", "h2"},
	// 		},
	// 		{
	// 			Step:    pipeline_types.Step{},
	// 			targets: []string{"h3", "h4"},
	// 		},
	// 	}, computeShards(pipeline_types.Properties{
	// 		Source: pipeline_types.Targets{
	// 			Parallelism: intPtr(2),
	// 		},
	// 	}, Step{
	// 		Step: pipeline_types.Step{},
	// 	}, []string{
	// 		"h1", "h2", "h3", "h4",
	// 	}))

	// 	require.Equal(t, shardsT{
	// 		{
	// 			Step:    pipeline_types.Step{},
	// 			targets: []string{"h1", "h2"},
	// 		},
	// 		{
	// 			Step:    pipeline_types.Step{},
	// 			targets: []string{"h3", "h4"},
	// 		},
	// 		{
	// 			Step:    pipeline_types.Step{},
	// 			targets: []string{"h5"},
	// 		},
	// 	}, computeShards(pipeline_types.Properties{
	// 		Source: pipeline_types.Targets{
	// 			Parallelism: intPtr(2),
	// 		},
	// 	}, Step{
	// 		Step: pipeline_types.Step{},
	// 	}, []string{
	// 		"h1", "h2", "h3", "h4", "h5",
	// 	}))

}
