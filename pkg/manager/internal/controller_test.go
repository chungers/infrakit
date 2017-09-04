package internal

import (
	"fmt"
	"testing"

	"github.com/docker/infrakit/pkg/controller"
	controller_test "github.com/docker/infrakit/pkg/testing/controller"
	"github.com/docker/infrakit/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestQueuedControllers(t *testing.T) {

	args := make(chan []interface{}, 50)

	expectSpec := types.Spec{
		Kind:       "group",
		Metadata:   types.Metadata{Name: "workers"},
		Properties: types.AnyValueMust("stuff"),
	}
	expectSpec2 := types.Spec{
		Kind:       "group",
		Metadata:   types.Metadata{Name: "managers"},
		Properties: types.AnyValueMust("stuff"),
	}
	expectObject := types.Object{Spec: expectSpec}
	expectPlan := controller.Plan{Message: []string{"create"}}

	f1 := &controller_test.Controller{}

	c1 := make(chan backendOp)
	go func() {
		for op := range c1 {
			err := op.operation()
			if err != nil {
				panic(err)
			}
		}
	}()

	var q1 controller.Controller = &queuedController{
		Queued:     queue(c1),
		Controller: f1,
	}

	/////////////////////////////////////
	// Test passing input values and multiple calls

	f1.DoPlan = func(operation controller.Operation,
		spec types.Spec) (object types.Object, plan controller.Plan, err error) {

		args <- []interface{}{operation, spec}

		object = expectObject
		plan = expectPlan

		return
	}

	actualObject, actualPlan, err := q1.Plan(controller.Enforce, expectSpec)
	require.NoError(t, err)
	require.Equal(t, expectObject, actualObject)
	require.Equal(t, expectPlan, actualPlan)

	q1.Plan(controller.Destroy, expectSpec)
	q1.Plan(controller.Enforce, expectSpec2)
	q1.Plan(controller.Destroy, expectSpec2)

	require.EqualValues(t, []interface{}{controller.Enforce, expectSpec}, <-args)
	require.EqualValues(t, []interface{}{controller.Destroy, expectSpec}, <-args)
	require.EqualValues(t, []interface{}{controller.Enforce, expectSpec2}, <-args)
	require.EqualValues(t, []interface{}{controller.Destroy, expectSpec2}, <-args)

	f1.DoCommit = func(operation controller.Operation, spec types.Spec) (object types.Object, err error) {

		args <- []interface{}{operation, spec}

		object = expectObject
		return
	}

	storeCalled := make(chan types.Spec, 10)
	q1.(*queuedController).updateSpecFunc = func(s types.Spec) error {
		storeCalled <- s
		return nil
	}
	q1.(*queuedController).removeSpecFunc = func(s types.Spec) error {
		storeCalled <- s
		return nil
	}

	actualObject, err = q1.Commit(controller.Enforce, expectSpec)
	require.NoError(t, err)
	require.Equal(t, expectPlan, actualPlan)

	q1.Commit(controller.Enforce, expectSpec2)
	q1.Commit(controller.Destroy, expectSpec)
	q1.Commit(controller.Destroy, expectSpec2)

	require.EqualValues(t, []interface{}{controller.Enforce, expectSpec}, <-args)
	require.EqualValues(t, []interface{}{controller.Enforce, expectSpec2}, <-args)
	require.EqualValues(t, []interface{}{controller.Destroy, expectSpec}, <-args)
	require.EqualValues(t, []interface{}{controller.Destroy, expectSpec2}, <-args)

	require.EqualValues(t, expectSpec, <-storeCalled)
	require.EqualValues(t, expectSpec2, <-storeCalled)
	require.EqualValues(t, expectSpec, <-storeCalled)
	require.EqualValues(t, expectSpec2, <-storeCalled)

	////////////
	f1.DoPlan = func(operation controller.Operation, spec types.Spec) (object types.Object, plan controller.Plan, err error) {
		object = expectObject
		plan = expectPlan
		err = fmt.Errorf("boom")
		return
	}
	a, b, err := q1.Plan(controller.Enforce, expectSpec)
	require.Error(t, err)
	require.Equal(t, expectObject, a)
	require.Equal(t, expectPlan, b)

	////////////
	f1.DoCommit = func(operation controller.Operation, spec types.Spec) (object types.Object, err error) {
		object = expectObject
		err = fmt.Errorf("boom")
		return
	}

	a, err = q1.Commit(controller.Enforce, expectSpec)
	require.Error(t, err)
	require.Equal(t, expectObject, a)

	////////////
	f1.DoCommit = func(operation controller.Operation, spec types.Spec) (object types.Object, err error) {
		object = expectObject
		err = fmt.Errorf("boom")
		return
	}

	a, err = q1.Commit(controller.Enforce, expectSpec)
	require.Error(t, err)
	require.Equal(t, expectObject, a)

	////////////
	expectObjects := []types.Object{expectObject}
	f1.DoDescribe = func(search *types.Metadata) (objects []types.Object, err error) {
		args <- []interface{}{search}
		objects = expectObjects
		err = fmt.Errorf("boom")
		return
	}

	actualObjects, err := q1.Describe(nil)
	require.Error(t, err)
	require.Equal(t, expectObjects, actualObjects)
	require.Equal(t, []interface{}{(*types.Metadata)(nil)}, <-args)

	actualObjects, err = q1.Describe(&expectSpec.Metadata)
	require.Error(t, err)
	require.Equal(t, expectObjects, actualObjects)
	require.Equal(t, expectSpec.Metadata, *(<-args)[0].(*types.Metadata))

	////////////
	searchCalled := make(chan *types.Metadata, 10)
	expectSpecs := []types.Spec{expectSpec}
	q1.(*queuedController).findSpecsFunc = func(s *types.Metadata) ([]types.Spec, error) {
		searchCalled <- s
		return expectSpecs, nil
	}
	f1.DoSpecs = func(search *types.Metadata) (specs []types.Spec, err error) {
		panic("shouldn't be called")
	}

	actualSpecs, err := q1.Specs(nil)
	require.NoError(t, err)
	require.Equal(t, expectSpecs, actualSpecs)
	require.Equal(t, (*types.Metadata)(nil), <-searchCalled)

	////////////
	q1.(*queuedController).findSpecsFunc = func(s *types.Metadata) ([]types.Spec, error) {
		args <- []interface{}{s, "find"}
		return expectSpecs, nil
	}
	q1.(*queuedController).removeSpecFunc = func(s types.Spec) error {
		args <- []interface{}{s, "remove"}
		return nil
	}
	f1.DoPause = func(search *types.Metadata) (objects []types.Object, err error) {
		args <- []interface{}{search, "pause"}
		objects = expectObjects
		return
	}

	actualObjects, err = q1.Pause(nil)
	require.NoError(t, err)
	require.Equal(t, expectObjects, actualObjects)
	require.Equal(t, []interface{}{(*types.Metadata)(nil), "find"}, <-args)
	require.Equal(t, []interface{}{expectSpec, "remove"}, <-args)
	require.Equal(t, []interface{}{(*types.Metadata)(nil), "pause"}, <-args)

	actualObjects, err = q1.Pause(&expectSpec.Metadata)
	require.NoError(t, err)
	require.Equal(t, expectObjects, actualObjects)
	require.Equal(t, expectSpec.Metadata, *(<-args)[0].(*types.Metadata))
	require.Equal(t, []interface{}{expectSpec, "remove"}, <-args)
	require.Equal(t, expectSpec.Metadata, *(<-args)[0].(*types.Metadata))

	f1.DoTerminate = func(search *types.Metadata) (objects []types.Object, err error) {
		args <- []interface{}{search, "terminate"}
		objects = expectObjects
		return
	}

	actualObjects, err = q1.Terminate(nil)
	require.NoError(t, err)
	require.Equal(t, expectObjects, actualObjects)
	require.Equal(t, []interface{}{(*types.Metadata)(nil), "find"}, <-args)
	require.Equal(t, []interface{}{expectSpec, "remove"}, <-args)
	require.Equal(t, []interface{}{(*types.Metadata)(nil), "terminate"}, <-args)

	actualObjects, err = q1.Terminate(&expectSpec.Metadata)
	require.NoError(t, err)
	require.Equal(t, expectObjects, actualObjects)
	require.Equal(t, expectSpec.Metadata, *(<-args)[0].(*types.Metadata))
	require.Equal(t, []interface{}{expectSpec, "remove"}, <-args)
	require.Equal(t, expectSpec.Metadata, *(<-args)[0].(*types.Metadata))

	close(c1)
	close(args)
	close(storeCalled)
	close(searchCalled)

}

func TestQueuedControllersMerged(t *testing.T) {

	args := make(chan []interface{}, 50)

	expectSpec := types.Spec{
		Kind:       "group",
		Metadata:   types.Metadata{Name: "workers"},
		Properties: types.AnyValueMust("stuff"),
	}
	expectSpec2 := types.Spec{
		Kind:       "group",
		Metadata:   types.Metadata{Name: "managers"},
		Properties: types.AnyValueMust("stuff"),
	}
	expectObject := types.Object{Spec: expectSpec}
	expectPlan := controller.Plan{Message: []string{"create"}}

	expectObject2 := types.Object{Spec: expectSpec2}
	expectPlan2 := controller.Plan{Message: []string{"create"}}

	f1 := &controller_test.Controller{
		DoPlan: func(operation controller.Operation, spec types.Spec) (object types.Object, plan controller.Plan, err error) {
			args <- []interface{}{operation, spec}

			object = expectObject
			plan = expectPlan

			return
		},
	}
	f2 := &controller_test.Controller{
		DoPlan: func(operation controller.Operation, spec types.Spec) (object types.Object, plan controller.Plan, err error) {
			args <- []interface{}{operation, spec}

			object = expectObject2
			plan = expectPlan2

			return
		},
	}

	c1 := make(chan backendOp)
	var q1 controller.Controller = &queuedController{
		Queued:     queue(c1),
		Controller: f1,
	}
	c2 := make(chan backendOp)
	var q2 controller.Controller = QueuedController(f2, c2, nil, nil, nil)

	done := make(chan struct{})
	main := merge(done, c1, c2)
	go func() {
		for op := range main {
			err := op.operation()
			if err != nil {
				panic(err)
			}
		}
	}()

	actualObject, actualPlan, err := q1.Plan(controller.Enforce, expectSpec)
	require.NoError(t, err)
	require.Equal(t, expectObject, actualObject)
	require.Equal(t, expectPlan, actualPlan)

	_, _, err = q2.Plan(controller.Destroy, expectSpec)

	actualObject2, actualPlan2, err := q2.Plan(controller.Enforce, expectSpec2)
	require.NoError(t, err)
	require.Equal(t, expectObject2, actualObject2)
	require.Equal(t, expectPlan2, actualPlan2)

	_, _, err = q2.Plan(controller.Destroy, expectSpec2)

	close(args)
	close(done)

	require.EqualValues(t, []interface{}{controller.Enforce, expectSpec}, <-args)
	require.EqualValues(t, []interface{}{controller.Destroy, expectSpec}, <-args)
	require.EqualValues(t, []interface{}{controller.Enforce, expectSpec2}, <-args)
	require.EqualValues(t, []interface{}{controller.Destroy, expectSpec2}, <-args)

}

func TestQueuedControllersMerged3(t *testing.T) {

	args := make(chan []interface{}, 50)

	expectSpec := types.Spec{
		Kind:       "group",
		Metadata:   types.Metadata{Name: "workers"},
		Properties: types.AnyValueMust("stuff"),
	}
	expectSpec2 := types.Spec{
		Kind:       "group",
		Metadata:   types.Metadata{Name: "managers"},
		Properties: types.AnyValueMust("stuff"),
	}
	expectSpec3 := types.Spec{
		Kind:       "group",
		Metadata:   types.Metadata{Name: "databases"},
		Properties: types.AnyValueMust("stuff"),
	}
	expectObject := types.Object{Spec: expectSpec}
	expectPlan := controller.Plan{Message: []string{"create"}}

	expectObject2 := types.Object{Spec: expectSpec2}
	expectPlan2 := controller.Plan{Message: []string{"create"}}

	expectObject3 := types.Object{Spec: expectSpec3}
	expectPlan3 := controller.Plan{Message: []string{"create"}}

	f1 := &controller_test.Controller{
		DoPlan: func(operation controller.Operation, spec types.Spec) (object types.Object, plan controller.Plan, err error) {
			args <- []interface{}{operation, spec}

			object = expectObject
			plan = expectPlan

			return
		},
	}
	f2 := &controller_test.Controller{
		DoPlan: func(operation controller.Operation, spec types.Spec) (object types.Object, plan controller.Plan, err error) {
			args <- []interface{}{operation, spec}

			object = expectObject2
			plan = expectPlan2

			return
		},
	}
	f3 := &controller_test.Controller{
		DoPlan: func(operation controller.Operation, spec types.Spec) (object types.Object, plan controller.Plan, err error) {
			args <- []interface{}{operation, spec}

			object = expectObject3
			plan = expectPlan3

			return
		},
	}

	c1 := make(chan backendOp)
	var q1 controller.Controller = QueuedController(f1, c1, nil, nil, nil)

	c2 := make(chan backendOp)
	var q2 controller.Controller = QueuedController(f2, c2, nil, nil, nil)

	c3 := make(chan backendOp)
	var q3 controller.Controller = QueuedController(f3, c3, nil, nil, nil)

	done := make(chan struct{})

	fan, main := newFanin(done)
	go func() {
		for op := range main {
			err := op.operation()
			if err != nil {
				panic(err)
			}
		}
	}()

	fan.add(c1)
	fan.add(c2)
	fan.add(c3)

	q1.Plan(controller.Enforce, expectSpec)
	q1.Plan(controller.Destroy, expectSpec)

	q2.Plan(controller.Enforce, expectSpec2)
	q2.Plan(controller.Destroy, expectSpec2)

	q3.Plan(controller.Enforce, expectSpec3)
	q3.Plan(controller.Destroy, expectSpec3)

	close(args)
	close(done)

	require.EqualValues(t, []interface{}{controller.Enforce, expectSpec}, <-args)
	require.EqualValues(t, []interface{}{controller.Destroy, expectSpec}, <-args)
	require.EqualValues(t, []interface{}{controller.Enforce, expectSpec2}, <-args)
	require.EqualValues(t, []interface{}{controller.Destroy, expectSpec2}, <-args)
	require.EqualValues(t, []interface{}{controller.Enforce, expectSpec3}, <-args)
	require.EqualValues(t, []interface{}{controller.Destroy, expectSpec3}, <-args)

}
