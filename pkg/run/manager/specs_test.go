package manager

import (
	"testing"

	"github.com/docker/infrakit/pkg/plugin"
	"github.com/docker/infrakit/pkg/types"
	"github.com/stretchr/testify/require"
)

func testNames(t *testing.T, kind string, pn plugin.Name, spec string) {
	s := types.Spec{}
	require.NoError(t, types.AnyYAMLMust([]byte(spec)).Decode(&s))
	q := specQuery{s}
	require.Equal(t, pn, q.Plugin())
	require.Equal(t, kind, q.Kind())

}

func TestDerivePluginNames(t *testing.T) {
	testNames(t, "ingress", plugin.Name("ingress/lb1"), `
kind: ingress
metadata:
  name: lb1
`)
	testNames(t, "ingress", plugin.Name("us-east/lb1"), `
kind: ingress
metadata:
  name: us-east/lb1
`)
	testNames(t, "group", plugin.Name("group/workers"), `
kind: group
metadata:
  name: workers
`)
	testNames(t, "group", plugin.Name("group/workers"), `
kind: group
metadata:
  name: group/workers
`)
	testNames(t, "group", plugin.Name("us-east/workers"), `
kind: group
metadata:
  name: us-east/workers
`)
	testNames(t, "resource", plugin.Name("resource/vpc1"), `
kind: resource
metadata:
  name: vpc1
`)
	testNames(t, "resource", plugin.Name("us-east/vpc1"), `
kind: resource
metadata:
  name: us-east/vpc1
`)
	testNames(t, "simulator", plugin.Name("simulator/disk"), `
kind: simulator/disk
metadata:
  name: disk1
`)
	testNames(t, "simulator", plugin.Name("us-east/disk"), `
kind: simulator/disk
metadata:
  name: us-east/disk1
`)
	testNames(t, "aws", plugin.Name("aws/ec2-instance"), `
kind: aws/ec2-instance
metadata:
  name: host1
`)
	testNames(t, "aws", plugin.Name("us-east/ec2-instance"), `
kind: aws/ec2-instance
metadata:
  name: us-east/host1
`)
}
