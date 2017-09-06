package internal

import (
	"sync"
)

// Queued is something that can process work in a queue
type Queued interface {
	// Run queues the work and blocks until the work is executed with results
	Run(context interface{}, work func() []interface{}) (result []interface{})
}

type queue chan<- backendOp

// Run executes work on the queue without timeout.  This could hang indefinitely.  (TODO - add timeout)
func (q queue) Run(context interface{}, work func() []interface{}) []interface{} {
	ch := make(chan []interface{}, 1) // default is report is called once
	q <- backendOp{
		context: context,
		operation: func() error {
			ch <- work()
			close(ch)
			return nil
		},
	}
	return <-ch
}

func newFanin(done <-chan struct{}) (*fanin, <-chan backendOp) {
	out := make(chan backendOp)
	return &fanin{
		done: done,
		out:  out,
	}, out
}

type fanin struct {
	done <-chan struct{}
	out  chan<- backendOp
	lock sync.RWMutex
	wg   sync.WaitGroup
}

func (f *fanin) stop() {
	f.wg.Wait()
	close(f.out)
}

func (f *fanin) add(c <-chan backendOp) {
	f.lock.Lock()
	defer f.lock.Unlock()

	output := func(c <-chan backendOp) {
		defer f.wg.Done()
		for n := range c {
			select {
			case f.out <- n:
			case <-f.done:
				return
			}
		}
	}
	f.wg.Add(1)
	go output(c)
}

func merge(done <-chan struct{}, cs ...<-chan backendOp) <-chan backendOp {
	var wg sync.WaitGroup
	out := make(chan backendOp)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c or done is closed, then calls
	// wg.Done.
	output := func(c <-chan backendOp) {
		defer wg.Done()
		for n := range c {
			select {
			case out <- n:
			case <-done:
				return
			}
		}
	}

	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
