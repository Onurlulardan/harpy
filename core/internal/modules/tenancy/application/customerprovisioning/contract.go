package customerprovisioning

import (
	"context"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

type DeploymentPolicy interface {
	Snapshot() (tenantapi.DeploymentPolicy, error)
	PrimaryTarget() (tenantapi.PrimaryTenantTarget, error)
}

type OfferingResolver interface {
	ResolveSaaSFreeBase(context.Context) (tenantapi.BaseOfferingPin, error)
}

type AuthorityRequest struct {
	OperationID, EnvironmentID, TenantID uuid.UUID
	FirstUserID, ProtectedAdminBindingID uuid.UUID
	IssuedAt                             time.Time
}

type AuthorityPlanner interface {
	PrepareInitialTenant(context.Context, AuthorityRequest) (
		tenantapi.ProtectedAuthorityPlan, error,
	)
}

type SetupRequest struct {
	OperationID, TenantID, UserID uuid.UUID
	Email                         string
	CreatedAt                     time.Time
}

type SetupReissueRequest struct {
	OperationID, ReissueOperationID, TenantID, UserID uuid.UUID
	SchemaName, Email                                 string
	CreatedAt                                         time.Time
}

type SetupReissueTarget struct {
	Operation          Operation
	ReissueOperationID uuid.UUID
	SchemaName         string
	UserID             uuid.UUID
	Email              string
}

type SetupPreparer interface {
	PrepareInitialAdmin(context.Context, SetupRequest) error
	ReissueInitialAdmin(context.Context, SetupReissueRequest) error
}

type AcceptedRequest struct {
	RequestedBySubjectID uuid.UUID
	ControlTenantID      uuid.UUID
	JobID                uuid.UUID
	Customer             CustomerRequest
}

type OperationStore interface {
	Create(context.Context, tenantapi.OnboardingPlan, AcceptedRequest) (Operation, error)
	Read(context.Context, uuid.UUID, uuid.UUID) (Operation, error)
	LoadSetupReissueTarget(context.Context, uuid.UUID) (SetupReissueTarget, error)
}

type Clock interface{ Now() time.Time }
