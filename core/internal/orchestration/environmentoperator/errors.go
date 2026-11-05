package environmentoperator

import "errors"

var (
	ErrInvalid                   = errors.New("environment initialize input is invalid")
	ErrDependencyMismatch        = errors.New("environment initialize dependency mismatch")
	ErrResumable                 = errors.New("environment initialize is resumable")
	ErrFailed                    = errors.New("environment initialize failed")
	ErrFenced                    = errors.New("environment initialize is fenced")
	ErrTenantProvisioningPending = errors.New(
		"tenant provisioning is waiting for host execution",
	)
)

type Outcome string

const (
	OutcomeCompleted Outcome = "completed"
	OutcomeNoop      Outcome = "completed_noop"
)
