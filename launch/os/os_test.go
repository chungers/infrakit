package os

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLaunchOSCommand(t *testing.T) {

	launcher, err := NewLauncher()
	require.NoError(t, err)

	starting, err := launcher.Launch("no-such-command")
	require.Error(t, err)
	require.Nil(t, starting)

	starting, err = launcher.Launch("sleep", "100")
	require.NoError(t, err)

	<-starting
	t.Log("started")
}
