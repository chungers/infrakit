package leader

import (
	"time"

	log "github.com/Sirupsen/logrus"
)

// Status indicates leadership status
type Status int

const (
	// StatusNotLeader means the current node is not a leader
	StatusNotLeader Status = iota

	// StatusLeader means the current node / instance is a leader
	StatusLeader

	// StatusUnkown indicates some exception happened while determining leadership.  Consumer will interpret accordingly.
	StatusUnknown
)

// Leadership is a struct that captures the leadership state, possibly error if exception occurs
type Leadership struct {
	Status Status
	Error  error
}

// Detector is the interface for determining whether this instance is a leader
type Detector interface {

	// Start starts leadership detection
	Start() (<-chan Leadership, error)

	// Stop stops
	Stop()
}

type poller struct {
	leaderChan   chan Leadership
	pollInterval time.Duration
	tick         <-chan time.Time
	stop         chan struct{}
	pollFunc     func() (bool, error)
}

// NewPoller returns a detector implementation given the poll interval and function that polls
func NewPoller(pollInterval time.Duration, f func() (bool, error)) Detector {
	return &poller{
		pollInterval: pollInterval,
		tick:         time.Tick(pollInterval),
		pollFunc:     f,
	}
}

// Start implements Detect.Start
func (l *poller) Start() (<-chan Leadership, error) {
	if l.leaderChan != nil {
		return l.leaderChan, nil
	}

	l.leaderChan = make(chan Leadership)
	l.stop = make(chan struct{})

	go l.poll()
	return l.leaderChan, nil
}

// Stop implements Detect.Stop
func (l *poller) Stop() {
	if l.stop != nil {
		close(l.stop)
	}
}

func (l *poller) poll() {
	for {
		select {

		case <-l.tick:

			isLeader, err := l.pollFunc()
			event := Leadership{}
			if err != nil {
				event.Status = StatusUnknown
				event.Error = err
			} else {
				if isLeader {
					event.Status = StatusLeader
				} else {
					event.Status = StatusNotLeader
				}
			}

			l.leaderChan <- event

		case <-l.stop:
			log.Infoln("Stopping leadership check")
			close(l.leaderChan)
			l.leaderChan = nil
			l.stop = nil
			return
		}
	}
}
