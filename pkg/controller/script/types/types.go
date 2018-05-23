package types

import (
	"context"
	"fmt"

	"github.com/docker/infrakit/pkg/fsm"
	"github.com/docker/infrakit/pkg/types"
)

// Use contains the playbook 'use' short name with the source URL
type Use struct {
	URL string
	As  string
}

// Step contains information about a single step of callable execution
type Step struct {
	Call   string
	Params *types.Any

	// ResultIsBytes is set true if the result from the call is just byte slice (default is string)
	ResultIsBytes bool `json:"result_is_bytes,omitempty" yaml:"result_is_bytes,omitempty"`

	// Target references a name in the map of Properties.Targets
	Target          *string
	Parallelism     *int
	Retries         *int
	WaitBeforeRetry *fsm.Tick
}

// Properties is the schema of the configuration in the types.Spec.Properties
type Properties struct {
	Use     []Use
	Targets map[string]*types.Any
	Steps   []Step
}

// ModelProperties contain fsm tuning parameters
type ModelProperties struct {
	TickUnit          types.Duration
	WaitBeforeStart   fsm.Tick
	ChannelBufferSize int

	// FSM tuning options
	fsm.Options `json:",inline" yaml:",inline"`
}

// Options is the controller options that is used at start up of the process.  It's one-time
type Options struct {
	// MinChannelBufferSize is the min size of the buffered chanels
	MinChannelBufferSize int

	// ModelProperties are the config parameters for the workflow model
	ModelProperties `json:",inline" yaml:",inline"`

	// StepDeadline is the timeout for executing a step.
	StepDeadline types.Duration
}

// Validate validates the input properties
func (p Properties) Validate(ctx context.Context) error {
	return nil
}

// Validate validates the controller's options
func (p Options) Validate(ctx context.Context) error {
	if p.MinChannelBufferSize == 0 {
		return fmt.Errorf("min channel buffer size cannot be 0")
	}
	if p.ChannelBufferSize < p.MinChannelBufferSize {
		return fmt.Errorf("channel buffer size can't be less than %v", p.MinChannelBufferSize)
	}
	return nil
}
