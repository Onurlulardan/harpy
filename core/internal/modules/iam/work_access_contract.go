package iam

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrWorkAccessAuthentication = errors.New("IAM Work authentication failed")
	ErrWorkAccessForbidden      = errors.New("IAM Work access is forbidden")
	ErrWorkAccessUnavailable    = errors.New("IAM Work access is unavailable")
)

type WorkAccessRequest struct {
	AuthorizationValues []string
	TenantIDValues      []string
	TenantSlugValues    []string
	Authority           string
	Operation           string
}

type WorkAccessDecision struct {
	principalID, tenantID                                  uuid.UUID
	tokenVersion, authorizationVersion, authorizationClock int64
	protectedAdmin                                         bool
}

func NewWorkAccessDecision(
	principalID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
) (WorkAccessDecision, error) {
	decision := WorkAccessDecision{
		principalID: principalID, tenantID: tenantID,
		tokenVersion: tokenVersion, authorizationVersion: authorizationVersion,
		authorizationClock: authorizationClock,
	}
	if principalID == uuid.Nil || tenantID == uuid.Nil || tokenVersion < 1 ||
		authorizationVersion < 1 || authorizationClock < 1 {
		return WorkAccessDecision{}, fmt.Errorf("IAM Work decision is invalid")
	}
	return decision, nil
}

func NewProtectedWorkAccessDecision(
	principalID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	protectedAdmin bool,
) (WorkAccessDecision, error) {
	decision, err := NewWorkAccessDecision(
		principalID, tenantID, tokenVersion, authorizationVersion, authorizationClock,
	)
	if err != nil {
		return WorkAccessDecision{}, err
	}
	decision.protectedAdmin = protectedAdmin
	return decision, nil
}

func (decision WorkAccessDecision) PrincipalID() uuid.UUID { return decision.principalID }
func (decision WorkAccessDecision) TenantID() uuid.UUID    { return decision.tenantID }
func (decision WorkAccessDecision) TokenVersion() int64    { return decision.tokenVersion }
func (decision WorkAccessDecision) AuthorizationVersion() int64 {
	return decision.authorizationVersion
}
func (decision WorkAccessDecision) AuthorizationClock() int64 {
	return decision.authorizationClock
}
func (decision WorkAccessDecision) ProtectedAdmin() bool { return decision.protectedAdmin }
