package api

import (
	"errors"
	"sync/atomic"

	licensebootstrapv1 "github.com/Onurlulardan/harpy/contracts/licensebootstrap/v1"
)

type scopeState struct {
	active   atomic.Bool
	consumed atomic.Bool
	grant    Grant
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

func (scope Scope) Consume(command licensebootstrapv1.Command) error {
	if scope.state == nil || !scope.state.active.Load() {
		return ErrClosed
	}
	if command != licensebootstrapv1.PublishLicenseDerivedOffering {
		return ErrDenied
	}
	if !scope.state.consumed.CompareAndSwap(false, true) {
		return ErrOrder
	}
	return nil
}

func (Scope) MarshalJSON() ([]byte, error) {
	return nil, errors.New("license bootstrap scope cannot be serialized")
}

func closeScope(scope Scope) error {
	if scope.state == nil || !scope.state.active.CompareAndSwap(true, false) {
		return ErrClosed
	}
	if !scope.state.consumed.Load() {
		return ErrOrder
	}
	return nil
}
