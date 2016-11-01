package os

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/infrakit/launch"
)

const (
	// LogDirEnvVar is the environment variable that may be used to customize the plugin logs location
	LogDirEnvVar = "INFRAKIT_LOG_DIR"
)

// DefaultLaderFile is the file that this detector uses to decide who the leader is.
// In a mult-host set up, it's assumed that the file system would be share (e.g. NFS mount or S3 FUSE etc.)
func DefaultLogDir() string {
	if logDir := os.Getenv(LogDirEnvVar); logDir != "" {
		return logDir
	}

	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return filepath.Join(usr.HomeDir, ".infrakit/logs")
}

// NewLauncher returns a Launcher that can install and start plugins.  The OS version is simple - it translates
// plugin names as command names and uses os.Exec
func NewLauncher(logDir string) (launch.Launcher, error) {
	return &launcher{
		logDir:    logDir,
		launching: make(map[string]<-chan error),
	}, nil
}

type launcher struct {
	logDir    string
	launching map[string]<-chan error
	lock      sync.Mutex
}

func (l *launcher) buildCmd(name string, args ...string) string {
	logging := "" //"--log 5"
	logfile := filepath.Join(l.logDir, fmt.Sprintf("%s-%d", name, time.Now().Unix()))
	return fmt.Sprintf("/bin/sh -c %s %s %s &>%s &", name, logging, strings.Join(args, " "), logfile)
}

// Launcn implements Launcher.Launch.  Returns a signal channel to block on optionally.
func (l *launcher) Launch(name string, args ...string) (<-chan error, error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if launching, has := l.launching[name]; has {
		return launching, nil
	}

	_, err := exec.LookPath(name)
	if err != nil {
		return nil, err
	}

	launching := make(chan error)
	l.launching[name] = launching

	go func() {

		sh := l.buildCmd(name, args...)
		parts := strings.Split(sh, " ")

		log.Infoln("OS launcher: Plugin", name, "starting", parts)

		cmd := exec.Command(parts[0], parts[1:]...)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
			Pgid:    0,
		}

		err = cmd.Start()

		if err != nil {
			log.Warningln("OS launcher: Plugin", name, "failed to start:", err, "cmd=", parts)
		}

		launching <- err
		close(launching)

	}()

	return launching, nil
}
