package onboarding

import (
	"context"

	activation "github.com/Onurlulardan/harpy/core/internal/capabilities/tenantactivation/api"
	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

type OperationStore interface {
	Accept(context.Context, tenantapi.OnboardingPlan) (tenantapi.OnboardingOperation, error)
	Claim(context.Context, tenantapi.ClaimAttemptRequest) (tenantapi.AttemptLease, error)
	RecordReceipt(context.Context, tenantapi.OnboardingReceipt, int64) (
		tenantapi.OnboardingTransition, error,
	)
	ReceiveAuthorityAck(context.Context, tenantapi.AuthorityAcknowledgement, int64) (
		tenantapi.OnboardingTransition, error,
	)
	Reconcile(context.Context, ReconcileRequest) (tenantapi.OnboardingTransition, error)
}

type ActivationUnitOfWork interface {
	Run(context.Context, func(context.Context, ActivationTransaction) error) error
}

type ActivationTransaction interface {
	Activate(
		context.Context,
		ActivationRequest,
		ActivationCompletion,
		ActivationTransactionCallback,
	) error
}

type ActivationTransactionCallback func(
	tenantapi.ClaimedActivation,
	ActivationResources,
) error

type ActivationResources struct {
	BindTenantSchema func(context.Context, string) error
	BindOwner        func(ActivationOwner) (ActivationOwnerExecutor, error)
}

type ActivationOwner string

const (
	ActivationOwnerTenancy           ActivationOwner = "tenancy"
	ActivationOwnerIAM               ActivationOwner = "iam"
	ActivationOwnerBilling           ActivationOwner = "billing"
	ActivationOwnerSystemResources   ActivationOwner = "system_resource_layout"
	ActivationOwnerPlatformOps       ActivationOwner = "platform_ops"
	ActivationOwnerPlatformAuthority ActivationOwner = "platform_authority"
	ActivationOwnerAsync             ActivationOwner = "platform_async"
)

type ActivationOwnerRow interface{ Scan(...any) error }

type ActivationOwnerExecutor struct {
	QueryRow func(context.Context, string, ...any) ActivationOwnerRow
	Execute  func(context.Context, string, ...any) error
}

type ActivationCallback func(
	context.Context,
	activation.Scope,
	ActivationResources,
) error

type ActivationRequest struct {
	Plan             tenantapi.OnboardingPlan
	OperationID      uuid.UUID
	LeaseTokenSHA256 string
	ExpectedRevision int64
}

type ActivationCompletion struct {
	OperationID      uuid.UUID
	ReceiptID        uuid.UUID
	ReceiptSHA256    string
	EvidenceSHA256   string
	ExpectedRevision int64
}

type ReconcileRequest struct {
	OperationID      uuid.UUID
	ReceiptID        uuid.UUID
	ReceiptSHA256    string
	EvidenceSHA256   string
	ExpectedRevision int64
}
