package launch

import (
	log "github.com/Sirupsen/logrus"
)

// Launcher is a service that is able to start plugins based on different
// mechanisms from running local binary to pulling and running docker containers or engine plugins
type Launcher interface {

	// Launch starts the plugin.  This can be an async process but the launcher will poll
	// for the running status of the plugin.  The client can receive and block on the returned channel
	// and add optional timeout in its own select statement.
	Launch(name string, args ...string) (<-chan error, error)
}

type noOp int

// NewNoOpLauncher doesn't actually launch the plugins.  It's a stub with no op and relies on manual plugin starts.
func NewNoOpLauncher() Launcher {
	return noOp(0)
}

// Launch starts the plugin given the name
func (n noOp) Launch(name string, args ...string) (<-chan error, error) {
	log.Infoln("NO-OP Launcher: not automatically starting plugin", name, "args=", args)

	starting := make(chan error)
	close(starting) // channel won't block
	return starting, nil
}
