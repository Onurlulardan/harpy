package postgres

import (
	"context"
	"sync"
	"sync/atomic"
)

type boundOperation struct {
	state      *bindingState
	context    context.Context
	cancel     context.CancelFunc
	gateHeld   atomic.Bool
	finishOnce sync.Once
	abortMutex sync.Mutex
	abort      func()
	abortAsked bool
	abortRun   bool
	stopAbort  func() bool
}

func newBoundOperation(ctx context.Context, state *bindingState) *boundOperation {
	operationContext, cancel := context.WithCancel(ctx)
	operation := &boundOperation{
		state: state, context: operationContext, cancel: cancel,
	}
	operation.stopAbort = context.AfterFunc(operationContext, operation.requestAbort)
	return operation
}

func (operation *boundOperation) installAbort(abort func()) {
	operation.abortMutex.Lock()
	operation.abort = abort
	start := operation.abortAsked && !operation.abortRun
	if start {
		operation.abortRun = true
	}
	operation.abortMutex.Unlock()
	if start {
		go abort()
	}
}

func (operation *boundOperation) requestAbort() {
	operation.cancel()
	operation.abortMutex.Lock()
	operation.abortAsked = true
	abort := operation.abort
	start := abort != nil && !operation.abortRun
	if start {
		operation.abortRun = true
	}
	operation.abortMutex.Unlock()
	if start {
		go abort()
	}
}

func (operation *boundOperation) finish() {
	operation.finishOnce.Do(func() {
		if operation.stopAbort != nil {
			operation.stopAbort()
		}
		operation.cancel()
		if operation.gateHeld.CompareAndSwap(true, false) {
			operation.state.gate <- struct{}{}
		}
		operation.state.complete(operation)
	})
}

func (operation *boundOperation) cancelWithoutRegistration() {
	if operation.stopAbort != nil {
		operation.stopAbort()
	}
	operation.cancel()
}
