package lifecycle

import (
	"context"
	"os"
	"sync"
)

type Process struct {
	startup       context.Context
	runtime       context.Context
	cancelStartup context.CancelFunc
	interrupts    chan os.Signal
	stopSource    func()
	stop          chan struct{}
	done          chan struct{}
	stopOnce      sync.Once
}

func LocalProcess(ctx context.Context) *Process {
	return &Process{startup: ctx, runtime: ctx}
}

func NotifyProcess(parent context.Context) *Process {
	source := NotifySignals()
	return newSignalProcess(parent, source.Channel(), source.Stop)
}

func newSignalProcess(
	parent context.Context,
	source <-chan os.Signal,
	stopSource func(),
) *Process {
	startup, cancel := context.WithCancel(parent)
	process := &Process{
		startup: startup, runtime: parent, cancelStartup: cancel,
		interrupts: make(chan os.Signal, 2), stopSource: stopSource,
		stop: make(chan struct{}), done: make(chan struct{}),
	}
	go process.forward(source)
	return process
}

func (process *Process) StartupContext() context.Context {
	if process == nil {
		return nil
	}
	return process.startup
}

func (process *Process) RuntimeContext() context.Context {
	if process == nil {
		return nil
	}
	return process.runtime
}

func (process *Process) Interrupts() <-chan os.Signal {
	if process == nil {
		return nil
	}
	return process.interrupts
}

func (process *Process) Stop() {
	if process == nil || process.done == nil {
		return
	}
	process.stopOnce.Do(func() {
		if process.stopSource != nil {
			process.stopSource()
		}
		close(process.stop)
		process.cancelStartup()
		<-process.done
	})
}

func (process *Process) forward(source <-chan os.Signal) {
	defer close(process.done)
	first := true
	for {
		select {
		case <-process.runtime.Done():
			process.cancelStartup()
			return
		case <-process.stop:
			return
		case signalValue, open := <-source:
			if !open {
				process.cancelStartup()
				return
			}
			if first {
				first = false
				process.cancelStartup()
			}
			select {
			case process.interrupts <- signalValue:
			case <-process.stop:
				return
			}
		}
	}
}
