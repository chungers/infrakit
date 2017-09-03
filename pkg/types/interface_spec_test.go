package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInterfaceSpec(t *testing.T) {

	ingress := InterfaceSpec{
		Name:    "Controller",
		Version: "0.1.1",
		Sub:     "ingress",
	}

	require.Equal(t, "Controller/0.1.1/ingress", ingress.Encode())
	require.Equal(t, ingress, DecodeInterfaceSpec("Controller/0.1.1/ingress"))
	require.Equal(t, "Controller", (DecodeInterfaceSpec("Controller/0.1.1/ingress")).Name)
	require.Equal(t, "0.1.1", (DecodeInterfaceSpec("Controller/0.1.1/ingress")).Version)
	require.Equal(t, "ingress", (DecodeInterfaceSpec("Controller/0.1.1/ingress")).Sub)

	group := InterfaceSpec{
		Name:    "Group",
		Version: "0.1",
	}

	require.Equal(t, "Group/0.1", group.Encode())
	require.Equal(t, group, DecodeInterfaceSpec("Group/0.1"))
	require.Equal(t, "Group", (DecodeInterfaceSpec("Group/0.1")).Name)
	require.Equal(t, "0.1", (DecodeInterfaceSpec("Group/0.1")).Version)
	require.Equal(t, "", (DecodeInterfaceSpec("Group/0.1")).Sub)

}
