package api

import (
	"errors"
	"sync/atomic"

	"github.com/google/uuid"
)

type scopeState struct {
	active atomic.Bool
	next   atomic.Int32
	grant  Grant
}

type Scope struct{ state *scopeState }

func (scope Scope) OperationID() uuid.UUID {
	if scope.state == nil || !scope.state.active.Load() {
		return uuid.Nil
	}
	return scope.state.grant.Binding.OperationID
}

func (scope Scope) TenantID() uuid.UUID {
	if scope.state == nil || !scope.state.active.Load() {
		return uuid.Nil
	}
	return scope.state.grant.Binding.TenantID
}

func (scope Scope) Kind() string {
	if scope.state == nil || !scope.state.active.Load() {
		return ""
	}
	return scope.state.grant.Binding.Kind
}

func (scope Scope) Binding() Binding {
	if scope.state == nil || !scope.state.active.Load() {
		return Binding{}
	}
	return scope.state.grant.Binding
}

func (scope Scope) Revision() int64 {
	if scope.state == nil || !scope.state.active.Load() {
		return 0
	}
	return scope.state.grant.Revision
}

func (scope Scope) Consume(command Command) error {
	if scope.state == nil || !scope.state.active.Load() {
		return ErrClosed
	}
	manifest := commands(scope.state.grant.Binding.Kind)
	index := int(scope.state.next.Load())
	if index >= len(manifest) || manifest[index] != command {
		for _, allowed := range manifest {
			if allowed == command {
				return ErrOrder
			}
		}
		return ErrDenied
	}
	if !scope.state.next.CompareAndSwap(int32(index), int32(index+1)) {
		return ErrOrder
	}
	return nil
}

func (Scope) MarshalJSON() ([]byte, error) {
	return nil, errors.New("tenant activation scope cannot be serialized")
}

func closeScope(scope Scope) error {
	if scope.state == nil || !scope.state.active.CompareAndSwap(true, false) {
		return ErrClosed
	}
	if int(scope.state.next.Load()) != len(commands(scope.state.grant.Binding.Kind)) {
		return ErrOrder
	}
	return nil
}
