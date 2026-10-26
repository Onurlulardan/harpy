package iam

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrDocumentsAccessAuthentication = errors.New("IAM Documents authentication failed")
	ErrDocumentsAccessForbidden      = errors.New("IAM Documents access is forbidden")
	ErrDocumentsAccessUnavailable    = errors.New("IAM Documents access is unavailable")
)

type DocumentsAccessRequest struct {
	AuthorizationValues []string
	TenantIDValues      []string
	TenantSlugValues    []string
	Permission          string
	Operation           string
	PurposeTarget       string
}

type DocumentsAccessDecision struct {
	principalID, tenantID                                  uuid.UUID
	tokenVersion, authorizationVersion, authorizationClock int64
	stepUpAuthenticatedAt                                  time.Time
}

func NewDocumentsAccessDecision(principalID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	stepUpAuthenticatedAt time.Time,
) (DocumentsAccessDecision, error) {
	decision := DocumentsAccessDecision{principalID: principalID, tenantID: tenantID,
		tokenVersion: tokenVersion, authorizationVersion: authorizationVersion,
		authorizationClock:    authorizationClock,
		stepUpAuthenticatedAt: stepUpAuthenticatedAt.UTC()}
	if principalID == uuid.Nil || tenantID == uuid.Nil || tokenVersion < 1 ||
		authorizationVersion < 1 || authorizationClock < 1 ||
		(!stepUpAuthenticatedAt.IsZero() && stepUpAuthenticatedAt.Nanosecond() != 0) {
		return DocumentsAccessDecision{}, fmt.Errorf("IAM Documents decision is invalid")
	}
	return decision, nil
}

func (decision DocumentsAccessDecision) PrincipalID() uuid.UUID { return decision.principalID }
func (decision DocumentsAccessDecision) TenantID() uuid.UUID    { return decision.tenantID }
func (decision DocumentsAccessDecision) TokenVersion() int64    { return decision.tokenVersion }
func (decision DocumentsAccessDecision) AuthorizationVersion() int64 {
	return decision.authorizationVersion
}
func (decision DocumentsAccessDecision) AuthorizationClock() int64 {
	return decision.authorizationClock
}
func (decision DocumentsAccessDecision) StepUpAuthenticatedAt() time.Time {
	return decision.stepUpAuthenticatedAt
}
