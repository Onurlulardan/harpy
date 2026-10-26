package iam

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrTenantSettingsAccessAuthentication = errors.New(
		"IAM tenant settings access authentication failed")
	ErrTenantSettingsAccessForbidden = errors.New(
		"IAM tenant settings access is forbidden")
	ErrTenantSettingsAccessUnavailable = errors.New(
		"IAM tenant settings access is unavailable")
)

type TenantSettingsAccessRequest struct {
	AuthorizationValues []string
	TenantIDValues      []string
	TenantSlugValues    []string
	Permission          string
	Operation           string
}

type TenantSettingsAccessDecision struct {
	principalID          uuid.UUID
	tenantID             uuid.UUID
	tokenVersion         int64
	authorizationVersion int64
	authorizationClock   int64
	permission           string
	operation            string
}

func NewTenantSettingsAccessDecision(
	principalID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	permission, operation string,
) (TenantSettingsAccessDecision, error) {
	decision := TenantSettingsAccessDecision{
		principalID: principalID, tenantID: tenantID,
		tokenVersion: tokenVersion, authorizationVersion: authorizationVersion,
		authorizationClock: authorizationClock,
		permission:         permission, operation: operation,
	}
	if principalID == uuid.Nil || tenantID == uuid.Nil || tokenVersion < 1 ||
		authorizationVersion < 1 || authorizationClock < 1 {
		return TenantSettingsAccessDecision{}, fmt.Errorf(
			"IAM tenant settings access decision is invalid")
	}
	if _, valid := resolveTenantSettingsAccess(permission, operation); !valid {
		return TenantSettingsAccessDecision{}, fmt.Errorf(
			"IAM tenant settings authority is invalid")
	}
	return decision, nil
}

func (decision TenantSettingsAccessDecision) PrincipalID() uuid.UUID {
	return decision.principalID
}
func (decision TenantSettingsAccessDecision) TenantID() uuid.UUID { return decision.tenantID }
func (decision TenantSettingsAccessDecision) TokenVersion() int64 {
	return decision.tokenVersion
}
func (decision TenantSettingsAccessDecision) AuthorizationVersion() int64 {
	return decision.authorizationVersion
}
func (decision TenantSettingsAccessDecision) AuthorizationClock() int64 {
	return decision.authorizationClock
}
func (decision TenantSettingsAccessDecision) Permission() string { return decision.permission }
func (decision TenantSettingsAccessDecision) Operation() string  { return decision.operation }
