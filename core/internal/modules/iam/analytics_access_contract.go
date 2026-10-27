package iam

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrAnalyticsAccessAuthentication = errors.New("IAM Analytics authentication failed")
	ErrAnalyticsAccessForbidden      = errors.New("IAM Analytics access is forbidden")
	ErrAnalyticsAccessUnavailable    = errors.New("IAM Analytics access is unavailable")
)

type AnalyticsAccessRequest struct {
	AuthorizationValues []string
	TenantIDValues      []string
	TenantSlugValues    []string
	Permission          string
	Operation           string
}

type AnalyticsAccessDecision struct {
	principalID, tenantID                                  uuid.UUID
	tokenVersion, authorizationVersion, authorizationClock int64
	stepUpAuthenticatedAt                                  time.Time
}

func NewAnalyticsAccessDecision(
	principalID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	stepUpAuthenticatedAt time.Time,
) (AnalyticsAccessDecision, error) {
	decision := AnalyticsAccessDecision{principalID: principalID, tenantID: tenantID,
		tokenVersion: tokenVersion, authorizationVersion: authorizationVersion,
		authorizationClock:    authorizationClock,
		stepUpAuthenticatedAt: stepUpAuthenticatedAt.UTC()}
	if principalID == uuid.Nil || tenantID == uuid.Nil || tokenVersion < 1 ||
		authorizationVersion < 1 || authorizationClock < 1 ||
		(!stepUpAuthenticatedAt.IsZero() && stepUpAuthenticatedAt.Nanosecond() != 0) {
		return AnalyticsAccessDecision{}, fmt.Errorf("IAM Analytics decision is invalid")
	}
	return decision, nil
}

func (decision AnalyticsAccessDecision) PrincipalID() uuid.UUID { return decision.principalID }
func (decision AnalyticsAccessDecision) TenantID() uuid.UUID    { return decision.tenantID }
func (decision AnalyticsAccessDecision) TokenVersion() int64    { return decision.tokenVersion }
func (decision AnalyticsAccessDecision) AuthorizationVersion() int64 {
	return decision.authorizationVersion
}
func (decision AnalyticsAccessDecision) AuthorizationClock() int64 {
	return decision.authorizationClock
}
func (decision AnalyticsAccessDecision) StepUpAuthenticatedAt() time.Time {
	return decision.stepUpAuthenticatedAt
}
