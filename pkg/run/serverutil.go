package run

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/Sirupsen/logrus"
	"github.com/docker/infrakit/pkg/discovery/local"
	"github.com/docker/infrakit/pkg/plugin"
	"github.com/docker/infrakit/pkg/rpc/server"
)

// EnsureDirExists makes sure the directory where the socket file will be placed exists.
func EnsureDirExists(dir string) {
	os.MkdirAll(dir, 0700)
}

// Plugin runs a plugin server, advertising with the provided name for discovery.
// The plugin should conform to the rpc call convention as implemented in the rpc package.
func Plugin(transport plugin.Transport, plugin server.VersionedInterface, more ...server.VersionedInterface) {
	_, running := BackgroundPlugin(transport, nil, plugin, more...)
	<-running
}

// BackgroundPlugin runs a plugin server, advertising with the provided name for discovery.
// The plugin should conform to the rpc call convention as implemented in the rpc package.
// This function does not block. Instead, use the returned channel to optionally block while the server is running.
// The provided callback, if not nil, will be called when the server stops, before the returned channel is closed.
func BackgroundPlugin(transport plugin.Transport, onStop func(), plugin server.VersionedInterface,
	more ...server.VersionedInterface) (server.Stoppable, <-chan struct{}) {

	dir := transport.Dir
	if dir == "" {
		dir = local.Dir()
	}
	EnsureDirExists(dir)
	name, _ := transport.Name.GetLookupAndType()

	socketPath := path.Join(dir, name)
	pidPath := path.Join(dir, name+".pid")
	return run(nil, socketPath, pidPath, onStop, plugin, more...)
}

// Listener runs a plugin server, listening at listen address, and
// advertising with the provided name for discovery.
// The plugin should conform to the rpc call convention as implemented in the rpc package.
func Listener(transport plugin.Transport, onStop func(), plugin server.VersionedInterface,
	more ...server.VersionedInterface) {

	_, running := BackgroundListener(transport, onStop, plugin, more...)
	<-running
}

// BackgroundListener runs a plugin server, listening at listen address, and
// advertising with the provided name for discovery.
// The plugin should conform to the rpc call convention as implemented in the rpc package.
// This function does not block. Use the returned channel to optionally block while the server is running.
func BackgroundListener(transport plugin.Transport, onStop func(), plugin server.VersionedInterface,
	more ...server.VersionedInterface) (server.Stoppable, <-chan struct{}) {

	dir := transport.Dir
	if dir == "" {
		dir = local.Dir()
	}
	EnsureDirExists(dir)
	name, _ := transport.Name.GetLookupAndType()

	discoverPath := path.Join(dir, name+".listen")
	pidPath := path.Join(dir, name+".pid")
	return run([]string{transport.Listen, transport.Advertise}, discoverPath, pidPath, onStop, plugin, more...)
}

func run(listen []string, discoverPath, pidPath string, onStop func(),
	plugin server.VersionedInterface, more ...server.VersionedInterface) (server.Stoppable, <-chan struct{}) {

	var stoppable server.Stoppable

	if len(listen) > 0 {
		s, err := server.StartListenerAtPath(listen, discoverPath, plugin, more...)
		if err != nil {
			logrus.Error(err)
		}
		stoppable = s
	} else {
		s, err := server.StartPluginAtPath(discoverPath, plugin, more...)
		if err != nil {
			logrus.Error(err)
		}
		stoppable = s
	}

	running := make(chan struct{})

	go func() {
		// write PID file
		err := ioutil.WriteFile(pidPath, []byte(fmt.Sprintf("%v", os.Getpid())), 0644)
		if err != nil {
			logrus.Error(err)
		}
		logrus.Infoln("PID file at", pidPath)
		if stoppable != nil {
			logrus.Infoln("Server waiting at", discoverPath)
			stoppable.AwaitStopped()
		}

		// clean up
		os.Remove(pidPath)
		logrus.Infoln("Removed PID file at", pidPath)

		os.Remove(discoverPath)
		logrus.Infoln("Removed discover file at", discoverPath)

		if onStop != nil {
			onStop()
		}

		close(running)
	}()

	return stoppable, running
}
