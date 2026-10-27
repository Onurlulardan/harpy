package dashboardexport

import (
	"context"
	"fmt"
	"sync"
)

type RenderDispatcher struct {
	clock     Clock
	ids       IDSource
	claims    RenderClaimStore
	spools    SpoolAllocator
	executor  RenderAttemptExecutor
	admission DispatcherAdmission

	mutex         sync.Mutex
	started       bool
	startupReaped bool
	scanCancel    context.CancelFunc
	forceCancel   context.CancelFunc
	done          chan struct{}
	failures      chan error
}

func NewRenderDispatcher(
	clock Clock, ids IDSource, claims RenderClaimStore, spools SpoolAllocator,
	executor RenderAttemptExecutor, admission DispatcherAdmission,
) (*RenderDispatcher, error) {
	for _, dependency := range []any{clock, ids, claims, spools, executor, admission} {
		if nilPort(dependency) {
			return nil, fmt.Errorf("analytics dashboard export dispatcher dependency is unavailable")
		}
	}
	return &RenderDispatcher{clock: clock, ids: ids, claims: claims, spools: spools,
		executor: executor, admission: admission, done: make(chan struct{}),
		failures: make(chan error, 1)}, nil
}

func (dispatcher *RenderDispatcher) Start(parent context.Context) error {
	if parent == nil || parent.Err() != nil {
		return ErrInvalid
	}
	dispatcher.mutex.Lock()
	defer dispatcher.mutex.Unlock()
	if dispatcher.started {
		return fmt.Errorf("analytics dashboard export dispatcher is already started")
	}
	if dispatcher.admission.Accepting() {
		if err := dispatcher.spools.Reap(parent, dispatcher.claims,
			dispatcher.clock.Now()); err != nil {
			return fmt.Errorf("reap Analytics dashboard export spool at startup: %w", err)
		}
		dispatcher.startupReaped = true
	}
	scan, scanCancel := context.WithCancel(parent)
	force, forceCancel := context.WithCancel(parent)
	dispatcher.scanCancel, dispatcher.forceCancel = scanCancel, forceCancel
	dispatcher.started = true
	go dispatcher.serve(scan, force)
	return nil
}

func (dispatcher *RenderDispatcher) serve(scan, force context.Context) {
	err := dispatcher.run(scan, force)
	if err != nil {
		select {
		case dispatcher.failures <- err:
		default:
		}
	}
	close(dispatcher.failures)
	close(dispatcher.done)
}

func (dispatcher *RenderDispatcher) Failures() <-chan error { return dispatcher.failures }

func (dispatcher *RenderDispatcher) BeginDrain() {
	dispatcher.mutex.Lock()
	defer dispatcher.mutex.Unlock()
	if dispatcher.scanCancel != nil {
		dispatcher.scanCancel()
	}
}

func (dispatcher *RenderDispatcher) Force() error {
	dispatcher.mutex.Lock()
	defer dispatcher.mutex.Unlock()
	if dispatcher.scanCancel != nil {
		dispatcher.scanCancel()
	}
	if dispatcher.forceCancel != nil {
		dispatcher.forceCancel()
	}
	return nil
}

func (dispatcher *RenderDispatcher) Wait(ctx context.Context) error {
	if ctx == nil {
		return ErrInvalid
	}
	select {
	case <-dispatcher.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
