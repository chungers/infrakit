package core

import (
	"testing"

	"github.com/docker/infrakit/pkg/plugin"
	"github.com/docker/infrakit/pkg/types"
	"github.com/stretchr/testify/require"
)

func testNames(t *testing.T, kind string, pn plugin.Name, instance string, spec string) {
	s := types.Spec{}
	require.NoError(t, types.AnyYAMLMust([]byte(spec)).Decode(&s))
	q := AsAddressable(&s)
	require.Equal(t, pn, q.Plugin())
	require.Equal(t, kind, q.Kind())
	require.Equal(t, instance, q.Instance())
}

func TestAddressable(t *testing.T) {
	a := NewAddressableFromMetadata("group", types.Metadata{Name: "workers"})
	require.Equal(t, "group", a.Kind())
	require.Equal(t, "group/workers", string(a.Plugin()))
	require.Equal(t, "workers", a.Instance())

}

func TestDerivePluginNames(t *testing.T) {
	testNames(t, "ingress", plugin.Name("ingress/lb1"), "lb1", `
kind: ingress
metadata:
  name: lb1
`)
	testNames(t, "ingress", plugin.Name("us-east/lb1"), "lb1", `
kind: ingress
metadata:
  name: us-east/lb1
`)
	testNames(t, "group", plugin.Name("group/workers"), "workers", `
kind: group
metadata:
  name: workers
`)
	testNames(t, "group", plugin.Name("group/workers"), "workers", `
kind: group
metadata:
  name: group/workers
`)
	testNames(t, "group", plugin.Name("us-east/workers"), "workers", `
kind: group
metadata:
  name: us-east/workers
`)
	testNames(t, "resource", plugin.Name("resource/vpc1"), "vpc1", `
kind: resource
metadata:
  name: vpc1
`)
	testNames(t, "resource", plugin.Name("us-east/vpc1"), "vpc1", `
kind: resource
metadata:
  name: us-east/vpc1
`)
	testNames(t, "simulator", plugin.Name("simulator/disk"), "disk1", `
kind: simulator/disk
metadata:
  name: disk1
`)
	testNames(t, "simulator", plugin.Name("us-east/disk"), "disk1", `
kind: simulator/disk
metadata:
  name: us-east/disk1
`)
	testNames(t, "aws", plugin.Name("aws/ec2-instance"), "host1", `
kind: aws/ec2-instance
metadata:
  name: host1
`)
	testNames(t, "aws", plugin.Name("us-east/ec2-instance"), "host1", `
kind: aws/ec2-instance
metadata:
  name: us-east/host1
`)
}
