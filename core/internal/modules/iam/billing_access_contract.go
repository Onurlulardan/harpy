package iam

import (
	"errors"
	"fmt"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

var (
	ErrBillingAccessAuthentication = errors.New("IAM Billing access authentication failed")
	ErrBillingAccessForbidden      = errors.New("IAM Billing access is forbidden")
	ErrBillingAccessUnavailable    = errors.New("IAM Billing access is unavailable")
)

type BillingAccessRequest struct {
	AuthorizationValues []string
	TenantIDValues      []string
	TenantSlugValues    []string
	Permission          string
	Operation           string
}

type BillingAccessDecision struct {
	principalID          uuid.UUID
	tenantID             uuid.UUID
	tokenVersion         int64
	authorizationVersion int64
	authorizationClock   int64
	globalCommand        platformtx.GlobalCommandAccess
	globalSnapshot       platformtx.GlobalSnapshotCapability
	globalMode           billingGlobalAccessMode
}

type billingGlobalAccessMode uint8

const (
	billingGlobalCommand billingGlobalAccessMode = iota + 1
	billingGlobalSnapshot
)

func NewTenantBillingAccessDecision(
	principalID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
) (BillingAccessDecision, error) {
	decision := BillingAccessDecision{
		principalID: principalID, tenantID: tenantID,
		tokenVersion: tokenVersion, authorizationVersion: authorizationVersion,
		authorizationClock: authorizationClock,
	}
	if !decision.valid() {
		return BillingAccessDecision{}, fmt.Errorf("IAM tenant Billing access decision is invalid")
	}
	return decision, nil
}

func NewGlobalBillingCommandAccessDecision(
	principalID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	capability platformtx.GlobalCommandAccess,
) (BillingAccessDecision, error) {
	decision, err := NewTenantBillingAccessDecision(
		principalID, tenantID, tokenVersion, authorizationVersion, authorizationClock,
	)
	if err != nil {
		return BillingAccessDecision{}, err
	}
	decision.globalCommand, decision.globalMode = capability, billingGlobalCommand
	return decision, nil
}

func NewGlobalBillingSnapshotAccessDecision(
	principalID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	capability platformtx.GlobalSnapshotCapability,
) (BillingAccessDecision, error) {
	decision, err := NewTenantBillingAccessDecision(
		principalID, tenantID, tokenVersion, authorizationVersion, authorizationClock,
	)
	if err != nil {
		return BillingAccessDecision{}, err
	}
	decision.globalSnapshot, decision.globalMode = capability, billingGlobalSnapshot
	return decision, nil
}

func (decision BillingAccessDecision) valid() bool {
	return decision.principalID != uuid.Nil && decision.tenantID != uuid.Nil &&
		decision.tokenVersion > 0 && decision.authorizationVersion > 0 &&
		decision.authorizationClock > 0
}

func (decision BillingAccessDecision) PrincipalID() uuid.UUID { return decision.principalID }
func (decision BillingAccessDecision) TenantID() uuid.UUID    { return decision.tenantID }
func (decision BillingAccessDecision) TokenVersion() int64    { return decision.tokenVersion }
func (decision BillingAccessDecision) AuthorizationVersion() int64 {
	return decision.authorizationVersion
}
func (decision BillingAccessDecision) AuthorizationClock() int64 {
	return decision.authorizationClock
}
func (decision BillingAccessDecision) GlobalCommandAccess() (
	platformtx.GlobalCommandAccess,
	bool,
) {
	return decision.globalCommand, decision.globalMode == billingGlobalCommand
}

func (decision BillingAccessDecision) GlobalSnapshotCapability() (
	platformtx.GlobalSnapshotCapability,
	bool,
) {
	return decision.globalSnapshot, decision.globalMode == billingGlobalSnapshot
}
