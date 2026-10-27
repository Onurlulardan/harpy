package iam

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrWorkflowAccessAuthentication = errors.New("IAM Workflow authentication failed")
	ErrWorkflowAccessForbidden      = errors.New("IAM Workflow access is forbidden")
	ErrWorkflowAccessUnavailable    = errors.New("IAM Workflow access is unavailable")
)

type WorkflowAccessRequest struct {
	AuthorizationValues []string
	TenantIDValues      []string
	TenantSlugValues    []string
	Permission          string
	Operation           string
	TargetID            string
}

type WorkflowAccessDecision struct {
	principalID, tenantID                                  uuid.UUID
	tokenVersion, authorizationVersion, authorizationClock int64
	stepUpPurpose                                          [32]byte
	stepUpAuthenticatedAt                                  time.Time
	amrPassword, amrWebAuthn, amrTOTP                      bool
	publicationAuthority                                   bool
}

func NewWorkflowPublicationAuthorityAccessDecision(
	principalID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	purpose [32]byte,
	authenticatedAt time.Time,
	amrPassword, amrWebAuthn, amrTOTP bool,
) (WorkflowAccessDecision, error) {
	decision, err := NewWorkflowAccessDecision(
		principalID, tenantID, tokenVersion, authorizationVersion, authorizationClock,
	)
	authenticatedAt = authenticatedAt.UTC()
	if err != nil || purpose == [32]byte{} || authenticatedAt.IsZero() ||
		authenticatedAt.Nanosecond() != 0 ||
		!(amrPassword || amrWebAuthn || amrTOTP) {
		return WorkflowAccessDecision{}, fmt.Errorf(
			"IAM Workflow publication authority decision is invalid",
		)
	}
	decision.stepUpPurpose = purpose
	decision.stepUpAuthenticatedAt = authenticatedAt
	decision.amrPassword = amrPassword
	decision.amrWebAuthn = amrWebAuthn
	decision.amrTOTP = amrTOTP
	decision.publicationAuthority = true
	return decision, nil
}

func NewWorkflowAccessDecision(
	principalID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
) (WorkflowAccessDecision, error) {
	decision := WorkflowAccessDecision{principalID: principalID, tenantID: tenantID,
		tokenVersion: tokenVersion, authorizationVersion: authorizationVersion,
		authorizationClock: authorizationClock}
	if principalID == uuid.Nil || tenantID == uuid.Nil || tokenVersion < 1 ||
		authorizationVersion < 1 || authorizationClock < 1 {
		return WorkflowAccessDecision{}, fmt.Errorf("IAM Workflow decision is invalid")
	}
	return decision, nil
}

func (decision WorkflowAccessDecision) PrincipalID() uuid.UUID { return decision.principalID }
func (decision WorkflowAccessDecision) TenantID() uuid.UUID    { return decision.tenantID }
func (decision WorkflowAccessDecision) TokenVersion() int64    { return decision.tokenVersion }
func (decision WorkflowAccessDecision) AuthorizationVersion() int64 {
	return decision.authorizationVersion
}
func (decision WorkflowAccessDecision) AuthorizationClock() int64 {
	return decision.authorizationClock
}
func (decision WorkflowAccessDecision) PublicationAuthorityEvidence() (
	[32]byte, time.Time, bool, bool, bool, bool,
) {
	return decision.stepUpPurpose, decision.stepUpAuthenticatedAt,
		decision.amrPassword, decision.amrWebAuthn, decision.amrTOTP,
		decision.publicationAuthority
}
