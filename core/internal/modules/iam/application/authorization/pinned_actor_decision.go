package authorization

import (
	"fmt"

	"github.com/google/uuid"
)

type PinnedActorDecision struct {
	decision     Decision
	tokenVersion int64
	valid        bool
}

func newPinnedActorDecision(decision Decision, tokenVersion int64) PinnedActorDecision {
	return PinnedActorDecision{decision: decision, tokenVersion: tokenVersion, valid: true}
}

func (decision PinnedActorDecision) Validate() error {
	if !decision.valid || decision.decision.Validate() != nil ||
		decision.tokenVersion < 0 || decision.decision.Allowed() &&
		decision.tokenVersion < 1 {
		return fmt.Errorf("IAM pinned actor authorization decision is invalid")
	}
	return nil
}

func (decision PinnedActorDecision) Allowed() bool {
	return decision.Validate() == nil && decision.decision.Allowed()
}

func (decision PinnedActorDecision) Reason() Reason {
	if decision.Validate() != nil {
		return ReasonEvidenceUnavailable
	}
	return decision.decision.Reason()
}

func (decision PinnedActorDecision) TokenVersion() int64 {
	if decision.Validate() != nil {
		return 0
	}
	return decision.tokenVersion
}

func (decision PinnedActorDecision) AuthorizationVersion() int64 {
	if decision.Validate() != nil {
		return 0
	}
	return decision.decision.AuthorizationVersion()
}

func (decision PinnedActorDecision) AuthorizationClock() int64 {
	if decision.Validate() != nil {
		return 0
	}
	return decision.decision.AuthorizationClock()
}

func (decision PinnedActorDecision) PermissionID() (uuid.UUID, bool) {
	if decision.Validate() != nil {
		return uuid.Nil, false
	}
	return decision.decision.PermissionID()
}
