package environmentoperator

import (
	"context"
	"io"
	"time"

	installationv1 "github.com/Onurlulardan/harpy/contracts/installation/v1"
	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

type DeploymentPolicy interface {
	CreateDeploymentPolicy(context.Context, tenantapi.InitializeDeploymentPolicyRequest) (
		tenantapi.DeploymentPolicyInitialization, error,
	)
}

type Onboarding interface {
	Accept(context.Context, tenantapi.OnboardingPlan) (tenantapi.OnboardingOperation, error)
	Claim(context.Context, tenantapi.ClaimAttemptRequest) (tenantapi.AttemptLease, error)
	Record(context.Context, tenantapi.OnboardingReceipt, int64) (
		tenantapi.OnboardingTransition, error,
	)
	ReceiveAuthorityAck(context.Context, tenantapi.AuthorityAcknowledgement, int64) (
		tenantapi.OnboardingTransition, error,
	)
}

type TenantMigration interface {
	CurrentReady(context.Context) (tenantapi.TenantMigrationReady, bool, error)
	Enqueue(context.Context, uuid.UUID) (tenantapi.TenantMigrationJob, error)
	Get(context.Context, uuid.UUID) (tenantapi.TenantMigrationJob, bool, error)
}

type Offering interface {
	Resolve(context.Context, OfferingRequest) (tenantapi.BaseOfferingPin, error)
}

type ProtectedAuthority interface {
	Prepare(context.Context, AuthorityRequest) (tenantapi.ProtectedAuthorityPlan, error)
	Acknowledge(context.Context, AuthorityRequest) (tenantapi.AuthorityAcknowledgement, error)
}

type Finalizer interface {
	Finalize(context.Context, FinalizationRequest) error
}

type Dependencies struct {
	Policy       DeploymentPolicy
	Onboarding   Onboarding
	Migration    TenantMigration
	Offering     Offering
	Authority    ProtectedAuthority
	Finalizer    Finalizer
	Entropy      io.Reader
	Sleep        func(context.Context, time.Duration) error
	LeaseOwner   string
	PollInterval time.Duration
}

type OfferingRequest struct {
	Manifest    installationv1.Manifest
	OperationID uuid.UUID
	CreatedAt   time.Time
}
