package os

import (
	"os/exec"
	"sync"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/infrakit/launch"
)

// NewLauncher returns a Launcher that can install and start plugins.  The OS version is simple - it translates
// plugin names as command names and uses os.Exec
func NewLauncher() (launch.Launcher, error) {
	return &launcher{
		launching: make(map[string]<-chan error),
	}, nil
}

type launcher struct {
	launching map[string]<-chan error
	lock      sync.Mutex
}

// Launcn implements Launcher.Launch.  Returns a signal channel to block on optionally.
func (l *launcher) Launch(name string, args ...string) (<-chan error, error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if launching, has := l.launching[name]; has {
		return launching, nil
	}

	path, err := exec.LookPath(name)
	if err != nil {
		return nil, err
	}

	launching := make(chan error)
	l.launching[name] = launching

	go func() {

		cmd := exec.Command(name, args...)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
			Pgid:    0,
		}

		err := cmd.Start()
		log.Infoln("OS launcher: Plugin", name, "starting. from path=", path, ",err=", err)

		launching <- err
		close(launching)
	}()

	return launching, nil
}
