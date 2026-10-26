package iam

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrUsageAccessAuthentication = errors.New("IAM Usage access authentication failed")
	ErrUsageAccessForbidden      = errors.New("IAM Usage access is forbidden")
	ErrUsageAccessUnavailable    = errors.New("IAM Usage access is unavailable")
)

type UsageAccessRequest struct {
	AuthorizationValues []string
	TenantIDValues      []string
	TenantSlugValues    []string
	Scope               string
	Permission          string
	Operation           string
}

type UsageAccessDecision struct {
	actorID, subjectTenantID, targetTenantID, environmentID uuid.UUID
	tokenVersion, authorizationVersion, authorityClock      int64
	scope, mode, purpose                                    string
	controlPrimary                                          bool
	permission, operation                                   string
}

func NewUsageAccessDecision(
	actorID, subjectTenantID, targetTenantID, environmentID uuid.UUID,
	tokenVersion, authorizationVersion, authorityClock int64,
	scope, mode, purpose string, controlPrimary bool,
	permission, operation string,
) (UsageAccessDecision, error) {
	decision := UsageAccessDecision{
		actorID: actorID, subjectTenantID: subjectTenantID,
		targetTenantID: targetTenantID, environmentID: environmentID,
		tokenVersion: tokenVersion, authorizationVersion: authorizationVersion,
		authorityClock: authorityClock, scope: scope, mode: mode, purpose: purpose,
		controlPrimary: controlPrimary, permission: permission, operation: operation,
	}
	if !decision.valid() {
		return UsageAccessDecision{}, fmt.Errorf("IAM Usage access decision is invalid")
	}
	return decision, nil
}

func (decision UsageAccessDecision) valid() bool {
	if decision.actorID == uuid.Nil || decision.subjectTenantID == uuid.Nil ||
		decision.environmentID == uuid.Nil || decision.tokenVersion < 1 ||
		decision.authorizationVersion < 1 || decision.authorityClock < 1 {
		return false
	}
	if _, valid := resolveUsageAccess(
		decision.scope, decision.permission, decision.operation,
	); !valid {
		return false
	}
	if decision.scope == "tenant" {
		purposeValid := decision.mode == "onpremise" && decision.purpose == "onpremise_primary" ||
			decision.mode == "saas" && decision.purpose == "saas_customer"
		return purposeValid && !decision.controlPrimary &&
			decision.targetTenantID == decision.subjectTenantID
	}
	return decision.scope == "platform" && decision.mode == "saas" &&
		decision.purpose == "saas_control" && decision.controlPrimary &&
		decision.targetTenantID == uuid.Nil
}

func (decision UsageAccessDecision) ActorID() uuid.UUID { return decision.actorID }
func (decision UsageAccessDecision) SubjectTenantID() uuid.UUID {
	return decision.subjectTenantID
}
func (decision UsageAccessDecision) TargetTenantID() uuid.UUID { return decision.targetTenantID }
func (decision UsageAccessDecision) EnvironmentID() uuid.UUID  { return decision.environmentID }
func (decision UsageAccessDecision) TokenVersion() int64       { return decision.tokenVersion }
func (decision UsageAccessDecision) AuthorizationVersion() int64 {
	return decision.authorizationVersion
}
func (decision UsageAccessDecision) AuthorityClock() int64 { return decision.authorityClock }
func (decision UsageAccessDecision) Scope() string         { return decision.scope }
func (decision UsageAccessDecision) Mode() string          { return decision.mode }
func (decision UsageAccessDecision) Purpose() string       { return decision.purpose }
func (decision UsageAccessDecision) ControlPrimary() bool  { return decision.controlPrimary }
func (decision UsageAccessDecision) Permission() string    { return decision.permission }
func (decision UsageAccessDecision) Operation() string     { return decision.operation }
