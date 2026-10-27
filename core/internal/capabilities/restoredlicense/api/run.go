package api

import "time"

type Callback func(Scope) error

func Run(grant Grant, databaseNow time.Time, callback Callback) error {
	if !grant.valid(databaseNow) || callback == nil {
		return ErrInvalid
	}
	state := &scopeState{grant: grant}
	state.active.Store(true)
	scope := Scope{state: state}
	defer state.active.Store(false)
	if err := callback(scope); err != nil {
		return err
	}
	return closeScope(scope)
}
