package client // import "github.com/docker/infrakit/pkg/rpc/client"

import (
	"fmt"
	"testing"

	"github.com/docker/infrakit/pkg/spi"
	"github.com/stretchr/testify/require"
)

func TestErrNotSupported(t *testing.T) {

	iface1 := spi.InterfaceSpec{
		Name: "test",
		Sub:  "sub1",
	}
	iface2 := spi.InterfaceSpec{
		Name: "test",
		Sub:  "sub2",
	}

	err1 := errNotSupported(iface1)
	require.True(t, IsErrNotSupported(err1))
	require.True(t, IsInterfaceNotSupported(err1, iface1))
	require.False(t, IsInterfaceNotSupported(err1, iface2))
}

func TestErrVersionMismatch(t *testing.T) {
	var e error

	e = errVersionMismatch("test")
	require.True(t, IsErrVersionMismatch(e))

	e = fmt.Errorf("untyped")
	require.False(t, IsErrVersionMismatch(e))
}
