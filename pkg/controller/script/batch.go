package script

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/infrakit/pkg/callable"
	"github.com/docker/infrakit/pkg/callable/backend"
	"github.com/docker/infrakit/pkg/controller/internal"
	script "github.com/docker/infrakit/pkg/controller/script/types"
	"github.com/docker/infrakit/pkg/fsm"
	"github.com/docker/infrakit/pkg/run/scope"
	"github.com/docker/infrakit/pkg/spi/event"
	"github.com/docker/infrakit/pkg/spi/instance"
	"github.com/docker/infrakit/pkg/types"
	"github.com/imdario/mergo"
)

type batch struct {
	*internal.Collection

	scope scope.Scope

	spec types.Spec

	properties script.Properties
	options    script.Options

	model     *Model
	modules   map[string]*callable.Module
	callables map[string]*callable.Callable

	targetParsers targetParsers

	cancel func()
}

// targetParsers is a list of parsers that takes a blob *types.Any to a list
// of hosts/ targets
type targetParsers []func(*types.Any) ([]string, error)

var (
	// TopicStatus is the topic for batch status
	TopicStatus = types.PathFromString("status")

	// TopicResults is the topic for results
	TopicResults = types.PathFromString("result")

	// TopicErr is the topic for error
	TopicErr = types.PathFromString("error")

	defaultTargetParsers = targetParsers{

		// Simple parsing a string list
		func(any *types.Any) ([]string, error) {
			list := []string{}
			err := types.Decode(any.Bytes(), &list)
			return list, err
		},
	}
)

func newBatch(scope scope.Scope, options script.Options) (internal.Managed, error) {

	if err := mergo.Merge(&options, DefaultOptions); err != nil {
		return nil, err
	}

	if err := options.Validate(context.Background()); err != nil {
		return nil, err
	}

	base, err := internal.NewCollection(scope,
		TopicErr,
		TopicStatus,
	)
	if err != nil {
		return nil, err
	}
	b := &batch{
		scope:      scope,
		Collection: base,
		options:    options,
	}
	// set the behaviors
	base.StartFunc = b.run
	base.StopFunc = b.stop
	base.UpdateSpecFunc = b.updateSpec
	base.TerminateFunc = b.terminate

	return b, nil
}

func (b *batch) updateSpec(spec types.Spec, previous *types.Spec) (err error) {

	prev := spec
	if previous != nil {
		prev = *previous
	}

	log.Debug("updateSpec", "spec", spec, "prev", prev)

	// parse input, then select the model to use
	properties := script.Properties{}
	err = spec.Properties.Decode(&properties)
	if err != nil {
		return
	}

	prevProperties := script.Properties{}
	err = prev.Properties.Decode(&prevProperties)
	if err != nil {
		return
	}

	options := b.options // the plugin options at initialization are the defaults
	err = spec.Options.Decode(&options)
	if err != nil {
		return
	}

	ctx := context.Background()
	if err = properties.Validate(ctx); err != nil {
		return
	}

	if err = options.Validate(ctx); err != nil {
		return
	}

	// load all the modules
	modules := map[string]*callable.Module{}
	for _, use := range properties.Use {
		mod := &callable.Module{
			Scope:    b.scope,
			IndexURL: use.URL,
			ParametersFunc: func() backend.Parameters {
				return &callable.Parameters{}
			},
		}

		if err = mod.Load(); err != nil {
			return
		}

		modules[use.As] = mod
	}

	callables := map[string]*callable.Callable{}
	// verify each callable is properly referenced:
	for _, step := range properties.Steps {
		path := strings.Split(step.Call, ".")
		if len(path) < 2 {
			return fmt.Errorf("call not completely specified")
		}
		mod, has := modules[path[0]]
		if !has {
			return fmt.Errorf("no such module: %v", path[0])
		}
		c, err := mod.Find(path[1:])
		if err != nil {
			return err
		}

		callables[step.Call] = c
	}

	log.Debug("Begin processing", "properties", properties, "previous", prevProperties, "options", options, "V", debugV2)

	// build the fsm model
	var model *Model
	model, err = BuildModel(properties, options)
	if err != nil {
		return
	}

	b.modules = modules
	b.callables = callables
	b.model = model
	b.spec = spec
	b.properties = properties
	b.options = options

	log.Debug("Starting with state", "properties", b.properties, "V", debugV)
	return
}

func (b *batch) sendResult(result script.Result) {

	if result.Error != nil {
		result.ErrorMessage = result.Error.Error()
	}

	any, err := types.AnyValue(result)
	if err != nil {
		log.Error("Error", "err", err)
	}

	key := fmt.Sprintf("%s/%s", result.Step.Call, result.Target)
	if result.Target == "" {
		key = result.Step.Call
	}

	b.EventCh() <- event.Event{
		Topic:   b.Topic(TopicResults),
		Type:    event.Type("Result"),
		ID:      key,
		Message: "Result",
		Data:    any,
	}.Init()

	// update metadata too
	exported := result.Output
	if result.Error != nil {
		exported = result.Error.Error()
	}
	b.MetadataExportKV(key, exported)

	log.Info("Result", "call", result.Step.Call, "target", result.Target, "result", any.String(), "exported", exported)

}

func (b *batch) run(ctx context.Context) {

	// Start the model
	b.model.Start()

	// channels that aggregate from all the instance accessors
	type observation struct {
		instances []instance.Description
	}

	ctx, cancel := context.WithCancel(context.Background())
	b.cancel = cancel

	go func() {

		type shardInfo struct {
			shards shardsT
			parent *internal.Item
		}
		shardInfos := map[fsm.ID]*shardInfo{}

	loop:
		for {

			select {

			case <-ctx.Done():
				b.terminate()
				break loop

			case f, ok := <-b.model.BatchStart():
				if !ok {
					return
				}
				item := b.Collection.GetByFSM(f)
				if item != nil {
					b.EventCh() <- event.Event{
						Topic:   b.Topic(TopicStatus),
						Type:    event.Type("Start"),
						ID:      b.EventID(item.Key),
						Message: "Batch init",
					}.Init()
				}

				item.State.Signal(exec)

			case f, ok := <-b.model.BatchDone():
				if !ok {
					return
				}
				item := b.Collection.GetByFSM(f)
				if item != nil {
					b.EventCh() <- event.Event{
						Topic:   b.Topic(TopicStatus),
						Type:    event.Type("Completed"),
						ID:      b.EventID(item.Key),
						Message: "Batch completed.",
					}.Init()

					b.EventCh() <- event.Event{
						Topic:   b.Topic(TopicResults),
						Type:    event.Type("ResultEnd"),
						ID:      b.EventID(item.Key),
						Message: "ResultEnd",
					}.Init()
				}

			case s, ok := <-b.model.ShardExec():
				if !ok {
					return
				}
				item := b.Collection.GetByFSM(s.FSM)
				if item == nil {
					continue
				}

				log.Debug("shard-exec", "item", item, "shard", s)

				b.EventCh() <- event.Event{
					Topic:   b.Topic(TopicStatus),
					Type:    event.Type("ExecShard"),
					ID:      b.EventID(item.Key),
					Message: "Exec shard of step: " + s.Call,
				}.Init()

				curShard := s
				// find the shard... the shard stored in the shardInfos contain the computed targets
				info, has := shardInfos[s.ID()]
				if has {
					for _, shard := range info.shards {
						if s.ID() == shard.ID() {
							curShard = shard
						}
					}
				}

				ctx := context.Background()
				if b.options.StepDeadline > 0 {
					ctx, _ = context.WithDeadline(ctx, time.Now().Add(b.options.StepDeadline.Duration()))
				}

				curShard.exec(ctx, b.sendResult, b, b.scope, defaultTargetParsers,
					func(e error) {
						if e == nil {
							item.State.Signal(shardDone)
						} else {
							item.State.Signal(fail)
						}
					})

			case s, ok := <-b.model.ShardDone():
				if !ok {
					return
				}
				shard := b.Collection.GetByFSM(s.FSM)
				if shard != nil {

					log.Debug("shard-done", "shard", s)

					b.EventCh() <- event.Event{
						Topic:   b.Topic(TopicStatus),
						Type:    event.Type("DoneShard"),
						ID:      b.EventID(shard.Key),
						Message: "Done step shard: " + s.Call,
					}.Init()

					log.Debug("advance to next shard", "shard", s, "ID", s.ID(), "shardInfo", shardInfos)

					// Advance to the next shard:
					info, has := shardInfos[s.ID()]
					if has {
						if len(info.shards) > 0 {
							info.shards = info.shards[1:]
						}
						if len(info.shards) > 0 {
							// next shard starts
							info.shards[0].Signal(shardExec)
						} else {
							// this is the last shard
							info.parent.State.Signal(done)
						}
					} else {
						shard.State.Signal(done)
					}
				}

			case s, ok := <-b.model.StepExec():
				if !ok {
					return
				}

				item := b.Collection.GetByFSM(s.FSM)
				if item == nil {
					continue
				}

				b.EventCh() <- event.Event{
					Topic:   b.Topic(TopicStatus),
					Type:    event.Type("Exec"),
					ID:      b.EventID(item.Key),
					Message: "Exec step: " + s.Call,
				}.Init()

				curStep := s

				// Do we need to shard this step?
				shards := shardsT{}
				if curStep.Target != nil {
					shards = computeShards(curStep, defaultTargetParsers.targets(b.properties, *curStep.Target))
				}

				if len(shards) == 0 {
					item.State.Signal(exec)
					continue loop
				}

				shardInfos[item.State.ID()] = &shardInfo{
					shards: shards,
					parent: item,
				}
				for i, shard := range shards {

					key := fmt.Sprintf("%s/%s/%s", b.spec.Metadata.Name, shard.Call, strings.Join(shard.targets, ","))

					fsm := b.model.NewShard(shard)
					shards[i].FSM = fsm

					// create an item for tracking
					b.Put(key, fsm, b.model.Spec(), nil)

					shardInfos[fsm.ID()] = shardInfos[item.State.ID()]

					log.Debug("scheduled shard",
						"call", shard.Call, "key", key, "state", b.model.spec.StateName(fsm.State()))

					if i == 0 {
						fsm.Signal(shardExec)
					}
				}

			case s, ok := <-b.model.StepDone():
				if !ok {
					return
				}
				item := b.Collection.GetByFSM(s.FSM)
				if item != nil {

					log.Debug("step-done", "shard", s, "item", item, "state", b.model.spec.StateName(item.State.State()))

					step := s.Step

					b.EventCh() <- event.Event{
						Topic:   b.Topic(TopicStatus),
						Type:    event.Type("DoneStep"),
						ID:      b.EventID(item.Key),
						Message: "Done step: " + step.Call,
					}.Init()

					item.State.Signal(done) // continue to next
				}

			case s, ok := <-b.model.StepErr():
				if !ok {
					return
				}
				item := b.Collection.GetByFSM(s.FSM)
				if item != nil {

					step := s.Step

					b.EventCh() <- event.Event{
						Topic:   b.Topic(TopicStatus),
						Type:    event.Type("Error"),
						ID:      b.EventID(item.Key),
						Message: "Error in exec step: " + step.Call,
					}.Init()
				}

			case f, ok := <-b.model.Cleanup():
				if !ok {
					return
				}
				item := b.Collection.GetByFSM(f)
				if item != nil {
					b.Collection.Delete(item.Key)
				}
			}
		}
	}()

	log.Debug("Seeding instances", "steps", len(b.properties.Steps), "V", debugV)

	item := b.Put(b.spec.Metadata.Name, b.model.NewBatch(), b.model.Spec(), nil)

	log.Debug("requested", "id", item.State.ID(), "key", item.Key, "ordinal", item.Ordinal)

	log.Debug("Seeded instance. Running.")
}

func (b *batch) terminate() error {
	b.Visit(func(item internal.Item) bool {
		item.State.Signal(terminate)
		return true
	})

	return nil
}

func (b *batch) stop() error {
	log.Info("stop")

	if b.model != nil {

		b.cancel()

		log.Debug("Stopping", "V", debugV)
		b.model.Stop()
		b.model = nil
	}
	return nil
}
