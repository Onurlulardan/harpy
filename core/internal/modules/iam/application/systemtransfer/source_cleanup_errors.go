package systemtransfer

import "errors"

var (
	ErrSourceCleanupInvalid     = errors.New("IAM import source cleanup input is invalid")
	ErrSourceCleanupUnavailable = errors.New(
		"IAM import source cleanup dependency is unavailable",
	)
	ErrSourceCleanupStale         = errors.New("IAM import source cleanup lease is stale")
	ErrSourceDeleteOutcomeUnknown = errors.New(
		"IAM import source delete outcome is unknown",
	)
)
