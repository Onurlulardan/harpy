package iam

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrAuditAccessAuthentication = errors.New("IAM Audit access authentication failed")
	ErrAuditAccessForbidden      = errors.New("IAM Audit access is forbidden")
	ErrAuditAccessUnavailable    = errors.New("IAM Audit access is unavailable")
)

type AuditAccessRequest struct {
	AuthorizationValues []string
	TenantIDValues      []string
	TenantSlugValues    []string
	Permission          string
	Operation           string
}

type AuditAccessDecision struct {
	principalID          uuid.UUID
	tenantID             uuid.UUID
	tokenVersion         int64
	authorizationVersion int64
	authorizationClock   int64
	permission           string
	operation            string
}

func NewAuditAccessDecision(
	principalID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	permission, operation string,
) (AuditAccessDecision, error) {
	decision := AuditAccessDecision{
		principalID: principalID, tenantID: tenantID,
		tokenVersion: tokenVersion, authorizationVersion: authorizationVersion,
		authorizationClock: authorizationClock,
		permission:         permission, operation: operation,
	}
	if principalID == uuid.Nil || tenantID == uuid.Nil || tokenVersion < 1 ||
		authorizationVersion < 1 || authorizationClock < 1 {
		return AuditAccessDecision{}, fmt.Errorf("IAM Audit access decision is invalid")
	}
	if _, valid := resolveAuditAccess(permission, operation); !valid {
		return AuditAccessDecision{}, fmt.Errorf("IAM Audit authority is invalid")
	}
	return decision, nil
}

func (decision AuditAccessDecision) PrincipalID() uuid.UUID { return decision.principalID }
func (decision AuditAccessDecision) TenantID() uuid.UUID    { return decision.tenantID }
func (decision AuditAccessDecision) TokenVersion() int64    { return decision.tokenVersion }
func (decision AuditAccessDecision) AuthorizationVersion() int64 {
	return decision.authorizationVersion
}
func (decision AuditAccessDecision) AuthorizationClock() int64 {
	return decision.authorizationClock
}
func (decision AuditAccessDecision) Permission() string { return decision.permission }
func (decision AuditAccessDecision) Operation() string  { return decision.operation }
