package authorization

import (
	"fmt"

	"github.com/google/uuid"
)

type Decision struct {
	allowed              bool
	reason               Reason
	authorizationClock   int64
	authorizationVersion int64
	permissionID         uuid.UUID
	valid                bool
}

func newDecision(
	allowed bool,
	reason Reason,
	clock int64,
	version int64,
	permissionID uuid.UUID,
) Decision {
	return Decision{
		allowed: allowed, reason: reason, authorizationClock: clock,
		authorizationVersion: version, permissionID: permissionID, valid: true,
	}
}

func deniedDecision(reason Reason, clock, version int64) Decision {
	return newDecision(false, reason, clock, version, uuid.Nil)
}

func allowedDecision(
	reason Reason,
	clock int64,
	version int64,
	permissionID uuid.UUID,
) Decision {
	return newDecision(true, reason, clock, version, permissionID)
}

func (decision Decision) Validate() error {
	if !decision.valid || decision.authorizationClock < 0 ||
		decision.authorizationVersion < 0 ||
		(decision.allowed != (decision.permissionID != uuid.Nil)) ||
		(decision.allowed && !decision.reason.allowed()) ||
		(!decision.allowed && !decision.reason.denied()) {
		return fmt.Errorf("IAM authorization decision is invalid")
	}
	return nil
}

func (decision Decision) Allowed() bool {
	return decision.Validate() == nil && decision.allowed
}

func (decision Decision) Reason() Reason {
	if decision.Validate() != nil {
		return ReasonEvidenceUnavailable
	}
	return decision.reason
}

func (decision Decision) AuthorizationClock() int64 {
	if decision.Validate() != nil {
		return 0
	}
	return decision.authorizationClock
}

func (decision Decision) AuthorizationVersion() int64 {
	if decision.Validate() != nil {
		return 0
	}
	return decision.authorizationVersion
}

func (decision Decision) PermissionID() (uuid.UUID, bool) {
	if decision.Validate() != nil || !decision.allowed {
		return uuid.Nil, false
	}
	return decision.permissionID, true
}
