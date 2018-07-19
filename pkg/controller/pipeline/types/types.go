package types

import (
	"context"
	"fmt"

	"github.com/docker/infrakit/pkg/fsm"
	"github.com/docker/infrakit/pkg/types"
)

// Result captures the output or error of the call.
type Result struct {
	Step         Step
	Target       string      `json:",omitempty" yaml:",omitempty"`
	Output       interface{} `json:",omitempty" yaml:",omitempty"`
	Error        error       `json:"-" yaml:"-"`
	ErrorMessage string      `json:"Error,omitempty" yaml:"Error,omitempty"`
}

// Use contains the playbook 'use' short name with the source URL
type Use struct {
	URL string
	As  string
}

// Step contains information about a single step of callable execution
type Step struct {
	Call   string
	Params *types.Any `json:",omitempty" yaml:",omitempty"`

	// ResultTemplate  is an expression with that will be evaluated as a template.  The expression can have
	// parts separated by `;` and each part will be enclosed in `{{}}` as interpolated template expression.
	ResultTemplate *string `json:",omitempty" yaml:",omitempty"`

	// ResultTemplateVar is the name of the var that can be used to retrieve a non-string/bytes, typed value
	// by variable name.  For example, you can retrieve a typed value named x this way:
	// resultTemplate: .Bytes | jsonDecode | var `x`
	// resultTemplateVar: x
	ResultTemplateVar *string `json:",omitempty" yaml:",omitempty"`

	Retries         *int      `json:",omitempty" yaml:",omitempty"`
	WaitBeforeRetry *fsm.Tick `json:",omitempty" yaml:",omitempty"`
}

// Targets specifies the source targets, ie. objects to operate the pipeline on.
// Each target will be tracked by the FSM as steps in the pipeline operates on it.
// Each stage has states associated with it and the states are tracked and visible.
type Targets struct {
	From        *types.Any
	Parallelism *int `json:",omitempty" yaml:",omitempty"`
}

// Properties is the schema of the configuration in the types.Spec.Properties
type Properties struct {
	Use    []Use   `json:",omitempty" yaml:",omitempty"`
	Source Targets `json:",omitempty" yaml:",omitempty"`
	Steps  []Step
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
