package systemtransferartifact

import "errors"

var (
	errInvalid        = errors.New("IAM system transfer artifact input is invalid")
	errUnavailable    = errors.New("IAM system transfer artifact runtime is unavailable")
	errIntegrity      = errors.New("IAM system transfer artifact integrity failed")
	errOutcomeUnknown = errors.New("IAM system transfer artifact outcome is unknown")
	errPlaintextLimit = errors.New("IAM system transfer artifact plaintext limit exceeded")
)
