package script

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	script "github.com/docker/infrakit/pkg/controller/script/types"
	"github.com/docker/infrakit/pkg/fsm"
	"github.com/docker/infrakit/pkg/run/scope"
	"github.com/docker/infrakit/pkg/template"
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

func (s Step) exec(ctx context.Context, output func(script.Result), b *batch, scope scope.Scope,
	targetParsers targetParsers, finish func(error)) {

	log.Debug("exec", "call", s.Call, "step", s, "targets", s.targets)

	targets := s.targets
	if len(targets) == 0 {

		// No target specified... just execute the callable as-is.  Otherwise, fork into N goroutines
		// and execute then collect the results.
		if s.Target == nil {

			key := fmt.Sprintf("%s/%s", b.spec.Metadata.Name, s.Call)
			sm := b.model.NewFork(s)

			// create an item for tracking
			b.Put(key, sm, b.model.Spec(), nil)

			// need to load *all* of the targets as specified in the 'targets' section:
			v, err := blockingExec(b, s.Step, scope, "")

			result := script.Result{Step: s.Step}
			if err == nil {
				sm.Signal(done)
				result.Output = v
			} else {
				log.Error("error", "step", s.Step, "key", key, "err", err)
				sm.Signal(fail)
				result.Error = err
			}

			output(result)

			// finish will signal to the caller the result of this execution
			finish(err)

			return

		} else {
			// load from the target section
			targets = targetParsers.targets(b.properties, *s.Target)
		}
	}

	results := make(chan script.Result, len(targets))
	// Need to 'fork' for each target.  The target list already takes into account parallelism.
	for _, target := range targets {

		ref := strings.Join(targets, ",")
		if s.Target != nil {
			ref = fmt.Sprintf("@%v", *s.Target)
		}

		key := fmt.Sprintf("%s/%s/%s/%s", b.spec.Metadata.Name, s.Call, ref, target)
		sm := b.model.NewFork(s)
		// create an item for tracking
		b.Put(key, sm, b.model.Spec(), nil)

		go func(t string) {
			v, err := blockingExec(b, s.Step, scope, t)

			result := script.Result{Step: s.Step, Target: target}
			if err == nil {
				sm.Signal(done)
				result.Output = v

			} else {
				log.Error("error", "step", s.Step, "key", key, "err", err)
				sm.Signal(fail)
				result.Error = err
			}

			results <- result

			output(result)

		}(target)
	}

	errors := errors{}
loop:
	for i := 0; i < len(targets); i++ {
		select {

		case <-ctx.Done():
			log.Info("Canceled", "step", s)
			break loop

		case result := <-results:
			if result.Error != nil {
				errors = append(errors, result.Error)
			}
		}
	}

	if len(errors) > 0 {
		log.Error("errors", "errors", errors.Error(), "step", s.Step)
		finish(errors)
	} else {
		log.Debug("no errors", "errors", errors.Error(), "step", s.Step)
		finish(nil)
	}
}

// blockingExec executes a single callable with a single target.  This is a blocking call.
func blockingExec(batch *batch, step script.Step, scope scope.Scope, target string) (interface{}, error) {

	call := batch.callables[step.Call]
	if call == nil {
		return nil, fmt.Errorf("no call: %v", step.Call)
	}

	// Set the parameters.  Very crude assumption that the types all match correctly:
	params := map[string]interface{}{}
	if step.Params != nil {
		err := step.Params.Decode(&params)
		if err != nil {
			return nil, err
		}
	}
	for k, v := range params {
		err := call.SetParameter(k, v)
		if err != nil {
			return nil, err
		}
	}

	// Setting additional parameters....  This assumes the callables implement common 'interfaces' to
	// receiving special parameters.  See the implementation of http and ssh for example.
	call.SetParameter("target", target)

	buffer := &bytes.Buffer{}
	args := []string{}

	ctx := context.Background()
	err := call.Execute(ctx, args, buffer)

	log.Debug("Ran exec", "step", step, "target", target, "call", call, "params", params, "err", err)

	if err != nil {
		log.Error("error", "err", err)
		return nil, err
	}

	if step.ResultTemplate != nil {
		return processResult(step, scope, target, buffer.String(), *step.ResultTemplate)
	}

	return buffer.String(), nil
}

func processResult(step script.Step, scope scope.Scope, target, result, expression string) (interface{}, error) {

	parts := []string{}
	for _, p := range strings.Split(expression, ";") {
		parts = append(parts, fmt.Sprintf("{{%v}}", p))
	}

	var processor *template.Template

	if len(parts) > 0 {

		expr := strings.Join(parts, "")

		t, err := scope.TemplateEngine("str://"+expr, template.Options{})
		if err != nil {
			log.Error("error", "err", err, "expr", expr)
			return nil, err
		}
		processor = t
	}

	if processor == nil {
		return result, nil
	}

	str, err := processor.Render(result)

	if err != nil {
		log.Error("error", "err", err, "parts", parts)
		return nil, err
	}

	if step.ResultTemplateVar != nil {
		return processor.Var(*step.ResultTemplateVar)
	}
	return str, nil
}
