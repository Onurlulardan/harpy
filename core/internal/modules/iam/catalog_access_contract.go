package iam

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrCatalogAccessAuthentication = errors.New("IAM Catalog authentication failed")
	ErrCatalogAccessForbidden      = errors.New("IAM Catalog access is forbidden")
	ErrCatalogAccessUnavailable    = errors.New("IAM Catalog access is unavailable")
)

type CatalogAccessRequest struct {
	AuthorizationValues []string
	TenantIDValues      []string
	TenantSlugValues    []string
	Permission          string
	Operation           string
	PurposeTarget       string
}

type CatalogAccessDecision struct {
	principalID, tenantID                                  uuid.UUID
	tokenVersion, authorizationVersion, authorizationClock int64
	stepUpAuthenticatedAt                                  time.Time
	permissionID                                           uuid.UUID
	authorizationEvidenceSHA256, stepUpPurposeSHA256       [sha256.Size]byte
}

func NewCatalogAccessDecision(
	principalID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	stepUpAuthenticatedAt time.Time,
) (CatalogAccessDecision, error) {
	decision := CatalogAccessDecision{principalID: principalID, tenantID: tenantID,
		tokenVersion: tokenVersion, authorizationVersion: authorizationVersion,
		authorizationClock:    authorizationClock,
		stepUpAuthenticatedAt: stepUpAuthenticatedAt.UTC()}
	if principalID == uuid.Nil || tenantID == uuid.Nil || tokenVersion < 1 ||
		authorizationVersion < 1 || authorizationClock < 1 ||
		(!stepUpAuthenticatedAt.IsZero() && stepUpAuthenticatedAt.Nanosecond() != 0) {
		return CatalogAccessDecision{}, fmt.Errorf("IAM Catalog decision is invalid")
	}
	return decision, nil
}

func NewCatalogAccessDecisionWithEvidence(
	principalID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	stepUpAuthenticatedAt time.Time,
	permissionID uuid.UUID,
	authorizationEvidenceSHA256, stepUpPurposeSHA256 [sha256.Size]byte,
) (CatalogAccessDecision, error) {
	decision, err := NewCatalogAccessDecision(
		principalID, tenantID, tokenVersion, authorizationVersion,
		authorizationClock, stepUpAuthenticatedAt,
	)
	if err != nil || permissionID == uuid.Nil ||
		authorizationEvidenceSHA256 == [sha256.Size]byte{} ||
		(!stepUpAuthenticatedAt.IsZero() &&
			stepUpPurposeSHA256 == [sha256.Size]byte{}) ||
		(stepUpAuthenticatedAt.IsZero() &&
			stepUpPurposeSHA256 != [sha256.Size]byte{}) {
		return CatalogAccessDecision{}, fmt.Errorf("IAM Catalog evidence is invalid")
	}
	decision.permissionID = permissionID
	decision.authorizationEvidenceSHA256 = authorizationEvidenceSHA256
	decision.stepUpPurposeSHA256 = stepUpPurposeSHA256
	return decision, nil
}

func (decision CatalogAccessDecision) PrincipalID() uuid.UUID { return decision.principalID }
func (decision CatalogAccessDecision) TenantID() uuid.UUID    { return decision.tenantID }
func (decision CatalogAccessDecision) TokenVersion() int64    { return decision.tokenVersion }
func (decision CatalogAccessDecision) AuthorizationVersion() int64 {
	return decision.authorizationVersion
}
func (decision CatalogAccessDecision) AuthorizationClock() int64 {
	return decision.authorizationClock
}
func (decision CatalogAccessDecision) StepUpAuthenticatedAt() time.Time {
	return decision.stepUpAuthenticatedAt
}
func (decision CatalogAccessDecision) PermissionID() uuid.UUID {
	return decision.permissionID
}
func (decision CatalogAccessDecision) AuthorizationEvidenceSHA256() [sha256.Size]byte {
	return decision.authorizationEvidenceSHA256
}
func (decision CatalogAccessDecision) StepUpPurposeSHA256() [sha256.Size]byte {
	return decision.stepUpPurposeSHA256
}
