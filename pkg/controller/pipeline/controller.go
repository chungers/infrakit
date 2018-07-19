package pipeline

import (
	"time"

	"github.com/docker/infrakit/pkg/controller/internal"
	script "github.com/docker/infrakit/pkg/controller/pipeline/types"
	"github.com/docker/infrakit/pkg/fsm"
	logutil "github.com/docker/infrakit/pkg/log"
	"github.com/docker/infrakit/pkg/run/scope"
	"github.com/docker/infrakit/pkg/spi/controller"
	"github.com/docker/infrakit/pkg/spi/event"
	"github.com/docker/infrakit/pkg/spi/metadata"
	"github.com/docker/infrakit/pkg/types"
)

var (
	log     = logutil.New("module", "controller/script")
	debugV  = logutil.V(500)
	debugV2 = logutil.V(1000)

	// DefaultModelProperties is the default properties for the fsm model
	DefaultModelProperties = script.ModelProperties{
		TickUnit:          types.FromDuration(1 * time.Second),
		WaitBeforeStart:   fsm.Tick(3),
		ChannelBufferSize: 4096,
		Options: fsm.Options{
			Name:                       "script",
			BufferSize:                 4096,
			IgnoreUndefinedTransitions: true,
			IgnoreUndefinedSignals:     true,
			IgnoreUndefinedStates:      true,
		},
	}

	// DefaultOptions is the default options of the controller. This can be controlled at starup
	// and is set once.
	DefaultOptions = script.Options{
		MinChannelBufferSize: 1024,
		ModelProperties:      DefaultModelProperties,
	}

	// DefaultProperties is the default properties for the controller, this is per collection / commit
	DefaultProperties = script.Properties{}
)

// Components contains a set of components in this controller.
type Components struct {
	Controllers func() (map[string]controller.Controller, error)
	Metadata    func() (map[string]metadata.Plugin, error)
	Events      event.Plugin
}

// NewComponents returns a controller implementation
func NewComponents(scope scope.Scope, options script.Options) *Components {

	controller := internal.NewController(
		// the constructor
		func(spec types.Spec) (internal.Managed, error) {
			return newPipeline(scope, options)
		},
		// the key function
		func(metadata types.Metadata) string {
			return metadata.Name
		},
	)

	return &Components{
		Controllers: controller.Controllers,
		Metadata:    controller.Metadata,
		Events:      controller,
	}
}
