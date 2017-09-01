package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/docker/infrakit/pkg/controller"
	"github.com/docker/infrakit/pkg/discovery/local"
	"github.com/docker/infrakit/pkg/leader"
	controller_mock "github.com/docker/infrakit/pkg/mock/controller"
	group_mock "github.com/docker/infrakit/pkg/mock/spi/group"
	store_mock "github.com/docker/infrakit/pkg/mock/store"
	"github.com/docker/infrakit/pkg/plugin"
	controller_rpc "github.com/docker/infrakit/pkg/rpc/controller"
	group_rpc "github.com/docker/infrakit/pkg/rpc/group"
	"github.com/docker/infrakit/pkg/rpc/server"
	"github.com/docker/infrakit/pkg/spi/group"
	"github.com/docker/infrakit/pkg/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type testLeaderDetector struct {
	t     *testing.T
	me    string
	input <-chan string
	stop  chan struct{}
	ch    chan leader.Leadership
}

func (l *testLeaderDetector) Receive() <-chan leader.Leadership {
	return l.ch
}

func (l *testLeaderDetector) Start() (<-chan leader.Leadership, error) {
	l.stop = make(chan struct{})
	l.ch = make(chan leader.Leadership)
	go func() {
		for {
			select {
			case <-l.stop:
				return
			case found := <-l.input:
				if found == l.me {
					l.ch <- leader.Leadership{Status: leader.Leader}
				} else {
					l.ch <- leader.Leadership{Status: leader.NotLeader}
				}
			}
		}
	}()
	return l.ch, nil
}

func (l *testLeaderDetector) Stop() {
	close(l.stop)
}

func testEnsemble(t *testing.T,
	dir, id string,
	leader chan string,
	ctrl *gomock.Controller,
	configStore func(*store_mock.MockSnapshot),
	configureGroup func(*group_mock.MockPlugin),
	configureController func(*controller_mock.MockController)) (Backend, server.Stoppable) {

	disc, err := local.NewPluginDiscoveryWithDir(dir)
	require.NoError(t, err)

	detector := &testLeaderDetector{t: t, me: id, input: leader}

	snap := store_mock.NewMockSnapshot(ctrl)
	configStore(snap)

	// start group
	gm := group_mock.NewMockPlugin(ctrl)
	configureGroup(gm)

	gs := group_rpc.PluginServer(gm)
	st, err := server.StartPluginAtPath(filepath.Join(dir, "group-stateless"), gs)
	require.NoError(t, err)

	// start controller
	cm := controller_mock.NewMockController(ctrl)
	configureController(cm)

	cs := controller_rpc.Server(cm)
	st2, err := server.StartPluginAtPath(filepath.Join(dir, "ingress"), cs)
	require.NoError(t, err)

	m := NewManager(disc, detector, nil, snap, "group-stateless")
	ms := controller_rpc.ServerWithNamed(m.Controllers)
	mt, err := server.StartPluginAtPath(filepath.Join(dir, "group"), ms)
	require.NoError(t, err)

	return m, stoppables{st, st2, mt}
}

type stoppables []server.Stoppable

func (s stoppables) Stop() {
	for _, ss := range s {
		ss.Stop()
	}
}
func (s stoppables) AwaitStopped()         {}
func (s stoppables) Wait() <-chan struct{} { return nil }

func testSetLeader(t *testing.T, c []chan string, l string) {
	for _, cc := range c {
		cc <- l
	}
}

func testDiscoveryDir(t *testing.T) string {
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("%v", time.Now().UnixNano()))
	err := os.MkdirAll(dir, 0744)
	require.NoError(t, err)
	return dir
}

func testToStruct(m *types.Any) interface{} {
	o := map[string]interface{}{}
	m.Decode(&o)
	return &o
}

func testCloseAll(c []chan string) {
	for _, cc := range c {
		close(cc)
	}
}

func TestNoCallsToGroupWhenNoLeader(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	leaderChans := []chan string{make(chan string), make(chan string)}

	manager1, stoppable1 := testEnsemble(t, testDiscoveryDir(t), "m1", leaderChans[0], ctrl,
		func(s *store_mock.MockSnapshot) {
			// no calls
		},
		func(g *group_mock.MockPlugin) {
			// no calls
		},
		func(g *controller_mock.MockController) {
			// no calls
		},
	)
	manager2, stoppable2 := testEnsemble(t, testDiscoveryDir(t), "m2", leaderChans[1], ctrl,
		func(s *store_mock.MockSnapshot) {
			// no calls
		},
		func(g *group_mock.MockPlugin) {
			// no calls
		},
		func(g *controller_mock.MockController) {
			// no calls
		},
	)

	manager1.Start()
	manager2.Start()

	testSetLeader(t, leaderChans, "nobody")

	manager1.Stop()
	manager2.Stop()

	stoppable1.Stop()
	stoppable2.Stop()

	testCloseAll(leaderChans)
}

func TestStartOneLeader(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	global := globalSpec{}
	managerSpec := group.Spec{
		ID:         group.ID("managers"),
		Properties: types.AnyValueMust("hello"),
	}
	global.updateGroupSpec(managerSpec, plugin.Name("group-stateless"))

	// add an ingress spec
	ingressSpec := types.Spec{
		Kind: "ingress",
		Metadata: types.Metadata{
			Name: "elb1",
		},
		Properties: types.AnyValueMust(map[string]interface{}{"a": 1, "b": 2}),
	}
	global.updateSpec(ingressSpec, plugin.Name("ingress"))

	err := global.store(fakeSnapshot{
		SaveFunc: func(v interface{}) error {
			return nil
		},
	})
	require.NoError(t, err)

	leaderChans := []chan string{make(chan string), make(chan string)}

	wg := sync.WaitGroup{}
	wg.Add(2)

	manager1, stoppable1 := testEnsemble(t, testDiscoveryDir(t), "m1", leaderChans[0], ctrl,
		func(s *store_mock.MockSnapshot) {
			empty := &[]persisted{}
			s.EXPECT().Load(gomock.Eq(empty)).Do(
				func(o interface{}) error {
					p, is := o.(*[]persisted)
					require.True(t, is)
					*p = global.data
					return nil
				}).Return(nil)
		},
		func(g *group_mock.MockPlugin) {
			g.EXPECT().CommitGroup(gomock.Any(), false).Do(
				func(spec group.Spec, pretend bool) (string, error) {

					defer wg.Done()

					require.Equal(t, managerSpec.ID, spec.ID)
					require.Equal(t, testToStruct(managerSpec.Properties), testToStruct(spec.Properties))
					return "ok", nil
				}).Return("ok", nil)
		},
		func(g *controller_mock.MockController) {
			g.EXPECT().Commit(gomock.Any(), gomock.Any()).Do(
				func(op controller.Operation, spec types.Spec) (types.Object, error) {

					defer wg.Done()

					require.EqualValues(t, controller.Enforce, op)
					require.EqualValues(t, types.AnyValueMust(ingressSpec), types.AnyValueMust(spec))
					return types.Object{Spec: spec}, nil
				}).Return(types.Object{Spec: ingressSpec}, nil)
		},
	)
	manager2, stoppable2 := testEnsemble(t, testDiscoveryDir(t), "m2", leaderChans[1], ctrl,
		func(s *store_mock.MockSnapshot) {
			// no calls expected
		},
		func(g *group_mock.MockPlugin) {
			// no calls expected
		},
		func(g *controller_mock.MockController) {
			// no calls
		},
	)

	manager1.Start()
	manager2.Start()

	testSetLeader(t, leaderChans, "m1")

	wg.Wait()

	manager1.Stop()
	manager2.Stop()

	stoppable1.Stop()
	stoppable2.Stop()

	testCloseAll(leaderChans)
}

func TestChangeLeadership(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	global := globalSpec{}
	managerSpec := group.Spec{
		ID:         group.ID("managers"),
		Properties: types.AnyValueMust("hello"),
	}
	global.updateGroupSpec(managerSpec, plugin.Name("group-stateless"))

	// add an ingress spec
	ingressSpec := types.Spec{
		Kind: "ingress",
		Metadata: types.Metadata{
			Name: "elb1",
		},
		Properties: types.AnyValueMust(map[string]interface{}{"a": 1, "b": 2}),
	}
	global.updateSpec(ingressSpec, plugin.Name("ingress"))

	err := global.store(fakeSnapshot{
		SaveFunc: func(v interface{}) error {
			return nil
		},
	})
	require.NoError(t, err)

	leaderChans := []chan string{make(chan string), make(chan string)}

	checkpoint1 := make(chan struct{})
	checkpoint1b := make(chan struct{})
	checkpoint2 := make(chan struct{})
	checkpoint2b := make(chan struct{})
	checkpoint3 := make(chan struct{})
	checkpoint3b := make(chan struct{})

	manager1, stoppable1 := testEnsemble(t, testDiscoveryDir(t), "m1", leaderChans[0], ctrl,
		func(s *store_mock.MockSnapshot) {
			empty := &[]persisted{}
			s.EXPECT().Load(gomock.Eq(empty)).Do(
				func(o interface{}) error {
					p, is := o.(*[]persisted)
					require.True(t, is)
					*p = global.data
					return nil
				},
			).Return(nil)
		},
		func(g *group_mock.MockPlugin) {
			g.EXPECT().CommitGroup(gomock.Any(), false).Do(
				func(spec group.Spec, pretend bool) (string, error) {

					defer close(checkpoint1)

					require.Equal(t, managerSpec.ID, spec.ID)
					require.Equal(t, testToStruct(managerSpec.Properties), testToStruct(spec.Properties))
					return "ok", nil
				},
			).Return("ok", nil)

			// We will get a call to inspect what's being watched
			g.EXPECT().InspectGroups().AnyTimes().Return([]group.Spec{managerSpec}, nil)

			// Now we lost leadership... need to unwatch
			g.EXPECT().FreeGroup(gomock.Eq(group.ID("managers"))).Do(
				func(id group.ID) error {

					defer close(checkpoint3)

					return nil
				},
			).Return(nil)
		},
		func(g *controller_mock.MockController) {
			g.EXPECT().Commit(gomock.Any(), gomock.Any()).Do(
				func(op controller.Operation, spec types.Spec) (types.Object, error) {

					defer close(checkpoint1b)

					require.EqualValues(t, controller.Enforce, op)
					require.EqualValues(t, types.AnyValueMust(ingressSpec), types.AnyValueMust(spec))
					return types.Object{Spec: spec}, nil
				}).Return(types.Object{Spec: ingressSpec}, nil)

			// There should be a FREE here later...

			// We will get a call to inspect what's being watched
			g.EXPECT().Describe(gomock.Eq(nil)).AnyTimes().Return([]types.Object{{Spec: ingressSpec}}, nil)

			// Now we lost leadership... need to unwatch
			g.EXPECT().Free(gomock.Eq(&types.Metadata{Name: "elb1"})).Do(
				func(metdata *types.Metadata) ([]types.Object, error) {

					defer close(checkpoint3b)

					return []types.Object{{Spec: ingressSpec}}, nil
				},
			).Return([]types.Object{{Spec: ingressSpec}}, nil)
		},
	)
	manager2, stoppable2 := testEnsemble(t, testDiscoveryDir(t), "m2", leaderChans[1], ctrl,
		func(s *store_mock.MockSnapshot) {
			empty := &[]persisted{}
			s.EXPECT().Load(gomock.Eq(empty)).Do(
				func(o interface{}) error {
					p, is := o.(*[]persisted)
					require.True(t, is)
					*p = global.data
					return nil
				},
			).Return(nil)
		},
		func(g *group_mock.MockPlugin) {
			g.EXPECT().CommitGroup(gomock.Any(), false).Do(
				func(spec group.Spec, pretend bool) (string, error) {

					defer close(checkpoint2)

					require.Equal(t, managerSpec.ID, spec.ID)
					require.Equal(t, testToStruct(managerSpec.Properties), testToStruct(spec.Properties))
					return "ok", nil
				},
			).Return("ok", nil)
		},
		func(g *controller_mock.MockController) {
			g.EXPECT().Commit(gomock.Any(), gomock.Any()).Do(
				func(op controller.Operation, spec types.Spec) (types.Object, error) {

					defer close(checkpoint2b)

					require.EqualValues(t, controller.Enforce, op)
					require.EqualValues(t, types.AnyValueMust(ingressSpec), types.AnyValueMust(spec))
					return types.Object{Spec: spec}, nil
				}).Return(types.Object{Spec: ingressSpec}, nil)
		},
	)

	manager1.Start()
	manager2.Start()

	testSetLeader(t, leaderChans, "m1")

	<-checkpoint1
	<-checkpoint1b

	testSetLeader(t, leaderChans, "m2")

	<-checkpoint2
	<-checkpoint2b
	time.Sleep(5 * time.Second)
	//	<-checkpoint3
	//	<-checkpoint3b

	manager1.Stop()
	manager2.Stop()

	stoppable1.Stop()
	stoppable2.Stop()

	testCloseAll(leaderChans)
}
