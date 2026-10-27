package iam

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrFormsLifecycleAccessAuthentication = errors.New("IAM Forms lifecycle authentication failed")
	ErrFormsLifecycleAccessForbidden      = errors.New("IAM Forms lifecycle access is forbidden")
	ErrFormsLifecycleAccessUnavailable    = errors.New("IAM Forms lifecycle access is unavailable")
)

type FormsLifecycleAccessRequest struct {
	AuthorizationValues []string
	TenantIDValues      []string
	TenantSlugValues    []string
	Permission          string
	Operation           string
}

type FormsLifecycleAccessDecision struct {
	principalID, tenantID                                  uuid.UUID
	tokenVersion, authorizationVersion, authorizationClock int64
}

func NewFormsLifecycleAccessDecision(
	principalID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
) (FormsLifecycleAccessDecision, error) {
	decision := FormsLifecycleAccessDecision{principalID: principalID, tenantID: tenantID,
		tokenVersion: tokenVersion, authorizationVersion: authorizationVersion,
		authorizationClock: authorizationClock}
	if principalID == uuid.Nil || tenantID == uuid.Nil || tokenVersion < 1 ||
		authorizationVersion < 1 || authorizationClock < 1 {
		return FormsLifecycleAccessDecision{}, fmt.Errorf("IAM Forms lifecycle decision is invalid")
	}
	return decision, nil
}

func (decision FormsLifecycleAccessDecision) PrincipalID() uuid.UUID { return decision.principalID }
func (decision FormsLifecycleAccessDecision) TenantID() uuid.UUID    { return decision.tenantID }
func (decision FormsLifecycleAccessDecision) TokenVersion() int64    { return decision.tokenVersion }
func (decision FormsLifecycleAccessDecision) AuthorizationVersion() int64 {
	return decision.authorizationVersion
}
func (decision FormsLifecycleAccessDecision) AuthorizationClock() int64 {
	return decision.authorizationClock
}
