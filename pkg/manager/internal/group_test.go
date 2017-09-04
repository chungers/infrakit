package internal

import (
	"fmt"
	"testing"

	group_types "github.com/docker/infrakit/pkg/plugin/group/types"
	"github.com/docker/infrakit/pkg/spi/group"
	group_test "github.com/docker/infrakit/pkg/testing/group"
	"github.com/docker/infrakit/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestQueuedGroupPlugins(t *testing.T) {

	args := make(chan []interface{}, 50)

	expectSpec := group.Spec{
		ID: group.ID("workers"),
		Properties: types.AnyValueMust(group_types.Spec{
			Allocation: group_types.AllocationMethod{
				Size: 10,
			},
		}),
	}
	expectResp := "ok"

	f1 := &group_test.Plugin{}

	c1 := make(chan backendOp)
	go func() {
		for op := range c1 {
			err := op.operation()
			if err != nil {
				panic(err)
			}
		}
	}()

	var q1 group.Plugin = &queuedGroupPlugin{
		Queued: queue(c1),
		Plugin: f1,
	}

	/////////////////////////////////////
	// Test passing input values and multiple calls

	f1.DoCommitGroup = func(gspec group.Spec, pretend bool) (resp string, err error) {

		args <- []interface{}{gspec, pretend}

		resp = expectResp
		err = fmt.Errorf("boom")

		return
	}

	// commit - pretend ==> no update
	actualResp, err := q1.CommitGroup(expectSpec, true)
	require.Error(t, err)
	require.Equal(t, expectResp, actualResp)
	require.EqualValues(t, []interface{}{expectSpec, true}, <-args)

	// commit - not pretend ==> updates
	(q1.(*queuedGroupPlugin)).updateGroupSpecFunc = func(gspec group.Spec) error {
		args <- []interface{}{gspec, "update"}
		return nil
	}
	actualResp, err = q1.CommitGroup(expectSpec, false)
	require.Error(t, err)
	require.Equal(t, expectResp, actualResp)
	require.EqualValues(t, []interface{}{expectSpec, "update"}, <-args) // update happens before commit
	require.EqualValues(t, []interface{}{expectSpec, false}, <-args)

	// describe ==> gets state directory from plugin
	expectDesc := group.Description{Converged: true}
	f1.DoDescribeGroup = func(id group.ID) (desc group.Description, err error) {

		args <- []interface{}{id}

		desc = expectDesc
		return
	}

	expectID := group.ID("workers")
	actualDesc, err := q1.DescribeGroup(expectID)
	require.NoError(t, err)
	require.Equal(t, expectDesc, actualDesc)
	require.EqualValues(t, []interface{}{expectID}, <-args)

	// inspect ==> loads spec
	expectSpecs := []group.Spec{expectSpec}
	(q1.(*queuedGroupPlugin)).allGroupSpecsFunc = func() ([]group.Spec, error) {
		return expectSpecs, nil
	}
	f1.DoInspectGroups = func() (specs []group.Spec, err error) {
		panic("shouldn't be called") // we intercept and show stored version
	}

	actualSpecs, err := q1.InspectGroups()
	require.NoError(t, err)
	require.Equal(t, expectSpecs, actualSpecs)

	// free group
	(q1.(*queuedGroupPlugin)).findGroupSpecFunc = func(id group.ID) (group.Spec, error) {
		args <- []interface{}{id, "find"}
		return expectSpec, nil
	}
	(q1.(*queuedGroupPlugin)).removeGroupSpecFunc = func(id group.ID) error {
		args <- []interface{}{id, "remove"}
		return nil
	}
	f1.DoFreeGroup = func(id group.ID) (err error) {

		args <- []interface{}{id, "free"}

		return
	}

	err = q1.FreeGroup(expectID)
	require.NoError(t, err)
	require.EqualValues(t, []interface{}{expectID, "find"}, <-args)
	require.EqualValues(t, []interface{}{expectID, "remove"}, <-args)
	require.EqualValues(t, []interface{}{expectID, "free"}, <-args)

	// destroy group
	f1.DoDestroyGroup = func(id group.ID) (err error) {

		args <- []interface{}{id, "destroy"}
		return
	}
	err = q1.DestroyGroup(expectID)
	require.NoError(t, err)
	require.EqualValues(t, []interface{}{expectID, "find"}, <-args)
	require.EqualValues(t, []interface{}{expectID, "remove"}, <-args)
	require.EqualValues(t, []interface{}{expectID, "destroy"}, <-args)

	// set size
	f1.DoSetSize = func(id group.ID, size int) (err error) {
		args <- []interface{}{id, size, "set-size"}
		return
	}
	f1.DoCommitGroup = func(gspec group.Spec, pretend bool) (resp string, err error) {

		args <- []interface{}{gspec, pretend, "commit"}
		resp = expectResp

		return
	}
	err = q1.SetSize(expectID, 100)
	require.NoError(t, err)
	updatedSpec := group.Spec{
		ID: group.ID("workers"),
		Properties: types.AnyValueMust(group_types.Spec{
			Allocation: group_types.AllocationMethod{
				Size: 100,
			},
		}),
	}

	require.EqualValues(t, []interface{}{expectID, "find"}, <-args)
	require.EqualValues(t, []interface{}{updatedSpec, "update"}, <-args)
	require.EqualValues(t, []interface{}{updatedSpec, false, "commit"}, <-args) // commit call

	close(c1)
	close(args)

}
