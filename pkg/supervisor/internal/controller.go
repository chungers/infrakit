package internal

import (
	"github.com/docker/infrakit/pkg/controller"
	"github.com/docker/infrakit/pkg/types"
)

type queuedController struct {
	Queued

	controller.Controller // the backend that does the real work

	findSpecsFunc  func(*types.Metadata) ([]types.Spec, error)
	updateSpecFunc func(types.Spec) error
	removeSpecFunc func(types.Spec) error
}

// QueuedController returns a controller.Controller that has a backing storage for specs and
// where all operations are serialized onto a work queue.
func QueuedController(c controller.Controller, ch chan<- backendOp,
	findSpecsFunc func(*types.Metadata) ([]types.Spec, error),
	updateSpecFunc func(types.Spec) error,
	removeSpecFunc func(types.Spec) error) controller.Controller {

	return &queuedController{
		Queued:         queue(ch),
		Controller:     c,
		findSpecsFunc:  findSpecsFunc,
		updateSpecFunc: updateSpecFunc,
		removeSpecFunc: removeSpecFunc,
	}
}

// Plan implements pkg/controller/Controller.Plan
func (q *queuedController) Plan(operation controller.Operation,
	spec types.Spec) (object types.Object, plan controller.Plan, err error) {
	result := q.Run(controller.Controller.Plan,
		func() []interface{} {
			r1, r2, r3 := q.Controller.Plan(operation, spec)
			return []interface{}{r1, r2, r3}
		})

	if v, is := result[0].(types.Object); is {
		object = v
	}
	if v, is := result[1].(controller.Plan); is {
		plan = v
	}
	if v, is := result[2].(error); is {
		err = v
	}
	return
}

// Commit implements pkg/controller/Controller.Commit
func (q *queuedController) Commit(operation controller.Operation, spec types.Spec) (object types.Object, err error) {
	result := q.Run(controller.Controller.Commit,
		func() []interface{} {

			switch operation {
			case controller.Enforce:
				if err := q.updateSpecFunc(spec); err != nil {
					return []interface{}{types.Object{}, err}
				}
			case controller.Destroy:
				if err := q.removeSpecFunc(spec); err != nil {
					return []interface{}{types.Object{}, err}
				}
			}

			r1, r2 := q.Controller.Commit(operation, spec)
			return []interface{}{r1, r2}
		})

	if v, is := result[0].(types.Object); is {
		object = v
	}
	if v, is := result[1].(error); is {
		err = v
	}
	return
}

// Describe implements pkg/controller/Controller.Describe
func (q *queuedController) Describe(metadata *types.Metadata) (objects []types.Object, err error) {

	search := metadata
	if metadata != nil {
		copy := *metadata
		search = &copy
	}

	result := q.Run(controller.Controller.Describe,
		func() []interface{} {
			r1, r2 := q.Controller.Describe(search)
			return []interface{}{r1, r2}
		})

	if v, is := result[0].([]types.Object); is {
		objects = v
	}
	if v, is := result[1].(error); is {
		err = v
	}
	return
}

// Specs implements pkg/controller/Controller.Specs
func (q *queuedController) Specs(metadata *types.Metadata) (specs []types.Spec, err error) {

	search := metadata
	if metadata != nil {
		copy := *metadata
		search = &copy
	}

	result := q.Run(controller.Controller.Plan,
		func() []interface{} {

			found, err := q.findSpecsFunc(search)
			if err != nil {
				return []interface{}{nil, err}
			}

			// Note we are not delegating the call to the backend. Instead, we are returning what's stored.
			return []interface{}{found, err}
		})

	if v, is := result[0].([]types.Spec); is {
		specs = v
	}
	if v, is := result[1].(error); is {
		err = v
	}
	return
}

// Pause implements pkg/controller/Controller.Pause
func (q *queuedController) Pause(metadata *types.Metadata) (specs []types.Object, err error) {

	search := metadata
	if metadata != nil {
		copy := *metadata
		search = &copy
	}

	result := q.Run(controller.Controller.Plan,
		func() []interface{} {

			// need to remove the desired specs from storage
			if found, err := q.findSpecsFunc(search); err == nil {
				for _, s := range found {
					// Note: this is not atomic.  Each individual remove is a commit.
					q.removeSpecFunc(s)
				}
			}

			// Note it's possible that the pause fails and we are out of sync with what is
			// in the storage.  The store operations commmit immediately.
			// It's more important to store the user's desired state accurately and then we can always reconcile later.
			r1, r2 := q.Controller.Pause(search)
			return []interface{}{r1, r2}
		})

	if v, is := result[0].([]types.Object); is {
		specs = v
	}
	if v, is := result[1].(error); is {
		err = v
	}
	return
}

// Terminate implements pkg/controller/Controller.Terminate
func (q *queuedController) Terminate(metadata *types.Metadata) (specs []types.Object, err error) {

	search := metadata
	if metadata != nil {
		copy := *metadata
		search = &copy
	}

	result := q.Run(controller.Controller.Plan,
		func() []interface{} {

			// need to remove the desired specs from storage
			if found, err := q.findSpecsFunc(search); err == nil {
				for _, s := range found {
					// Note: this is not atomic.  Each individual remove is a commit.
					q.removeSpecFunc(s)
				}
			}

			// Note it's possible that the terminate fails and we are out of sync with what is
			// in the storage.  The store operations commmit immediately.
			// It's more important to store the user's desired state accurately and then we can always reconcile later.
			r1, r2 := q.Controller.Terminate(search)
			return []interface{}{r1, r2}
		})

	if v, is := result[0].([]types.Object); is {
		specs = v
	}
	if v, is := result[1].(error); is {
		err = v
	}
	return
}
