package application

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
	"github.com/google/uuid"
)

type DeploymentMode string
type TenantPurpose string

const (
	ModeOnPremise DeploymentMode = "onpremise"
	ModeSaaS      DeploymentMode = "saas"

	PurposeOnPremisePrimary TenantPurpose = "onpremise_primary"
	PurposeSaaSCustomer     TenantPurpose = "saas_customer"
	PurposeSaaSControl      TenantPurpose = "saas_control"

	PermissionTenantRead   = "analytics.usage.read"
	PermissionPlatformRead = "platform.analytics.usage.read"
)

type Access struct {
	actorID, subjectTenantID, targetTenantID, environmentID uuid.UUID
	tokenVersion, authorizationVersion, authorityClock      int64
	scope                                                   domain.Scope
	mode                                                    DeploymentMode
	purpose                                                 TenantPurpose
	controlPrimary                                          bool
	permission, operation                                   string
}

type AccessInput struct {
	ActorID, SubjectTenantID, TargetTenantID, EnvironmentID uuid.UUID
	TokenVersion, AuthorizationVersion, AuthorityClock      int64
	Scope                                                   domain.Scope
	Mode                                                    DeploymentMode
	Purpose                                                 TenantPurpose
	ControlPrimary                                          bool
	Permission, Operation                                   string
}

func NewAccess(input AccessInput) (Access, error) {
	access := Access{actorID: input.ActorID, subjectTenantID: input.SubjectTenantID,
		targetTenantID: input.TargetTenantID, environmentID: input.EnvironmentID,
		tokenVersion: input.TokenVersion, authorizationVersion: input.AuthorizationVersion,
		authorityClock: input.AuthorityClock, scope: input.Scope, mode: input.Mode,
		purpose: input.Purpose, controlPrimary: input.ControlPrimary,
		permission: input.Permission, operation: input.Operation}
	if !access.valid() {
		return Access{}, domain.ErrForbidden
	}
	return access, nil
}

func (access Access) valid() bool {
	if access.actorID == uuid.Nil || access.subjectTenantID == uuid.Nil ||
		access.environmentID == uuid.Nil || access.tokenVersion < 1 ||
		access.authorizationVersion < 1 || access.authorityClock < 1 ||
		access.permission == "" || access.operation == "" {
		return false
	}
	if access.scope == domain.TenantScope {
		purposeValid := access.mode == ModeOnPremise && access.purpose == PurposeOnPremisePrimary ||
			access.mode == ModeSaaS && access.purpose == PurposeSaaSCustomer
		return purposeValid && access.targetTenantID == access.subjectTenantID &&
			access.permission == PermissionTenantRead && !access.controlPrimary
	}
	return access.scope == domain.PlatformScope && access.mode == ModeSaaS &&
		access.purpose == PurposeSaaSControl && access.controlPrimary &&
		access.targetTenantID == uuid.Nil && access.permission == PermissionPlatformRead
}

func (access Access) Allows(scope domain.Scope, operation string) bool {
	return access.valid() && access.scope == scope && access.operation == operation
}

func (access Access) ActorID() uuid.UUID          { return access.actorID }
func (access Access) SubjectTenantID() uuid.UUID  { return access.subjectTenantID }
func (access Access) TargetTenantID() uuid.UUID   { return access.targetTenantID }
func (access Access) EnvironmentID() uuid.UUID    { return access.environmentID }
func (access Access) TokenVersion() int64         { return access.tokenVersion }
func (access Access) AuthorizationVersion() int64 { return access.authorizationVersion }
func (access Access) AuthorityClock() int64       { return access.authorityClock }
func (access Access) Scope() domain.Scope         { return access.scope }
func (access Access) Mode() DeploymentMode        { return access.mode }
func (access Access) Purpose() TenantPurpose      { return access.purpose }
func (access Access) ControlPrimary() bool        { return access.controlPrimary }
func (access Access) Permission() string          { return access.permission }
func (access Access) Operation() string           { return access.operation }
