package offering

import "errors"

var (
	ErrKey     = errors.New("billing offering key is invalid")
	ErrPrice   = errors.New("billing offering price is invalid")
	ErrGrant   = errors.New("billing offering entitlement grant is invalid")
	ErrVersion = errors.New("billing offering version is invalid")
	ErrInstant = errors.New("billing offering instant is invalid")
)
