package iam

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrFormsRecordAccessAuthentication = errors.New("IAM Forms record access authentication failed")
	ErrFormsRecordAccessForbidden      = errors.New("IAM Forms record access is forbidden")
	ErrFormsRecordAccessUnavailable    = errors.New("IAM Forms record access is unavailable")
)

type FormsRecordAccessRequest struct {
	AuthorizationValues []string
	TenantIDValues      []string
	TenantSlugValues    []string
	Action              string
	Operation           string
}

type FormsRecordAccessDecision struct {
	principalID          uuid.UUID
	tenantID             uuid.UUID
	tokenVersion         int64
	authorizationVersion int64
	authorizationClock   int64
}

func NewFormsRecordAccessDecision(
	principalID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
) (FormsRecordAccessDecision, error) {
	decision := FormsRecordAccessDecision{
		principalID: principalID, tenantID: tenantID,
		tokenVersion: tokenVersion, authorizationVersion: authorizationVersion,
		authorizationClock: authorizationClock,
	}
	if principalID == uuid.Nil || tenantID == uuid.Nil || tokenVersion < 1 ||
		authorizationVersion < 1 || authorizationClock < 1 {
		return FormsRecordAccessDecision{}, fmt.Errorf("IAM Forms record access decision is invalid")
	}
	return decision, nil
}

func (decision FormsRecordAccessDecision) PrincipalID() uuid.UUID { return decision.principalID }
func (decision FormsRecordAccessDecision) TenantID() uuid.UUID    { return decision.tenantID }
func (decision FormsRecordAccessDecision) TokenVersion() int64    { return decision.tokenVersion }
func (decision FormsRecordAccessDecision) AuthorizationVersion() int64 {
	return decision.authorizationVersion
}
func (decision FormsRecordAccessDecision) AuthorizationClock() int64 {
	return decision.authorizationClock
}
