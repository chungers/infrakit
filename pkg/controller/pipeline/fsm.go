package pipeline

import (
	"fmt"
	"sync"
	"time"

	script "github.com/docker/infrakit/pkg/controller/pipeline/types"
	"github.com/docker/infrakit/pkg/fsm"
)

const (

	// States
	starting fsm.Index = iota
	complete
	stopped
	terminated

	// user states begin at 100

	// Signals
	start fsm.Signal = iota
	exec
	done
	shardExec
	shardDone
	fail
	terminate
)

// Model encapsulates the workflow / state machines for provisioning resources
type Model struct {
	spec     *fsm.Spec
	set      *fsm.Set
	clock    *fsm.Clock
	tickSize time.Duration

	script.Properties
	script.Options

	shardExecChan  chan Step
	shardDoneChan  chan Step
	stepExecChan   chan Step
	stepDoneChan   chan Step
	stepErrChan    chan Step
	batchStartChan chan fsm.FSM
	batchDoneChan  chan fsm.FSM

	cleanupChan chan fsm.FSM

	lock sync.RWMutex
}

// NewBatch adds a new fsm in the start state
func (m *Model) NewBatch() fsm.FSM {
	return m.set.Add(starting)
}

// NewShard adds a new fsm to track a shard in a step
func (m *Model) NewShard(step Step) fsm.FSM {
	return m.set.Add(step.start) // assumes the waiting state is at first of the block of state codes
}

// NewThread is a fork within a shard.  For parallelism := N, there are N forks
func (m *Model) NewFork(step Step) fsm.FSM {
	return m.set.Add(fsm.Index(step.start + 10))
}

// BatchStart is the channel to get signals of entire batch starting
func (m *Model) BatchStart() <-chan fsm.FSM {
	return m.batchStartChan
}

// BatchDone is the channel to get signals of entire batch completing successfully.
func (m *Model) BatchDone() <-chan fsm.FSM {
	return m.batchDoneChan
}

// ShardExec is the channel to get signal to exec a shard of a step
func (m *Model) ShardExec() <-chan Step {
	return m.shardExecChan
}

// StepExec is the channel to get signal to exec a step
func (m *Model) StepExec() <-chan Step {
	return m.stepExecChan
}

// ShardDone is the channel to get signals of a shard of a single step completing successfully.
func (m *Model) ShardDone() <-chan Step {
	return m.shardDoneChan
}

// StepDone is the channel to get signals of a single step completing successfully.
func (m *Model) StepDone() <-chan Step {
	return m.stepDoneChan
}

// StepErr is the channel to get signals of a single step failed to execute.
func (m *Model) StepErr() <-chan Step {
	return m.stepErrChan
}

// Cleanup is the channel to get signals to clean up
func (m *Model) Cleanup() <-chan fsm.FSM {
	return m.cleanupChan
}

// Spec returns the model description
func (m *Model) Spec() *fsm.Spec {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return m.spec
}

// Start starts the model
func (m *Model) Start() {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.set == nil {
		m.clock.Start()

		log.Info("model starting", "options", m.Options.Options)
		m.set = fsm.NewSet(m.spec, m.clock, m.Options.Options)
	}
}

// Stop stops the model
func (m *Model) Stop() {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.set != nil {
		m.set.Stop()
		m.clock.Stop()

		close(m.batchStartChan)
		close(m.batchDoneChan)
		close(m.shardExecChan)
		close(m.shardDoneChan)
		close(m.stepExecChan)
		close(m.stepDoneChan)
		close(m.stepErrChan)
		close(m.cleanupChan)
		m.set = nil
	}
}

// BuildModel constructs a workflow model given the configuration blob provided by user in the Properties
func BuildModel(properties script.Properties, options script.Options) (*Model, error) {

	log.Info("Build model", "properties", properties, "options", options)
	model := &Model{
		Options:        options,
		Properties:     properties,
		shardExecChan:  make(chan Step, options.ChannelBufferSize),
		shardDoneChan:  make(chan Step, options.ChannelBufferSize),
		stepExecChan:   make(chan Step, options.ChannelBufferSize),
		stepDoneChan:   make(chan Step, options.ChannelBufferSize),
		stepErrChan:    make(chan Step, options.ChannelBufferSize),
		batchStartChan: make(chan fsm.FSM, options.ChannelBufferSize),
		batchDoneChan:  make(chan fsm.FSM, options.ChannelBufferSize),
		cleanupChan:    make(chan fsm.FSM, options.ChannelBufferSize),
		tickSize:       1 * time.Second,
	}

	model.clock = fsm.Wall(time.Tick(model.tickSize))

	stateOffset := 100
	// for n steps there are n states each corresponding to the 'running' of that step
	states := []fsm.State{
		{
			Index: starting,
			TTL:   fsm.Expiry{options.WaitBeforeStart, start},
			Transitions: map[fsm.Signal]fsm.Index{
				start:     fsm.Index(stateOffset),
				terminate: terminated,
			},
			Actions: map[fsm.Signal]fsm.Action{
				start: func(n fsm.FSM) error {
					model.batchStartChan <- n
					return nil
				},
			},
		},
		{
			Index: complete,
		},
		{
			Index: stopped,
		},
		{
			Index: terminated,
		},
	}

	stateNames := map[fsm.Index]string{
		starting: "START",
		complete: "COMPLETE",
		stopped:  "INCOMPLETE",
	}

	for i, ss := range properties.Steps {

		s := ss // make a copy so we can use it in callbacks

		block := (i + 1) * stateOffset
		wait := fsm.Index(block)
		run := fsm.Index(block + 1)
		completed := fsm.Index(block + 2)
		erred := fsm.Index(block + 3)
		fork := fsm.Index(block + 10)
		next := fsm.Index((i + 2) * stateOffset)

		if i == len(properties.Steps)-1 {
			next = complete
		}

		stateNames[wait] = fmt.Sprintf("WAITING[%s]", s.Call)
		stateNames[run] = fmt.Sprintf("RUNNING[%s]", s.Call)
		stateNames[completed] = fmt.Sprintf("DONE[%s]", s.Call)
		stateNames[fork] = fmt.Sprintf("FORK[%s]", s.Call)
		stateNames[erred] = fmt.Sprintf("ERROR[%s]", s.Call)

		// the 'waiting' state -- this is where a step/shard begins
		waiting := fsm.State{
			Index: wait,
			Transitions: map[fsm.Signal]fsm.Index{
				exec:      run,
				shardExec: run,
				terminate: terminated,
			},
			Actions: map[fsm.Signal]fsm.Action{
				shardExec: func(n fsm.FSM) error {
					model.shardExecChan <- Step{
						start: wait,
						state: wait,
						Step:  s,
						FSM:   n,
					}
					return nil
				},
				exec: func(n fsm.FSM) error {
					model.stepExecChan <- Step{
						start: wait,
						state: wait,
						Step:  s,
						FSM:   n,
					}
					return nil
				},
				terminate: func(n fsm.FSM) error {
					model.cleanupChan <- n
					return nil
				},
			},
		}

		// the 'running' state
		running := fsm.State{
			Index: run,
			Transitions: map[fsm.Signal]fsm.Index{
				exec:      run,
				done:      completed,
				shardDone: completed,
				fail:      erred,
				terminate: terminated,
			},
			Actions: map[fsm.Signal]fsm.Action{
				exec: func(n fsm.FSM) error {
					model.shardExecChan <- Step{
						start: wait,
						state: run,
						Step:  s,
						FSM:   n,
					}
					return nil
				},
				done: func(n fsm.FSM) error {
					model.stepDoneChan <- Step{
						start: wait,
						state: run,
						Step:  s,
						FSM:   n,
					}
					return nil
				},
				shardDone: func(n fsm.FSM) error {
					model.shardDoneChan <- Step{
						start: wait,
						state: run,
						Step:  s,
						FSM:   n,
					}
					return nil
				},
				fail: func(n fsm.FSM) error {
					model.stepErrChan <- Step{
						start: wait,
						state: run,
						Step:  s,
						FSM:   n,
					}
					return nil
				},
				terminate: func(n fsm.FSM) error {
					model.cleanupChan <- n
					return nil
				},
			},
		}

		completing := fsm.State{
			Index: completed,
			Transitions: map[fsm.Signal]fsm.Index{
				done:      next,
				terminate: terminated,
			},
			Actions: map[fsm.Signal]fsm.Action{
				done: func(n fsm.FSM) error {
					if next == complete {
						model.batchDoneChan <- n
						return nil
					}
					n.Signal(exec)
					return nil
				},
				terminate: func(n fsm.FSM) error {
					model.cleanupChan <- n
					return nil
				},
			},
		}

		// the 'error' state
		error := fsm.State{
			Index: erred,
			Transitions: map[fsm.Signal]fsm.Index{
				exec: run,
				done: stopped,
			},
		}

		// the 'forking' state
		forking := fsm.State{
			Index: fork,
			Transitions: map[fsm.Signal]fsm.Index{
				done: completed,
				fail: erred,
			},
		}

		if s.Retries != nil {
			wait := options.WaitBeforeStart
			if s.WaitBeforeRetry != nil {
				wait = *s.WaitBeforeRetry
			}
			error.TTL = fsm.Expiry{wait, exec}
			if s.Retries != nil {
				error.Visit = fsm.Limit{*s.Retries, done}
			}
		}
		states = append(states, waiting, running, completing, forking, error)
	}

	spec, err := fsm.Define(states[0], states[1:]...)
	if err != nil {
		return nil, err
	}

	spec.SetStateNames(stateNames).SetSignalNames(map[fsm.Signal]string{
		start:     "start",
		exec:      "exec",
		shardExec: "exec-shard",
		done:      "complete",
		shardDone: "complete-shard",
		fail:      "error",
	})
	model.spec = spec
	return model, nil
}
