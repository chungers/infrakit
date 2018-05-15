package script

import (
	"context"
	"fmt"
	"strings"
	"time"

	script "github.com/docker/infrakit/pkg/controller/script/types"
	"github.com/docker/infrakit/pkg/fsm"
	"github.com/docker/infrakit/pkg/run/scope"
)

func (l targetParsers) targets(properties script.Properties, target string) []string {
	if blob, has := properties.Targets[target]; has {
		// Go through all the defined parsers and return whatever on first success.
		for _, parser := range l {
			if t, err := parser(blob); err == nil {
				return t
			}
		}
	}
	return nil
}

type shardsT []Step

// Step is the runtime object for an executable step
type Step struct {
	fsm.FSM
	script.Step           // The original spec
	start       fsm.Index // the state the step should start with
	state       fsm.Index // the current state of the step
	targets     []string
}

func computeShards(step Step, targets []string) shardsT {
	parallelism := 0
	if step.Parallelism != nil {
		parallelism = *step.Parallelism
	}

	// if parallelism is 0, it means we just execute a single step and give
	// the step the entire list of targets. So there are no shards.
	if parallelism == 0 {
		return []Step{} // no shards
	}

	if parallelism > len(targets) {
		parallelism = len(targets)
	}

	shards := shardsT{}
	for i := 0; i < len(targets); {
		right := i + parallelism
		if right >= len(targets) {
			right = len(targets)
		}
		shards = append(shards,
			Step{
				Step:    step.Step,
				start:   step.start,
				state:   step.state,
				targets: targets[i:right],
			})
		i = right
	}
	return shards
}

func (s *Step) updateTargets(properties script.Properties, parsers targetParsers) {
	if s.Target != nil {
		s.targets = parsers.targets(properties, *s.Target)
	}
	return
}

type errors []error

// Error returns an error string
func (e errors) Error() string {
	errs := []string{}
	for _, err := range e {
		errs = append(errs, err.Error())
	}
	return strings.Join(errs, ",")
}

func (s Step) exec(ctx context.Context, b *batch, scope scope.Scope, properties script.Properties, finish func(error)) {

	log.Debug("exec", "call", s.Call, "step", s, "targets", s.targets)

	if len(s.targets) == 0 {
		finish(blockingExec(s.Step, scope, properties, ""))
		return
	}

	results := make(chan error)
	// Need to 'fork' for each target.  The target list already takes into account parallelism.
	for _, target := range s.targets {

		key := fmt.Sprintf("%s/%s/%s/%s", b.spec.Metadata.Name, s.Call, strings.Join(s.targets, ","), target)
		fsm := b.model.NewFork(s)
		// create an item for tracking
		b.Put(key, fsm, b.model.Spec(), nil)

		fmt.Println("******************", key)

		go func() {
			err := blockingExec(s.Step, scope, properties, target)
			results <- err
			if err == nil {
				fsm.Signal(done)
			} else {
				fsm.Signal(fail)
			}
		}()
	}

	errors := errors{}
loop:
	for i := 0; i < len(s.targets); i++ {
		select {

		case <-ctx.Done():
			log.Info("Canceled", "step", s)
			break loop

		case err := <-results:
			if err != nil {
				errors = append(errors, err)
			}
		}
	}

	if len(errors) > 0 {
		finish(errors)
	} else {
		finish(nil)
	}
}

// blockingExec executes a single callable with a single target.  This is a blocking call.
func blockingExec(step script.Step, scope scope.Scope, properties script.Properties, target string) error {
	log.Info("running exec", "step", step, "target", target)
	time.Sleep(2 * time.Second)
	return nil
}
