package entitlement

import "errors"

var (
	ErrConfiguration = errors.New("billing entitlement queries are misconfigured")
	ErrSnapshot      = errors.New("billing entitlement snapshot is invalid")
)
