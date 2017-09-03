package group

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/docker/infrakit/pkg/spi/instance"
	"github.com/stretchr/testify/require"
)

func TestLazyBlockAndCancel(t *testing.T) {

	called := make(chan struct{})
	g := LazyConnect(func() (Plugin, error) {
		close(called)
		return nil, fmt.Errorf("boom")
	}, 100*time.Second)

	errs := make(chan error, 1)
	go func() {
		_, err := g.DescribeGroup(ID("test"))
		errs <- err
		close(errs)
	}()

	<-called

	CancelWait(g)

	require.Equal(t, "cancelled", (<-errs).Error())
}

func TestLazyNoBlock(t *testing.T) {

	called := make(chan struct{})
	g := LazyConnect(func() (Plugin, error) {
		close(called)
		return nil, fmt.Errorf("boom")
	}, 0)

	errs := make(chan error, 1)
	go func() {
		_, err := g.DescribeGroup(ID("test"))
		errs <- err
		close(errs)
	}()

	<-called

	require.Equal(t, "boom", (<-errs).Error())
}

type fake chan []interface{}

func (f fake) CommitGroup(grp Spec, pretend bool) (string, error) {
	f <- []interface{}{Plugin.CommitGroup, grp, pretend}
	return "", nil
}

func (f fake) FreeGroup(id ID) error {
	f <- []interface{}{Plugin.FreeGroup, id}
	return nil
}

func (f fake) DescribeGroup(id ID) (Description, error) {
	f <- []interface{}{Plugin.DescribeGroup, id}
	return Description{}, nil
}

func (f fake) DestroyGroup(id ID) error {
	f <- []interface{}{Plugin.DestroyGroup, id}
	return nil
}

func (f fake) InspectGroups() ([]Spec, error) {
	f <- []interface{}{Plugin.InspectGroups}
	return nil, nil
}

func (f fake) DestroyInstances(id ID, sub []instance.ID) error {
	f <- []interface{}{Plugin.DestroyInstances, id, sub}
	return nil
}

func (f fake) Size(id ID) (int, error) {
	f <- []interface{}{Plugin.Size, id}
	return 100, nil
}

func (f fake) SetSize(id ID, sz int) error {
	f <- []interface{}{Plugin.SetSize, id, sz}
	return nil
}

func checkCalls(t *testing.T, ch chan []interface{}, args ...interface{}) {
	found := <-ch
	for i, a := range args {
		if reflect.ValueOf(a) != reflect.ValueOf(found[i]) {
			if a != found[i] {
				t.Fatal("Not equal:", found[i], "vs", a)
			}
		}
	}
}

func TestLazyNoBlockConnect(t *testing.T) {

	called := make(chan struct{})
	called2 := make(chan []interface{}, 2)

	g := LazyConnect(func() (Plugin, error) {
		close(called)
		return fake(called2), nil
	}, 0)

	errs := make(chan error, 1)
	go func() {
		_, err := g.DescribeGroup(ID("test"))
		errs <- err
		close(errs)

		g.Size(ID("test"))
		close(called2)
	}()

	<-called

	require.NoError(t, <-errs)
	checkCalls(t, called2, Plugin.DescribeGroup, ID("test"))
	checkCalls(t, called2, Plugin.Size, ID("test"))
}
