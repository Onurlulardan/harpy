package iam

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrNotificationsAccessAuthentication = errors.New("IAM Notifications authentication failed")
	ErrNotificationsAccessForbidden      = errors.New("IAM Notifications access is forbidden")
	ErrNotificationsAccessUnavailable    = errors.New("IAM Notifications access is unavailable")
)

type NotificationsAccessRequest struct {
	AuthorizationValues []string
	CookieValues        []string
	TenantIDValues      []string
	TenantSlugValues    []string
	CSRFValues          []string
	OriginValues        []string
	Permission          string
	Operation           string
	RequireSession      bool
	RequireCSRF         bool
	RequireHTTPSOrigin  bool
	RequireSecure       bool
	RequireConsumer     bool
}

type NotificationsAccessDecision struct {
	principalID, tenantID              uuid.UUID
	tokenVersion, authorizationVersion int64
	origin                             string
	sessionAuthenticated               bool
	csrfVerified                       bool
}

func NewNotificationsAccessDecision(principalID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion int64, origin string,
	sessionAuthenticated, csrfVerified bool,
) (NotificationsAccessDecision, error) {
	decision := NotificationsAccessDecision{principalID: principalID, tenantID: tenantID,
		tokenVersion: tokenVersion, authorizationVersion: authorizationVersion,
		origin: origin, sessionAuthenticated: sessionAuthenticated,
		csrfVerified: csrfVerified}
	if principalID == uuid.Nil || tenantID == uuid.Nil || tokenVersion < 1 ||
		authorizationVersion < 1 || csrfVerified && !sessionAuthenticated {
		return NotificationsAccessDecision{}, fmt.Errorf("IAM Notifications decision is invalid")
	}
	return decision, nil
}

func (decision NotificationsAccessDecision) PrincipalID() uuid.UUID {
	return decision.principalID
}
func (decision NotificationsAccessDecision) TenantID() uuid.UUID { return decision.tenantID }
func (decision NotificationsAccessDecision) TokenVersion() int64 { return decision.tokenVersion }
func (decision NotificationsAccessDecision) AuthorizationVersion() int64 {
	return decision.authorizationVersion
}
func (decision NotificationsAccessDecision) Origin() string { return decision.origin }
func (decision NotificationsAccessDecision) SessionAuthenticated() bool {
	return decision.sessionAuthenticated
}
func (decision NotificationsAccessDecision) CSRFVerified() bool {
	return decision.csrfVerified
}
