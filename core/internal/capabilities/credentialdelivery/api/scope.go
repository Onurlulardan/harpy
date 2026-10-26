package api

import (
	"errors"
	"sync/atomic"

	credentialdeliveryv1 "github.com/Onurlulardan/harpy/contracts/credentialdelivery/v1"
)

type scopeState struct {
	active atomic.Bool
	next   atomic.Uint32
	grant  Grant
}

type Scope struct{ state *scopeState }

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

func (scope Scope) Consume(command credentialdeliveryv1.Command) error {
	if scope.state == nil || !scope.state.active.Load() {
		return ErrClosed
	}
	manifest := credentialdeliveryv1.Manifest()
	index := scope.state.next.Load()
	if index >= uint32(len(manifest)) {
		return ErrOrder
	}
	if command != manifest[index] {
		for _, allowed := range manifest {
			if command == allowed {
				return ErrOrder
			}
		}
		return ErrDenied
	}
	if !scope.state.next.CompareAndSwap(index, index+1) {
		return ErrOrder
	}
	return nil
}

func (Scope) MarshalJSON() ([]byte, error) {
	return nil, errors.New("credential delivery scope cannot be serialized")
}

func closeScope(scope Scope) error {
	if scope.state == nil || !scope.state.active.CompareAndSwap(true, false) {
		return ErrClosed
	}
	if scope.state.next.Load() != uint32(len(credentialdeliveryv1.Manifest())) {
		return ErrOrder
	}
	return nil
}
