package iam

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrSupportAccessAuthentication = errors.New("IAM Support access authentication failed")
	ErrSupportAccessForbidden      = errors.New("IAM Support access is forbidden")
	ErrSupportAccessUnavailable    = errors.New("IAM Support access is unavailable")
)

type SupportAccessRequest struct {
	AuthorizationValues []string
	TenantIDValues      []string
	TenantSlugValues    []string
	Scope               string
	Permission          string
	Operation           string
}

type SupportAccessDecision struct {
	actorID, tenantID                                   uuid.UUID
	tokenVersion, authorizationVersion, authorityClock  int64
	scope, permission, operation                        string
	reporterName, reporterEmail, tenantSlug, tenantName string
}

func NewSupportAccessDecision(
	actorID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorityClock int64,
	scope, permission, operation string,
	reporterName, reporterEmail, tenantSlug, tenantName string,
) (SupportAccessDecision, error) {
	decision := SupportAccessDecision{
		actorID: actorID, tenantID: tenantID, tokenVersion: tokenVersion,
		authorizationVersion: authorizationVersion, authorityClock: authorityClock,
		scope: scope, permission: permission, operation: operation,
		reporterName:  strings.TrimSpace(reporterName),
		reporterEmail: strings.TrimSpace(strings.ToLower(reporterEmail)),
		tenantSlug:    strings.TrimSpace(tenantSlug), tenantName: strings.TrimSpace(tenantName),
	}
	if !decision.valid() {
		return SupportAccessDecision{}, fmt.Errorf("IAM Support access decision is invalid")
	}
	return decision, nil
}

func (decision SupportAccessDecision) valid() bool {
	if decision.actorID == uuid.Nil || decision.tokenVersion < 1 ||
		decision.authorizationVersion < 1 || decision.authorityClock < 1 {
		return false
	}
	if _, valid := resolveSupportAccess(
		decision.scope, decision.permission, decision.operation,
	); !valid {
		return false
	}
	if decision.scope == "platform" {
		return decision.tenantID == uuid.Nil
	}
	return (decision.scope == "reporter" || decision.scope == "tenant") &&
		decision.tenantID != uuid.Nil && decision.tenantSlug != "" && decision.tenantName != ""
}

func (decision SupportAccessDecision) ActorID() uuid.UUID  { return decision.actorID }
func (decision SupportAccessDecision) TenantID() uuid.UUID { return decision.tenantID }
func (decision SupportAccessDecision) TokenVersion() int64 { return decision.tokenVersion }
func (decision SupportAccessDecision) AuthorizationVersion() int64 {
	return decision.authorizationVersion
}
func (decision SupportAccessDecision) AuthorityClock() int64 { return decision.authorityClock }
func (decision SupportAccessDecision) Scope() string         { return decision.scope }
func (decision SupportAccessDecision) Permission() string    { return decision.permission }
func (decision SupportAccessDecision) Operation() string     { return decision.operation }
func (decision SupportAccessDecision) ReporterName() string  { return decision.reporterName }
func (decision SupportAccessDecision) ReporterEmail() string { return decision.reporterEmail }
func (decision SupportAccessDecision) TenantSlug() string    { return decision.tenantSlug }
func (decision SupportAccessDecision) TenantName() string    { return decision.tenantName }
