package postgres

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type bindingState struct {
	accepting  atomic.Bool
	sealing    atomic.Bool
	mutex      sync.Mutex
	operations map[*boundOperation]struct{}
	gate       chan struct{}
	idle       chan struct{}
	idleOnce   sync.Once
	sealOnce   sync.Once
	sealDone   chan struct{}
	sealErr    error
}

func newBindingState() *bindingState {
	state := &bindingState{
		operations: make(map[*boundOperation]struct{}),
		gate:       make(chan struct{}, 1), idle: make(chan struct{}),
		sealDone: make(chan struct{}),
	}
	state.gate <- struct{}{}
	state.accepting.Store(true)
	return state
}

func (state *bindingState) usable(transaction *RuntimeTransaction) bool {
	return state != nil && state.accepting.Load() &&
		transaction != nil && transaction.usable()
}

func (state *bindingState) acquire(
	ctx context.Context,
	transaction *RuntimeTransaction,
) (*boundOperation, error) {
	if state == nil || ctx == nil || !state.accepting.Load() ||
		transaction == nil || !transaction.usable() {
		return nil, ErrTransactionClosed
	}
	operation := newBoundOperation(ctx, state)
	state.mutex.Lock()
	if !state.accepting.Load() || !transaction.usable() {
		state.mutex.Unlock()
		operation.cancelWithoutRegistration()
		return nil, ErrTransactionClosed
	}
	state.operations[operation] = struct{}{}
	state.mutex.Unlock()
	select {
	case <-operation.context.Done():
		operation.finish()
		return nil, state.operationError(operation.context)
	case <-state.gate:
		operation.gateHeld.Store(true)
	}
	if !state.accepting.Load() || !transaction.usable() || operation.context.Err() != nil {
		operation.finish()
		return nil, state.operationError(operation.context)
	}
	return operation, nil
}

func (state *bindingState) operationError(ctx context.Context) error {
	if state == nil || !state.accepting.Load() {
		return ErrTransactionClosed
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return ErrTransactionClosed
}

func (state *bindingState) complete(operation *boundOperation) {
	state.mutex.Lock()
	delete(state.operations, operation)
	if state.sealing.Load() && len(state.operations) == 0 {
		state.idleOnce.Do(func() { close(state.idle) })
	}
	state.mutex.Unlock()
}

func (state *bindingState) seal(timeout time.Duration) error {
	if state == nil {
		return nil
	}
	state.sealOnce.Do(func() {
		state.accepting.Store(false)
		state.sealing.Store(true)
		state.sealErr = state.stopOperations(timeout)
		close(state.sealDone)
	})
	<-state.sealDone
	return state.sealErr
}

func (state *bindingState) stopOperations(timeout time.Duration) error {
	state.mutex.Lock()
	operations := make([]*boundOperation, 0, len(state.operations))
	for operation := range state.operations {
		operations = append(operations, operation)
	}
	if len(operations) == 0 {
		state.idleOnce.Do(func() { close(state.idle) })
	}
	state.mutex.Unlock()
	if len(operations) == 0 {
		return nil
	}
	for _, operation := range operations {
		operation.requestAbort()
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-state.idle:
		return ErrBindingAbandoned
	case <-timer.C:
		return ErrBindingCleanupTimeout
	}
}
